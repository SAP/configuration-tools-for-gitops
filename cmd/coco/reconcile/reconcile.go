package reconcile

import (
	"context"
	"fmt"
	"os"
	"strings"

	"github.com/google/go-github/v51/github"
	"golang.org/x/oauth2"
)

func StartReconcilition(sourceBranch string, targetBranch string, ownerName string, repoName string, dryRun bool) error {
	if sourceBranch == "" || targetBranch == "" {
		return fmt.Errorf("source and target branches must be specified")
	}

	if ownerName == "" || repoName == "" {
		return fmt.Errorf("owner name and repository name must be specified")
	}

	reconcileBranchName := fmt.Sprintf("reconcile/%s", targetBranch)

	ctx := context.Background()

	// Authenticate with Github
	client, err := authenticateWithGithub(ctx)
	if err != nil {
		return fmt.Errorf("failed to authenticate with Github: %v", err)
	}

	merge := &github.RepositoryMergeRequest{
		CommitMessage: github.String("Merge branch " + sourceBranch + " into " + targetBranch),
		Base:          github.String(targetBranch),
		Head:          github.String(sourceBranch),
	}

	_, _, err = client.Repositories.Merge(ctx, ownerName, repoName, merge)
	if err != nil {
		if strings.Contains(err.Error(), "Merge conflict") {
			return handleMergeConflict(ctx, client, ownerName, repoName, reconcileBranchName, targetBranch, sourceBranch, dryRun)
		} else {
			return fmt.Errorf("failed to merge branches: %v", err)
		}
	} else {
		if dryRun {
			fmt.Println("No merge conflicts found")
			return nil
		}
		fmt.Println("Merged successfully")
	}

	fmt.Println("Reconcile complete")
	return nil
}

func authenticateWithGithub(ctx context.Context) (*github.Client, error) {
	token := os.Getenv("GITHUB_TOKEN")
	if token == "" {
		return nil, fmt.Errorf("GITHUB_TOKEN environment variable is not set")
	}
	ts := oauth2.StaticTokenSource(
		&oauth2.Token{AccessToken: token},
	)
	tc := oauth2.NewClient(ctx, ts)
	return github.NewClient(tc), nil
}

func handleMergeConflict(ctx context.Context, client *github.Client, ownerName, repoName, reconcileBranchName, targetBranch, sourceBranch string, dryRun bool) error {
	if dryRun {
		return fmt.Errorf("merge conflicts detected")
	}

	// get a list of branches
	branches, _, err := client.Repositories.ListBranches(ctx, ownerName, repoName, nil)
	if err != nil {
		return fmt.Errorf("failed to list branches: %v", err)
	}

	// check if reconcile/target branch is already created
	reconcileBranchExists := false
	var reconcileBranch *github.Branch
	for _, b := range branches {
		if b.GetName() == reconcileBranchName {
			reconcileBranchExists = true
			reconcileBranch = b
			break
		}
	}

	if reconcileBranchExists {
		return handleExistingReconcileBranch(ctx, client, ownerName, repoName, reconcileBranchName, targetBranch, reconcileBranch, sourceBranch)
	}

	return handleNewReconcileBranch(ctx, client, ownerName, repoName, reconcileBranchName, targetBranch, sourceBranch)
}

func handleExistingReconcileBranch(ctx context.Context, client *github.Client, ownerName, repoName, reconcileBranchName, targetBranch string, reconcileBranch *github.Branch, sourceBranch string) error {
	// Compare the latest target branch and reconcile/target branch
	target, _, err := client.Repositories.GetBranch(ctx, ownerName, repoName, targetBranch, false)
	if err != nil {
		return fmt.Errorf("failed to get target branch: %v", err)
	}
	if target.GetCommit().GetSHA() != reconcileBranch.GetCommit().GetSHA() {
		// Check if there are new commits in target branch
		options := &github.ListOptions{}
		commits, _, err := client.Repositories.CompareCommits(
			ctx,
			ownerName,
			repoName,
			reconcileBranch.GetCommit().GetSHA(),
			target.GetCommit().GetSHA(), options)
		if err != nil {
			return fmt.Errorf("failed to compare commits: %v", err)
		}
		if len(commits.Commits) > 0 {
			// Delete the reconcile/target branch
			if _, err := client.Git.DeleteRef(ctx, ownerName, repoName,
				"refs/heads/"+reconcileBranchName); err != nil {
				return fmt.Errorf("failed to delete branch: %v", err)
			}
			fmt.Printf("Deleted existing reconcile branch: %s\n", reconcileBranchName)
		} else {
			// Merge PR/merge directly
			// TODO: Implement PR/merge logic
			return nil
		}
	} else {
		fmt.Println("Reconcile branch is up to date with target branch")
		// Merge PR/merge directly
		// TODO: Implement PR/merge logic
		return nil
	}
	return nil
}

func handleNewReconcileBranch(ctx context.Context, client *github.Client, ownerName, repoName, reconcileBranchName, targetBranch string, sourceBranch string) error {
	// Create a new branch reconcile/target branch from target branch
	targetRef := "refs/heads/" + targetBranch
	target, _, err := client.Git.GetRef(ctx, ownerName, repoName, targetRef)
	if err != nil {
		return fmt.Errorf("Failed to get target branch reference: %v", err)
	}
	if _, _, err = client.Git.CreateRef(ctx, ownerName, repoName, &github.Reference{
		Ref:    github.String("refs/heads/" + reconcileBranchName),
		Object: target.Object,
	}); err != nil {
		return fmt.Errorf("Failed to create reconcile branch: %v", err)
	}
	fmt.Println("Created new reconcile branch from target branch")

	pr, _, err := client.PullRequests.Create(ctx, ownerName, repoName, &github.NewPullRequest{
		Title: github.String("Draft PR: Merge " + sourceBranch + " into " + reconcileBranchName),
		Head:  github.String(sourceBranch),
		Base:  github.String(reconcileBranchName),
		Body: github.String(
			"This is an auto-generated draft pull request for merging " +
				sourceBranch +
				" into " +
				reconcileBranchName),
		Draft: github.Bool(true),
	})
	if err != nil {
		return fmt.Errorf("failed to create a draft PR: %v", err)
	}
	fmt.Printf("Draft pull request #%d created: %s\n", pr.GetNumber(), pr.GetHTMLURL())
	return nil
}
