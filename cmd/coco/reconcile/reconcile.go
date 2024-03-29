package reconcile

import (
	"context"
	"fmt"

	"net/http"
	"os"

	"github.com/SAP/configuration-tools-for-gitops/v2/pkg/github"
	"github.com/SAP/configuration-tools-for-gitops/v2/pkg/terminal"
	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/config"
	"github.com/go-git/go-git/v5/plumbing"
	githttp "github.com/go-git/go-git/v5/plumbing/transport/http"
	gogithub "github.com/google/go-github/v51/github"
)

type BranchConfig struct {
	Name   string
	Remote string
}
type Client struct {
	client              github.Interface
	target              BranchConfig
	source              BranchConfig
	reconcileBranchName string
	owner               string
	repo                string
	logger              Logger
}

type Logger interface {
	Info(args ...interface{})
	Debugf(template string, args ...interface{})
	Infof(template string, args ...interface{})
}

const (
	notUsed    = "notUsed"
	allAllowed = 0777
)

func New(
	ctx context.Context, owner, repo, token, githubBaseURL string,
	targetBranch, sourceBranch BranchConfig,
	logger Logger,
) (*Client, error) {
	reconcileBranchName := fmt.Sprintf("reconcile/%s-%s", sourceBranch.Name, targetBranch.Name)

	// Authenticate with Github
	isEnterprise := false
	if githubBaseURL != "https://github.com" &&
		githubBaseURL != "https://www.github.com" &&
		githubBaseURL != "" {
		isEnterprise = true
	}
	// target is base and source is head

	if targetBranch.Remote != sourceBranch.Remote {
		if err := differentRemotes(targetBranch, sourceBranch, token, logger); err != nil {
			return nil, err
		}
		reconcileBranchName = fmt.Sprintf("reconcile/remote-replica/%s-%s", sourceBranch.Name, targetBranch.Name)
	}

	client, err := githubClient(ctx, token, owner, repo, githubBaseURL, isEnterprise)
	if err != nil {
		return nil, fmt.Errorf("failed to authenticate with Github: %w", err)
	}
	return &Client{
		client:              client,
		target:              targetBranch,
		source:              sourceBranch,
		reconcileBranchName: reconcileBranchName,
		owner:               owner,
		repo:                repo,
		logger:              logger,
	}, nil
}

func (r *Client) Reconcile(force bool) error {
	if r.target.Remote != r.source.Remote {
		// change source.name to remote-replica/source.Name and continue
		r.source.Name = fmt.Sprintf("remote-replica/%s", r.source.Name)
	}
	return r.merge(force)
}

func differentRemotes(targetBranch, sourceBranch BranchConfig, token string, logger Logger) error {
	// 		1. clone target repo
	// 		2. add source repo as remote
	// 		3. git fetch [remote-name]
	// 		4. git branch remote-replica/[sourceBranch] [remote-name]/[sourceBranch]
	// 		5. git push origin remote-replica/[sourceBranch]
	logger.Debugf("target and source have different remotes")
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return err
	}
	// avoided utilizing a temporary directory due to the high frequency of deletion,
	// necessitating numerous repo cloning operations.
	targetPath := fmt.Sprintf("%s/.coco-cache/reconcile/target/%s", homeDir, targetBranch.Name)
	err = os.MkdirAll(targetPath, allAllowed)
	if err != nil {
		return err
	}

	targetRepo, err := gitClone(targetPath, false, &git.CloneOptions{
		URL:             targetBranch.Remote,
		Auth:            &githttp.BasicAuth{Username: notUsed, Password: token},
		RemoteName:      "origin",
		ReferenceName:   plumbing.ReferenceName(fmt.Sprintf("refs/heads/%s", targetBranch.Name)),
		Tags:            0,
		InsecureSkipTLS: false,
		CABundle:        []byte{},
		Progress:        os.Stdout,
	})
	if err == git.ErrRepositoryAlreadyExists {
		logger.Debugf("Target repository already exists")
	} else if err != nil {
		return err
	}

	if targetRepo == nil {
		targetRepo, err = gitOpen(targetPath)
		if err != nil {
			return err
		}
	}

	worktree, err := targetRepo.Worktree()
	if err != nil {
		return err
	}
	err = gitPull(worktree, &git.PullOptions{
		Auth:          &githttp.BasicAuth{Username: notUsed, Password: token},
		RemoteName:    "origin",
		ReferenceName: plumbing.ReferenceName(fmt.Sprintf("refs/heads/%s", targetBranch.Name)),
		Progress:      os.Stdout,
	})
	if err == git.NoErrAlreadyUpToDate {
		logger.Debugf("Target branch is already up to date")
	} else if err != nil {
		return fmt.Errorf("target pull failed: %w", err)
	}

	// If the remotes are different, add the remote repository of the 'source' branch
	remoteName := fmt.Sprintf("reconcile/source/%s", sourceBranch.Name)

	err = gitCreateRemote(targetRepo, sourceBranch.Remote, remoteName)
	if err != nil {
		return err
	}

	// Fetch the source branch from the added remote
	err = gitFetch(targetRepo, &git.FetchOptions{
		Auth:       &githttp.BasicAuth{Username: notUsed, Password: token},
		RemoteName: remoteName,
	})
	if err != nil && err != git.NoErrAlreadyUpToDate {
		return err
	}

	// Replicate the branch from the source repository into the target repository
	return copyBranch(targetRepo, sourceBranch, remoteName, token)
}

