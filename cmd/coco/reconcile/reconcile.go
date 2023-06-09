package reconcile

import (
	"context"
	"fmt"
	"net/http"

	"github.com/SAP/configuration-tools-for-gitops/pkg/github"
	"github.com/SAP/configuration-tools-for-gitops/pkg/terminal"
	gogithub "github.com/google/go-github/v51/github"
)

type Client struct {
	client              github.Interface
	target              string
	source              string
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

func New(
	ctx context.Context,
	sourceBranch, targetBranch, owner, repo, token, githubBaseURL string,
	logger Logger,
) (*Client, error) {
	reconcileBranchName := fmt.Sprintf("reconcile/%s-%s", sourceBranch, targetBranch)

	// Authenticate with Github
	isEnterprise := false
	if githubBaseURL != "https://github.com" &&
		githubBaseURL != "https://www.github.com" &&
		githubBaseURL != "" {
		isEnterprise = true
	}
	// target is base and source is head
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
	return r.merge(force)
}

func (r *Client) merge(force bool) error {
	success, err := r.client.MergeBranches(r.target, r.source)
	if err == nil {
		if !success {
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

func (r *Client) handleNewReconcileBranch() error {
	// Create a new branch reconcile branch from target branch
	target, err := r.client.GetBranchRef(r.target)
	if err != nil {
		return fmt.Errorf("failed to get target branch reference: %w", err)
	}
	if err = r.client.CreateBranch(r.reconcileBranchName, target); err != nil {
		return fmt.Errorf("failed to create reconcile branch: %w", err)
	}
	r.logger.Debugf("Created new reconcile branch from %s", r.target)

	pr, err := r.client.CreatePullRequest(r.source, r.reconcileBranchName)
	if err != nil {
		return fmt.Errorf("failed to create a draft PR: %w", err)
	}
	r.logger.Info("Draft pull request #%d created: %s\n", pr.GetNumber(), pr.GetHTMLURL())
	return nil
}

func (r *Client) checkMergeability() (bool, error) {
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
		r.logger.Infof("the pull request was not found")

		pr, err = r.client.CreatePullRequest(r.source, r.reconcileBranchName)
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
)
