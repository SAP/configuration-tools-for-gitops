package githubclient

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
	repo string,
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
	}, nil
}

func (gh *Mock) MergeBranches() error {
	if gh.mergeSuccessful {
		return nil
	} else {
		return fmt.Errorf("Merge conflict")
	}
}

func (gh *Mock) GetBranch(branchName string) (*github.Branch, error) {
	dummySHA := "dd0b557d0696d2e1b8a1cf9de6b3c6d3a3a8a8f9"
	return &github.Branch{
		Name: &gh.base,
		Commit: &github.RepositoryCommit{
			SHA: &dummySHA,
		},
	}, nil
}

func (gh *Mock) CompareCommits(branch1 *github.Branch, branch2 *github.Branch) (*github.CommitsComparison, error) {
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
	var nilSlice []*github.PullRequest
	return nilSlice, nil
}