func copyBranch(targetRepo *git.Repository, sourceBranch BranchConfig, remoteName, token string) error {
	replicaBranchName := plumbing.NewBranchReferenceName(fmt.Sprintf("remote-replica/%s", sourceBranch.Name))
	var sourceRef *plumbing.Reference
	refs, err := targetRepo.References()
	if err != nil {
		return err
	}

	err = refs.ForEach(func(ref *plumbing.Reference) error {
		if ref.Name() == plumbing.ReferenceName(fmt.Sprintf("refs/remotes/%s/%s", remoteName, sourceBranch.Name)) {
			sourceRef = ref
		}
		return nil
	})
	if err != nil {
		return err
	}
	ref := plumbing.NewHashReference(replicaBranchName, sourceRef.Hash())
	err = targetRepo.Storer.SetReference(ref)
	if err != nil {
		return err
	}
	err = gitPush(targetRepo, &git.PushOptions{
		RemoteName: "origin",
		Auth:       &githttp.BasicAuth{Username: notUsed, Password: token},
		Progress:   os.Stdout,
		Force:      true,
	})
	if err == git.NoErrAlreadyUpToDate {
		return nil
	}
	return err
}

func (r *Client) merge(force bool) error {
	// perform merge source -> target
	success, err := r.client.MergeBranches(r.target.Name, r.source.Name)
	if err == nil {
		if !success {
			r.logger.Infof("merge conflict encountered")
			return r.handleMergeConflict(force)
		}
		r.logger.Info("Merged successfully")
		var status int
		_, status, err = r.client.GetBranch(r.reconcileBranchName)

		if status == http.StatusOK {
			err = r.client.DeleteBranch(r.reconcileBranchName, true)
		}

		if status == http.StatusNotFound {
			return nil
		}
		return err
	}

	return fmt.Errorf("failed to merge branches: %w", err)
}

func (r *Client) handleMergeConflict(force bool) error {
	reconcileBranch, status, err := r.client.GetBranch(r.reconcileBranchName)

	if status == http.StatusNotFound {
		return r.handleNewReconcileBranch()
	}

	if status == http.StatusOK {
		var resolved bool
		resolved, err = r.handleExistingReconcileBranch(reconcileBranch, force)
		if err != nil {
			return err
		}
		if resolved {
			return nil
		}
		return r.handleNewReconcileBranch()
	}

	return err
}

func (r *Client) handleExistingReconcileBranch(
	reconcileBranch *gogithub.Branch, force bool,
) (bool, error) {
	// Compare the latest target branch and reconcile branch
	target, status, err := r.client.GetBranch(r.target.Name)
	if err != nil || status != http.StatusOK {
		return false, fmt.Errorf("failed to get target branch: %w", err)
	}
	commits, err := r.client.CompareCommits(
		reconcileBranch,
		target)
	if err != nil {
		return false, fmt.Errorf("failed to compare commits: %w", err)
	}
	if commits.GetAheadBy() > 0 {
		return r.handleTargetAhead(force)
	}
	// check mergability
	return r.checkMergeability()
}

