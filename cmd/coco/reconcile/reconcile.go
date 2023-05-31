package reconcile

import (
	"context"
	"fmt"
	"net/http"

	"github.com/SAP/configuration-tools-for-gitops/pkg/github"
	"github.com/SAP/configuration-tools-for-gitops/pkg/log"
	"github.com/SAP/configuration-tools-for-gitops/pkg/terminal"
	gogithub "github.com/google/go-github/v51/github"
)

type githubClient interface {
	CompareCommits(branch1 *gogithub.Branch, branch2 *gogithub.Branch) (*gogithub.CommitsComparison, error)
	CreateBranch(branchName string, target *gogithub.Reference) error
	CreatePullRequest(head, base string) (*gogithub.PullRequest, error)
	DeleteBranch(branchName string) error
	GetBranch(branchName string) (*gogithub.Branch, int, error)
	GetBranchRef(branchName string) (*gogithub.Reference, error)
	ListPullRequests() ([]*gogithub.PullRequest, error)
	MergeBranches(base, head string) (bool, error)
}
type ReconcileClient struct {
	client              githubClient
	target              string
	source              string
	reconcileBranchName string
	owner               string
	repo                string
}

func New(sourceBranch, targetBranch, owner, repo, token string, ctx context.Context) (*ReconcileClient, error) {
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

func (r *ReconcileClient) Reconcile() error {
	return r.merge()
}

func (r *ReconcileClient) merge() error {
	success, err := r.client.MergeBranches(r.target, r.source)
	if err == nil {
		if !success {
			return r.handleMergeConflict()
		}
		log.Sugar.Info("Merged successfully")
		return nil
	}

	return fmt.Errorf("failed to merge branches: %w", err)
}

func (r *ReconcileClient) handleMergeConflict() error {

	reconcileBranch, status, err := r.client.GetBranch(r.reconcileBranchName)

	if status == http.StatusNotFound {
		return r.handleNewReconcileBranch()
	}

	if status == http.StatusOK {
		var resolved bool
		resolved, err = r.handleExistingReconcileBranch(reconcileBranch)
		if err != nil {
			return err
		}
		if resolved {
			return nil
		} else {
			return r.handleNewReconcileBranch()
		}
	}

	return err
}

func (r *ReconcileClient) handleExistingReconcileBranch(reconcileBranch *gogithub.Branch) (bool, error) {
	// Compare the latest target branch and reconcile branch
	target, status, err := r.client.GetBranch(r.target)
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
		return r.handleTargetAhead()
	}
	// check mergability
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
	printTerminal(fmt.Sprintf(
		"The target branch has new commits, choose one of the following options:\n\n"+
			"Option 1: Merge the target branch into the reconcile branch manually and rerun command `coco reconcile`\n\n"+
			"Option 2: Automatically delete the reconcile branch and rerun the command "+
			"`coco reconcile --source %s --target %s --owner %s --repo %s`\n\n"+
			"Enter [1] for Option 1 or [2] for Option 2: ",
		r.source, r.target, r.owner, r.repo,
	))
	input, err := readTerminal()
	if err != nil {
		return false, fmt.Errorf("illegal input %v - allowed options are: [1, 2]", input)
	}

	switch input {
	case 1:
		printTerminal(fmt.Sprintf(
			"Please merge the branch `%q` into the branch `%q` and rerun the `coco reconcile` command",
			r.target, r.reconcileBranchName,
		))
	case 2:
		return false, r.client.DeleteBranch(r.reconcileBranchName)
	default:
		return false, fmt.Errorf("illegal input %v - allowed options are: [1, 2]", input)
	}
	return true, nil
}

var (
	newGithubClient = func(token, owner, repo string, ctx context.Context) (githubClient, error) {
		return github.New(token, owner, repo, ctx)
	}
	printTerminal = terminal.Output
	readTerminal  = terminal.ReadInt
)
