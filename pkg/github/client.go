package github

import (
	"context"
	"fmt"
	"strings"

	"github.com/SAP/configuration-tools-for-gitops/pkg/terminal"
	gogithub "github.com/google/go-github/v51/github"
	"golang.org/x/oauth2"
)

type Github struct {
	client *gogithub.Client
	ctx    context.Context
	owner  string
	repo   string
}

func New(token, owner, repo string, ctx context.Context) (*Github, error) {
	// Authenticate with Github
	ts := oauth2.StaticTokenSource(
		&oauth2.Token{AccessToken: token},
	)
	tc := oauth2.NewClient(ctx, ts)
	client := gogithub.NewClient(tc)
	return &Github{
		client,
		ctx,
		owner,
		repo,
	}, nil
}

func (gh *Github) MergeBranches(base string, head string) (bool, error) {
	merge := &gogithub.RepositoryMergeRequest{
		CommitMessage: gogithub.String("Merge branch " + head + " into " + base),
		Base:          gogithub.String(base),
		Head:          gogithub.String(head),
	}
	_, response, err := gh.client.Repositories.Merge(gh.ctx, gh.owner, gh.repo, merge)

	if err != nil {
		return false, err
	}
	// checkout https://docs.github.com/en/rest/branches/branches?apiVersion=2022-11-28#merge-a-branch
	// Success response
	if response.StatusCode == 201 || response.StatusCode == 204 {
		return true, nil
	}
	// Merge conflict
	if response.StatusCode == 409 {
		return false, nil
	}
	return false, fmt.Errorf("github server error(%v): %v", response.StatusCode, response.Status)
}

func (gh *Github) GetBranch(branchName string) (*gogithub.Branch, error) {
	branch, _, err := gh.client.Repositories.GetBranch(gh.ctx, gh.owner, gh.repo, branchName, true)
	return branch, err
}

func (gh *Github) CompareCommits(
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

func (gh *Github) DeleteBranch(branchName string) error {
	print(
		fmt.Sprintf(
			"\n\nYou will lose all the changes made in the reconcile branch. "+
				"Are you sure you want to delete the branch %s?\n\n"+
				"Enter [y] for Yes and [n] for No: ",
			branchName,
		))
	rawInput := read()
	input, ok := rawInput.(string)
	if !ok {
		print("abort on user input")
		return nil
	}

	if strings.EqualFold(input, "y") {
		_, err := gh.client.Git.DeleteRef(gh.ctx, gh.owner, gh.repo,
			"refs/heads/"+branchName)
		if err != nil {
			return fmt.Errorf("failed to delete branch %s: %w", branchName, err)
		}
		print(fmt.Sprintf("%s branch deleted successfully", branchName))
		return nil
	}
	return nil
}

func (gh *Github) GetBranchRef(branchName string) (*gogithub.Reference, error) {
	branchRef := "refs/heads/" + branchName
	branch, _, err := gh.client.Git.GetRef(gh.ctx, gh.owner, gh.repo, branchRef)
	return branch, err
}

func (gh *Github) CreateBranch(branchName string, target *gogithub.Reference) error {
	_, _, err := gh.client.Git.CreateRef(gh.ctx, gh.owner, gh.repo, &gogithub.Reference{
		Ref:    gogithub.String("refs/heads/" + branchName),
		Object: target.Object,
	})

	return err
}

func (gh *Github) CreatePullRequest(head, base string) (*gogithub.PullRequest, error) {
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

func (gh *Github) ListPullRequests() ([]*gogithub.PullRequest, error) {
	prs, _, err := gh.client.PullRequests.List(gh.ctx, gh.owner, gh.repo, nil)

	return prs, err
}

var (
	print = terminal.Output
	read  = terminal.Read
)
