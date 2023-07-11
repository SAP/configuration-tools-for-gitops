package github

import (
	"context"
	"fmt"
	"net/http"
	"strings"

	"github.com/SAP/configuration-tools-for-gitops/pkg/terminal"
	gogithub "github.com/google/go-github/v51/github"
	"golang.org/x/oauth2"
)

type Interface interface {
	CompareCommits(branch1 *gogithub.Branch, branch2 *gogithub.Branch) (*gogithub.CommitsComparison, error)
	CreateBranch(branchName string, target *gogithub.Reference) error
	CreatePullRequest(head, base string) (*gogithub.PullRequest, error)
	DeleteBranch(branchName string, force bool) error
	GetBranch(branchName string) (*gogithub.Branch, int, error)
	GetBranchRef(branchName string) (*gogithub.Reference, error)
	ListPullRequests() ([]*gogithub.PullRequest, error)
	MergeBranches(base, head string) (bool, error)
	MergePullRequest(pr int) (int, error)
}

func New(ctx context.Context, token, owner, repo, baseURL string, isEnterprise bool) (Interface, error) {
	// Authenticate with Github
	ts := oauth2.StaticTokenSource(
		&oauth2.Token{AccessToken: token},
	)
	tc := oauth2.NewClient(ctx, ts)
	var client *gogithub.Client
	var err error

	if isEnterprise {
		client, err = gogithub.NewEnterpriseClient(baseURL, baseURL, tc)
		if err != nil {
			return nil, err
		}
	} else {
		client = gogithub.NewClient(tc)
	}

	return &github{
		client,
		ctx,
		owner,
		repo,
	}, nil
}

type github struct {
	client *gogithub.Client
	ctx    context.Context
	owner  string
	repo   string
}

func (gh *github) MergeBranches(base, head string) (bool, error) {
	merge := &gogithub.RepositoryMergeRequest{
		CommitMessage: gogithub.String("Merge branch " + head + " into " + base),
		Base:          gogithub.String(base),
		Head:          gogithub.String(head),
	}
	_, response, err := gh.client.Repositories.Merge(gh.ctx, gh.owner, gh.repo, merge)

	// Merge conflict
	if response.StatusCode == http.StatusConflict {
		return false, nil
	}

	if err != nil {
		return false, err
	}
	// checkout https://docs.github.com/en/rest/branches/branches?apiVersion=2022-11-28#merge-a-branch
	// Success response
	if response.StatusCode == http.StatusCreated || response.StatusCode == http.StatusNoContent {
		return true, nil
	}
	return false, fmt.Errorf("github server error(%v): %v", response.StatusCode, response.Status)
}

func (gh *github) GetBranch(branchName string) (*gogithub.Branch, int, error) {
	branch, response, err := gh.client.Repositories.GetBranch(gh.ctx, gh.owner, gh.repo, branchName, true)
	return branch, response.StatusCode, err
}

func (gh *github) CompareCommits(
	branch1 *gogithub.Branch, branch2 *gogithub.Branch,
) (*gogithub.CommitsComparison, error) {
	options := &gogithub.ListOptions{}
	commits, _, err := gh.client.Repositories.CompareCommits(
		gh.ctx,
		gh.owner,
		gh.repo,
		branch1.GetCommit().GetSHA(),
		branch2.GetCommit().GetSHA(), options)
	return commits, err
}

func (gh *github) DeleteBranch(branchName string, forceDelete bool) error {
	if !forceDelete {
		printTerminal(
			fmt.Sprintf(
				"\n\nYou will lose all the changes made in the reconcile branch. "+
					"Are you sure you want to delete the branch %s?\n\n"+
					"Enter [y] for Yes and [n] for No: ",
				branchName,
			))

		input, err := readTerminal()
		if err != nil {
			printTerminal("abort on user input")
			return nil
		}
		if strings.EqualFold(input, "y") {
			forceDelete = true
		}
	}

	if forceDelete {
		_, err := gh.client.Git.DeleteRef(
			gh.ctx, gh.owner, gh.repo, "refs/heads/"+branchName,
		)
		if err != nil {
			return fmt.Errorf("failed to delete branch %q: %w", branchName, err)
		}
		printTerminal(fmt.Sprintf("%q branch deleted successfully", branchName))
		return nil
	}
	return fmt.Errorf("delete branch %q aborted due to user input", branchName)
}

func (gh *github) GetBranchRef(branchName string) (*gogithub.Reference, error) {
	branchRef := "refs/heads/" + branchName
	branch, _, err := gh.client.Git.GetRef(gh.ctx, gh.owner, gh.repo, branchRef)
	return branch, err
}

func (gh *github) CreateBranch(branchName string, target *gogithub.Reference) error {
	_, _, err := gh.client.Git.CreateRef(gh.ctx, gh.owner, gh.repo, &gogithub.Reference{
		Ref:    gogithub.String("refs/heads/" + branchName),
		Object: target.Object,
	})

	return err
}

func (gh *github) CreatePullRequest(head, base string) (*gogithub.PullRequest, error) {
	pr, _, err := gh.client.PullRequests.Create(gh.ctx, gh.owner, gh.repo, &gogithub.NewPullRequest{
		Title: gogithub.String("Draft PR: Merge " + head + " into " + base),
		Head:  gogithub.String(head),
		Base:  gogithub.String(base),
		Body: gogithub.String(
			"This is an auto-generated draft pull request for merging " +
				head +
				" into " +
				base),
		Draft: gogithub.Bool(true),
	})
	return pr, err
}

func (gh *github) ListPullRequests() ([]*gogithub.PullRequest, error) {
	prs, _, err := gh.client.PullRequests.List(gh.ctx, gh.owner, gh.repo, nil)

	return prs, err
}

func (gh *github) MergePullRequest(pr int) (int, error) {
	_, res, err := gh.client.PullRequests.Merge(gh.ctx, gh.owner, gh.repo, pr, "", nil)

	return res.StatusCode, err
}

var (
	printTerminal = terminal.Output
	readTerminal  = terminal.ReadStr
)