func (r *Client) handleNewReconcileBranch() error {
	// Create a new branch reconcile branch from target branch
	target, err := r.client.GetBranchRef(r.target.Name)
	if err != nil {
		return fmt.Errorf("failed to get target branch reference: %w", err)
	}
	if err = r.client.CreateBranch(r.reconcileBranchName, target); err != nil {
		return fmt.Errorf("failed to create reconcile branch: %w", err)
	}
	r.logger.Debugf("Created new reconcile branch from %s", r.target)

	pr, err := r.client.CreatePullRequest(r.source.Name, r.reconcileBranchName)
	if err != nil {
		return fmt.Errorf("failed to create a draft PR: %w", err)
	}
	r.logger.Infof("Draft pull request #%d created: %s\n", pr.GetNumber(), pr.GetHTMLURL())
	return nil
}

func (r *Client) checkMergeability() (bool, error) {
	prs, err := r.client.ListPullRequests()
	if err != nil {
		return false, err
	}
	var pr *gogithub.PullRequest
	for _, p := range prs {
		if p.Head.GetRef() == r.source.Name && p.Base.GetRef() == r.reconcileBranchName {
			pr = p
			break
		}
	}
	if pr == nil {
		r.logger.Infof("the pull request was not found")

		pr, err = r.client.CreatePullRequest(r.source.Name, r.reconcileBranchName)
		if err != nil {
			return false, fmt.Errorf("failed to create a draft PR: %w", err)
		}
		r.logger.Infof("Draft pull request #%d created: %s", pr.GetNumber(), pr.GetHTMLURL())
		return true, nil
	}
	// check if the pull request is mergable
	if !pr.GetMergeable() {
		return false, fmt.Errorf("please re-try after resolving the merge conflicts here: %s", pr.GetURL())
	}

	return true, nil
}

func (r *Client) handleTargetAhead(force bool) (bool, error) {
	if force {
		return false, r.client.DeleteBranch(r.reconcileBranchName, force)
	}

	printTerminal(fmt.Sprintf(
		"The target branch has new commits. "+
			"Do you want to delete the reconcile branch and rerun the command? %+v",
		terminal.AffirmationOptions,
	))
	deleteBranch, err := confirmed()
	if err != nil {
		return false, fmt.Errorf("input must be in %+v: %w", terminal.AffirmationOptions, err)
	}

	if deleteBranch {
		return false, r.client.DeleteBranch(r.reconcileBranchName, true)
	}
	printTerminal(fmt.Sprintf(
		"Please merge the branch %q into the branch %q and rerun the command"+
			"`coco reconcile --source %s --target %s --owner %s --repo %s`",
		r.target, r.reconcileBranchName, r.source, r.target, r.owner, r.repo,
	))
	return true, nil
}

var (
	githubClient  = github.New
	printTerminal = terminal.Output
	confirmed     = terminal.IsYes
	gitClone      = git.PlainClone
	gitOpen       = git.PlainOpen
)

var gitPull = func(worktree *git.Worktree, o *git.PullOptions) error {
	return worktree.Pull(o)
}

var gitFetch = func(repo *git.Repository, o *git.FetchOptions) error {
	return repo.Fetch(o)
}

var gitPush = func(repo *git.Repository, o *git.PushOptions) error {
	return repo.Push(o)
}

var gitCreateRemote = func(repo *git.Repository, remoteURL, remoteName string) error {
	remotes, err := repo.Remotes()
	if err != nil {
		return err
	}
	remoteExists := false

	for _, remote := range remotes {
		if remoteName == remote.Config().Name {
			remoteExists = true
		}
	}

	if !remoteExists {
		_, err = repo.CreateRemote(&config.RemoteConfig{
			Name: remoteName,
			URLs: []string{remoteURL},
		})
		if err != nil {
			return err
		}
	}

	return nil
}
