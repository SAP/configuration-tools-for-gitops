package githubclient

import (
	"context"
	"fmt"
	"strings"

	"github.com/configuration-tools-for-gitops/pkg/log"
	"github.com/google/go-github/v51/github"
	"golang.org/x/oauth2"
)

type Mock struct {
	client                *github.Client
	ctx                   context.Context
	owner                 string
	repo                  string
	base                  string
	head                  string
	reconcileBranchName   string
	reconcileBranchExists bool
	targetAhead           bool
	mergeSuccessful       bool
	reconcileMergable     bool
}

func NewMock(
	token,
	owner,
	repo,
	base,
	head,
	reconcileBranchName string,
	ctx context.Context,
	reconcileBranchExists,
	targetAhead,
	mergeSuccessful,
	reconcileMergable bool) (*Github, error) {
	//Authenticate with Github
	ts := oauth2.StaticTokenSource(
		&oauth2.Token{AccessToken: token},
	)
	tc := oauth2.NewClient(ctx, ts)
	client := github.NewClient(tc)
	return &Github{
		client,
		ctx,
		owner,
		repo,
		base,
		head,
		reconcileBranchName,
	}, nil
}

func (gh *Mock) Merge(dryRun bool) error {
	err := gh.mergeBranches()
	if err != nil {
		if strings.Contains(err.Error(), "Merge conflict") {
			return gh.handleMergeConflict(dryRun)
		} else {
			return fmt.Errorf("failed to merge branches: %v", err)
		}
	} else {
		if dryRun {
			log.Sugar.Debug("No merge conflicts found")
			return nil
		}
		log.Sugar.Info("Merged successfully")
	}

	return nil
}

func (gh *Mock) handleMergeConflict(dryRun bool) error {
	if dryRun {
		return fmt.Errorf("merge conflicts detected")
	}

	reconcileBranch, err := gh.getBranch(gh.reconcileBranchName)

	// get a list of branches
	// branches, err := getBranchList(client, ctx, owner, repo)

	if err == nil {
		var resolved bool
		resolved, err = gh.handleExistingReconcileBranch(reconcileBranch)
		if err != nil {
			return err
		}
		if resolved {
			return nil
		}
	}

	return gh.handleNewReconcileBranch()
}

func (gh *Mock) handleExistingReconcileBranch(reconcileBranch *github.Branch) (bool, error) {
	// Compare the latest target branch and reconcile/target branch
	target, err := gh.getBranch(gh.base)
	if err != nil {
		return false, fmt.Errorf("failed to get target branch: %v", err)
	}
	commits, err := gh.compareCommits(
		reconcileBranch,
		target)
	if err != nil {
		return false, fmt.Errorf("failed to compare commits: %v", err)
	}
	if commits.GetAheadBy() > 0 {
		return gh.handleTargetAhead()
	} else {
		//check mergability
		return gh.checkMergeability()
		// return true, fmt.Errorf("%s already exists for the latest target branch", reconcileBranchName)
	}
}

func (gh *Mock) handleNewReconcileBranch() error {
	// Create a new branch reconcile/target branch from target branch
	target, err := gh.getBranchRef(gh.base)
	if err != nil {
		return fmt.Errorf("Failed to get target branch reference: %v", err)
	}
	if err = gh.createBranch(gh.reconcileBranchName, target); err != nil {
		return fmt.Errorf("Failed to create reconcile branch: %v", err)
	}
	log.Sugar.Debugf("Created new reconcile branch from %s", gh.base)

	pr, err := gh.createPullRequest(gh.head, gh.reconcileBranchName)
	if err != nil {
		return fmt.Errorf("failed to create a draft PR: %v", err)
	}
	log.Sugar.Info("Draft pull request #%d created: %s\n", pr.GetNumber(), pr.GetHTMLURL())
	return nil
}

func (gh *Mock) checkMergeability() (bool, error) {
	return true, nil
}

func (gh *Mock) handleTargetAhead() (bool, error) {
	return true, nil
}

func (gh *Mock) mergeBranches() error {
	if gh.mergeSuccessful {
		return nil
	} else {
		return fmt.Errorf("Merge conflict")
	}
}

func (gh *Mock) getBranch(branchName string) (*github.Branch, error) {
	dummySHA := "dd0b557d0696d2e1b8a1cf9de6b3c6d3a3a8a8f9"
	return &github.Branch{
		Name: &gh.base,
		Commit: &github.RepositoryCommit{
			SHA: &dummySHA,
		},
	}, nil
}

func (gh *Mock) compareCommits(branch1 *github.Branch, branch2 *github.Branch) (*github.CommitsComparison, error) {
	var aheadBy int
	if gh.targetAhead {
		aheadBy = 2
	} else {
		aheadBy = 0
	}
	return &github.CommitsComparison{
		//head is ahead of base
		AheadBy: &aheadBy,
	}, nil
}

func (gh *Mock) deleteBranch(branchName string) error {
	return nil
}

func (gh *Mock) getBranchRef(branchName string) (*github.Reference, error) {
	return &github.Reference{}, nil
}

func (gh *Mock) createBranch(branchName string, target *github.Reference) error {
	return nil
}

func (gh *Mock) createPullRequest(head, base string) (*github.PullRequest, error) {
	prNumber := 1
	url := fmt.Sprintf("www.github.com/%s/%s/pulls/%v", gh.owner, gh.repo, prNumber)
	return &github.PullRequest{
		Number:  &prNumber,
		HTMLURL: &url,
	}, nil
}
