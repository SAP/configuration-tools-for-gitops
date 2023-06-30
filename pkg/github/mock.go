package github

import (
	"fmt"
	"net/http"

	gogithub "github.com/google/go-github/v51/github"
)

type mock struct {
	owner                 string
	repo                  string
	reconcileBranchExists bool
	targetAhead           bool
	mergeSuccessful       bool
}

func Mock(
	owner, repo string,
	reconcileBranchExists, targetAhead, mergeSuccessful bool,
) (Interface, error) {
	return &mock{
		owner,
		repo,
		reconcileBranchExists,
		targetAhead,
		mergeSuccessful,
	}, nil
}

func (gh *mock) MergeBranches(base, head string) (bool, error) {
	if gh.mergeSuccessful {
		return true, nil
	}
	return false, nil
}

func (gh *mock) GetBranch(branchName string) (*gogithub.Branch, int, error) {
	dummySHA := "dd0b557d0696d2e1b8a1cf9de6b3c6d3a3a8a8f9"
	var status int
	if gh.reconcileBranchExists {
		status = 200
	} else {
		status = 404
	}
	return &gogithub.Branch{
		Name: &branchName,
		Commit: &gogithub.RepositoryCommit{
			SHA: &dummySHA,
		},
	}, status, nil
}

func (gh *mock) CompareCommits(branch1, branch2 *gogithub.Branch) (*gogithub.CommitsComparison, error) {
	var aheadBy int
	if gh.targetAhead {
		aheadBy = 2
	} else {
		aheadBy = 0
	}
	return &gogithub.CommitsComparison{
		// head is ahead of base
		AheadBy: &aheadBy,
	}, nil
}

func (gh *mock) DeleteBranch(branchName string, force bool) error {
	return nil
}

func (gh *mock) GetBranchRef(branchName string) (*gogithub.Reference, error) {
	return &gogithub.Reference{}, nil
}

func (gh *mock) CreateBranch(branchName string, target *gogithub.Reference) error {
	return nil
}

func (gh *mock) CreatePullRequest(head, base string) (*gogithub.PullRequest, error) {
	prNumber := 1
	url := fmt.Sprintf("www.gogithub.com/%s/%s/pulls/%v", gh.owner, gh.repo, prNumber)
	return &gogithub.PullRequest{
		Number:  &prNumber,
		HTMLURL: &url,
	}, nil
}

func (gh *mock) ListPullRequests() ([]*gogithub.PullRequest, error) {
	var prs []*gogithub.PullRequest
	reconcileBranchName := fmt.Sprintf("reconcile/%s-%s", "feature", "main")
	pr := &gogithub.PullRequest{
		Head: &gogithub.PullRequestBranch{
			Ref: gogithub.String("feature"),
		},
		Base: &gogithub.PullRequestBranch{
			Ref: gogithub.String(reconcileBranchName),
		},
		Mergeable: gogithub.Bool(true),
	}
	prs = append(prs, pr)
	return prs, nil
}

func (gh *mock) MergePullRequest(pr int) (int, error) {
	return http.StatusOK, nil
}
