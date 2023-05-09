package githubclient

import (
	"context"
	"fmt"
	"strings"

	"github.com/google/go-github/v51/github"
	"golang.org/x/oauth2"
)

type Github struct {
	client *github.Client
	ctx    context.Context
	owner  string
	repo   string
}

func New(token, owner, repo, base, head, reconcileBranchName string, ctx context.Context) (*Github, error) {
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

func (gh *Github) MergeBranches(base string, head string) error {
	merge := &github.RepositoryMergeRequest{
		CommitMessage: github.String("Merge branch " + head + " into " + base),
		Base:          github.String(base),
		Head:          github.String(head),
	}
	_, _, err := gh.client.Repositories.Merge(gh.ctx, gh.owner, gh.repo, merge)
	return err
}

func (gh *Github) GetBranch(branchName string) (*github.Branch, error) {
	branch, _, err := gh.client.Repositories.GetBranch(gh.ctx, gh.owner, gh.repo, branchName, true)
	return branch, err
}

func (gh *Github) CompareCommits(branch1 *github.Branch, branch2 *github.Branch) (*github.CommitsComparison, error) {
	options := &github.ListOptions{}
	commits, _, err := gh.client.Repositories.CompareCommits(
		gh.ctx,
		gh.owner,
		gh.repo,
		branch1.GetCommit().GetSHA(),
		branch2.GetCommit().GetSHA(), options)
	return commits, err
}

func (gh *Github) DeleteBranch(branchName string) error {
	warningPrompt := fmt.Sprintf("\n\nYou will lose all the changes made in the reconcile branch. Are you sure you want to delete the branch %s?\n\n", branchName) +
		"Enter [y] for Yes and [n] for No: "
	fmt.Print(warningPrompt)
	var input string
	fmt.Scanln(&input)

	if strings.ToLower(input) == "y" {
		_, err := gh.client.Git.DeleteRef(gh.ctx, gh.owner, gh.repo,
			"refs/heads/"+branchName)
		if err == nil {
			fmt.Printf("%s branch deleted successfully", branchName)
		} else {
			return fmt.Errorf("failed to delete branch %s: %v", branchName, err)
		}
		return err
	}

	fmt.Println("Exiting gracefully.")

	return nil
}

func (gh *Github) GetBranchRef(branchName string) (*github.Reference, error) {
	branchRef := "refs/heads/" + branchName
	branch, _, err := gh.client.Git.GetRef(gh.ctx, gh.owner, gh.repo, branchRef)
	return branch, err
}

func (gh *Github) CreateBranch(branchName string, target *github.Reference) error {
	_, _, err := gh.client.Git.CreateRef(gh.ctx, gh.owner, gh.repo, &github.Reference{
		Ref:    github.String("refs/heads/" + branchName),
		Object: target.Object,
	})

	return err
}

func (gh *Github) CreatePullRequest(head, base string) (*github.PullRequest, error) {
	pr, _, err := gh.client.PullRequests.Create(gh.ctx, gh.owner, gh.repo, &github.NewPullRequest{
		Title: github.String("Draft PR: Merge " + head + " into " + base),
		Head:  github.String(head),
		Base:  github.String(base),
		Body: github.String(
			"This is an auto-generated draft pull request for merging " +
				head +
				" into " +
				base),
		Draft: github.Bool(true),
	})
	return pr, err
}

func (gh *Github) ListPullRequests() ([]*github.PullRequest, error) {
	prs, _, err := gh.client.PullRequests.List(gh.ctx, gh.owner, gh.repo, nil)

	return prs, err
}
