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

type ReconcileClient struct {
	client              github.Interface
	target              string
	source              string
	reconcileBranchName string
	owner               string
	repo                string
}

func New(sourceBranch, targetBranch, owner, repo, token, githubBaseURL string, isEnterprise bool,
	ctx context.Context) (*ReconcileClient, error) {
	reconcileBranchName := fmt.Sprintf("reconcile/%s-%s", sourceBranch, targetBranch)

	// Authenticate with Github
	// target is base and source is head
	client, err := githubClient(token, owner, repo, githubBaseURL, ctx, isEnterprise)
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

func (r *ReconcileClient) Reconcile(force bool) error {
	return r.merge(force)
}

func (r *ReconcileClient) merge(force bool) error {
	success, err := r.client.MergeBranches(r.target, r.source)
	if err == nil {
		if !success {
			return r.handleMergeConflict(force)
		}
		log.Sugar.Info("Merged successfully")
		return nil
	}

	return fmt.Errorf("failed to merge branches: %w", err)
}

func (r *ReconcileClient) handleMergeConflict(force bool) error {
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
		} else {
			return r.handleNewReconcileBranch()
		}
	}

	return err
}

func (r *ReconcileClient) handleExistingReconcileBranch(
	reconcileBranch *gogithub.Branch, force bool,
) (bool, error) {
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
		return r.handleTargetAhead(force)
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

func (r *ReconcileClient) handleTargetAhead(force bool) (bool, error) {
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
)
