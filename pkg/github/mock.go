package github

import (
	"context"
	"fmt"

	"github.com/google/go-github/v51/github"
	"golang.org/x/oauth2"
)

type Mock struct {
	client                *github.Client
	ctx                   context.Context
	owner                 string
	repo                  string
	reconcileBranchExists bool
	targetAhead           bool
	mergeSuccessful       bool
}

func NewMock(
	token,
	owner,
	repo string,
	ctx context.Context,
	reconcileBranchExists,
	targetAhead,
	mergeSuccessful bool) (*Mock, error) {
	// Authenticate with Github
	ts := oauth2.StaticTokenSource(
		&oauth2.Token{AccessToken: token},
	)
	tc := oauth2.NewClient(ctx, ts)
	client := github.NewClient(tc)
	return &Mock{
		client,
		ctx,
		owner,
		repo,
		reconcileBranchExists,
		targetAhead,
		mergeSuccessful,
	}, nil
}

func (gh *Mock) MergeBranches(base string, head string) (bool, error) {
	if gh.mergeSuccessful {
		return true, nil
	} else {
		return false, nil
	}
}

func (gh *Mock) GetBranch(branchName string) (*github.Branch, int, error) {
	dummySHA := "dd0b557d0696d2e1b8a1cf9de6b3c6d3a3a8a8f9"
	var status int
	if gh.reconcileBranchExists {
		status = 200
	} else {
		status = 404
	}
	return &github.Branch{
		Name: &branchName,
		Commit: &github.RepositoryCommit{
			SHA: &dummySHA,
		},
	}, status, nil
}

func (gh *Mock) CompareCommits(branch1 *github.Branch, branch2 *github.Branch) (*github.CommitsComparison, error) {
	var aheadBy int
	if gh.targetAhead {
		aheadBy = 2
	} else {
		aheadBy = 0
	}
	return &github.CommitsComparison{
		// head is ahead of base
		AheadBy: &aheadBy,
	}, nil
}

func (gh *Mock) DeleteBranch(branchName string) error {
	return nil
}

func (gh *Mock) GetBranchRef(branchName string) (*github.Reference, error) {
	return &github.Reference{}, nil
}

func (gh *Mock) CreateBranch(branchName string, target *github.Reference) error {
	return nil
}

func (gh *Mock) CreatePullRequest(head, base string) (*github.PullRequest, error) {
	prNumber := 1
	url := fmt.Sprintf("www.github.com/%s/%s/pulls/%v", gh.owner, gh.repo, prNumber)
	return &github.PullRequest{
		Number:  &prNumber,
		HTMLURL: &url,
	}, nil
}

func (gh *Mock) ListPullRequests() ([]*github.PullRequest, error) {
	var prs []*github.PullRequest
	reconcileBranchName := fmt.Sprintf("reconcile/%s-%s", "feature", "main")
	pr := &github.PullRequest{
		Head: &github.PullRequestBranch{
			Ref: github.String("feature"),
		},
		Base: &github.PullRequestBranch{
			Ref: github.String(reconcileBranchName),
		},
		Mergeable: github.Bool(true),
	}
	prs = append(prs, pr)
	return prs, nil
}
