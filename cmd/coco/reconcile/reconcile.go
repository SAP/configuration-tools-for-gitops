package reconcile

import (
	"context"
	"fmt"
	"time"

	"github.com/SAP/configuration-tools-for-gitops/pkg/github"
	"github.com/SAP/configuration-tools-for-gitops/pkg/log"
	"github.com/SAP/configuration-tools-for-gitops/pkg/terminal"
	gogithub "github.com/google/go-github/v51/github"
)

var (
	timeout = 100 * time.Millisecond
)

type githubClient interface {
	CompareCommits(branch1 *gogithub.Branch, branch2 *gogithub.Branch) (*gogithub.CommitsComparison, error)
	CreateBranch(branchName string, target *gogithub.Reference) error
	CreatePullRequest(head string, base string) (*gogithub.PullRequest, error)
	DeleteBranch(branchName string) error
	GetBranch(branchName string) (*gogithub.Branch, error)
	GetBranchRef(branchName string) (*gogithub.Reference, error)
	ListPullRequests() ([]*gogithub.PullRequest, error)
	MergeBranches(base string, head string) (bool, error)
}
type ReconcileClient struct {
	client              githubClient
	target              string
	source              string
	reconcileBranchName string
	owner               string
	repo                string
}

func New(sourceBranch, targetBranch, owner, repo, token string) (*ReconcileClient, error) {
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	reconcileBranchName := fmt.Sprintf("reconcile/%s-%s", sourceBranch, targetBranch)

	// Authenticate with Github
	// target is base and source is head
	client, err := newGithubClient(token, owner, repo, ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to authenticate with Github: %w", err)
	}
	return &ReconcileClient{
		client:              client,
		target:              targetBranch,
		source:              sourceBranch,
		reconcileBranchName: reconcileBranchName,
		owner:               owner,
		repo:                repo,
	}, nil
}

func (r *ReconcileClient) Reconcile(dryRun bool) error {
	return r.merge(dryRun)
}

func (r *ReconcileClient) merge(dryRun bool) error {
	success, err := r.client.MergeBranches(r.target, r.source)
	if err == nil {
		if !success {
			return r.handleMergeConflict(dryRun)
		}
		if dryRun {
			log.Sugar.Debug("No merge conflicts found (dry-run mode)")
			return nil
		}
		log.Sugar.Info("Merged successfully")
		return nil
	}

	return fmt.Errorf("failed to merge branches: %w", err)
}

func (r *ReconcileClient) handleMergeConflict(dryRun bool) error {
	if dryRun {
		return fmt.Errorf("merge conflicts detected")
	}

	reconcileBranch, err := r.client.GetBranch(r.reconcileBranchName)

	if err != nil {
		var resolved bool
		resolved, err = r.handleExistingReconcileBranch(reconcileBranch)
		if err != nil {
			return err
		}
		if resolved {
			return nil
		}
	}

	return r.handleNewReconcileBranch()
}

func (r *ReconcileClient) handleExistingReconcileBranch(reconcileBranch *gogithub.Branch) (bool, error) {
	// Compare the latest target branch and reconcile branch
	target, err := r.client.GetBranch(r.target)
	if err != nil {
		return false, fmt.Errorf("failed to get target branch: %w", err)
	}
	commits, err := r.client.CompareCommits(
		reconcileBranch,
		target)
	if err != nil {
		return false, fmt.Errorf("failed to compare commits: %w", err)
	}
	if commits.GetAheadBy() > 0 {
		return r.handleTargetAhead()
	}
	//check mergability
	return r.checkMergeability()
}

func (r *ReconcileClient) handleNewReconcileBranch() error {
	// Create a new branch reconcile branch from target branch
	target, err := r.client.GetBranchRef(r.target)
	if err != nil {
		return fmt.Errorf("failed to get target branch reference: %w", err)
	}
	if err = r.client.CreateBranch(r.reconcileBranchName, target); err != nil {
		return fmt.Errorf("failed to create reconcile branch: %w", err)
	}
	log.Sugar.Debugf("Created new reconcile branch from %s", r.target)

	pr, err := r.client.CreatePullRequest(r.source, r.reconcileBranchName)
	if err != nil {
		return fmt.Errorf("failed to create a draft PR: %w", err)
	}
	log.Sugar.Info("Draft pull request #%d created: %s\n", pr.GetNumber(), pr.GetHTMLURL())
	return nil
}

func (r *ReconcileClient) checkMergeability() (bool, error) {
	prs, err := r.client.ListPullRequests()
	if err != nil {
		return false, err
	}
	var pr *gogithub.PullRequest
	for _, p := range prs {
		if p.Head.GetRef() == r.source && p.Base.GetRef() == r.reconcileBranchName {
			pr = p
			break
		}
	}
	if pr == nil {
		return false, fmt.Errorf("the pull request was not found")
	}
	// check if the pull request is mergable
	if !pr.GetMergeable() {
		return false, fmt.Errorf("please re-try after resolving the merge conflicts here: %s", pr.GetURL())
	}
	// perform the merge
	_, err = r.client.MergeBranches(r.target, r.reconcileBranchName)
	if err == nil {
		log.Sugar.Infof("Successfully merged %s to %s", r.reconcileBranchName, r.target)
		return true, err
	}
	return false, err
}

func (r *ReconcileClient) handleTargetAhead() (bool, error) {
	print(fmt.Sprint(
		"The target branch has new commits, choose one of the following options:\n\n" +
			"Option 1: Merge the target branch into the reconcile branch manually and rerun command `coco reconcile`\n\n" +
			"Option 2: Automatically delete the reconcile branch and rerun the command " +
			fmt.Sprintf("`coco reconcile --source %s --target %s --owner %s --repo %s`",
				r.source, r.target, r.owner, r.repo) + "\n\n" +
			"Enter [1] for Option 1 or [2] for Option 2: ",
	))
	rawInput := read()
	input, ok := rawInput.(int)
	if !ok {
		return false, fmt.Errorf("illegal input %q - allowed options are: [1, 2]", input)
	}

	switch input {
	case 1:
		fmt.Printf("\nPlease merge the branch `%q` into the branch `%q` and rerun the `coco reconcile` command", r.target, r.reconcileBranchName)
	case 2:
		return false, r.client.DeleteBranch(r.reconcileBranchName)
	default:
		return false, fmt.Errorf("illegal input %q - allowed options are: [1, 2]", input)
	}
	return true, nil
}

var (
	newGithubClient = func(token, owner, repo string, ctx context.Context) (githubClient, error) {
		return github.New(token, owner, repo, ctx)
	}
	print = terminal.Output
	read  = terminal.Read
)
