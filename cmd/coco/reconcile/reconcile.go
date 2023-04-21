package reconcile

import (
	"context"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/configuration-tools-for-gitops/pkg/log"
	"github.com/google/go-github/v51/github"
	"golang.org/x/oauth2"
)

func Reconcile(sourceBranch string, targetBranch string, owner string, repo string, token string, dryRun bool) error {

	reconcileBranchName := fmt.Sprintf("reconcile/%s-%s", sourceBranch, targetBranch)

	timeout := 100 * time.Millisecond
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	// Authenticate with Github
	client, err := authenticateWithGithub(ctx)
	if err != nil {
		return fmt.Errorf("failed to authenticate with Github: %v", err)
	}

	err = mergeBranches(ctx, client, owner, repo, targetBranch, sourceBranch)
	if err != nil {
		if strings.Contains(err.Error(), "Merge conflict") {
			return handleMergeConflict(ctx, client, owner, repo, reconcileBranchName, targetBranch, sourceBranch, dryRun)
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

	log.Sugar.Info("Reconcile complete")
	return nil
}

func handleMergeConflict(ctx context.Context, client *github.Client, owner, repo, reconcileBranchName, targetBranch, sourceBranch string, dryRun bool) error {
	if dryRun {
		return fmt.Errorf("merge conflicts detected")
	}

	reconcileBranch, err := getBranch(client, ctx, owner, repo, reconcileBranchName)

	// get a list of branches
	// branches, err := getBranchList(client, ctx, owner, repo)

	if err == nil {
		var resolved bool
		resolved, err = handleExistingReconcileBranch(ctx, client, owner, repo, reconcileBranchName, targetBranch, reconcileBranch, sourceBranch)
		if err != nil {
			return err
		}
		if resolved {
			return nil
		}
	}

	return handleNewReconcileBranch(ctx, client, owner, repo, reconcileBranchName, targetBranch, sourceBranch)
}

func handleExistingReconcileBranch(ctx context.Context, client *github.Client, owner, repo, reconcileBranchName, targetBranch string, reconcileBranch *github.Branch, sourceBranch string) (bool, error) {
	// Compare the latest target branch and reconcile/target branch
	target, err := getBranch(client, ctx, owner, repo, targetBranch)
	if err != nil {
		return false, fmt.Errorf("failed to get target branch: %v", err)
	}
	commits, err := compareCommits(
		client,
		ctx,
		owner,
		repo,
		reconcileBranch,
		target)
	if err != nil {
		return false, fmt.Errorf("failed to compare commits: %v", err)
	}
	if commits.GetAheadBy() > 0 {
		return handleTargetAhead(reconcileBranchName, owner, repo, client, ctx)
	} else {
		//check mergability
		return checkMergeability(ctx, reconcileBranchName, sourceBranch, targetBranch, owner, repo, client)
		// return true, fmt.Errorf("%s already exists for the latest target branch", reconcileBranchName)
	}
}

func handleNewReconcileBranch(ctx context.Context, client *github.Client, owner, repo, reconcileBranchName, targetBranch string, sourceBranch string) error {
	// Create a new branch reconcile/target branch from target branch
	target, err := getBranchRef(client, ctx, owner, repo, targetBranch)
	if err != nil {
		return fmt.Errorf("Failed to get target branch reference: %v", err)
	}
	if err = createBranch(client, ctx, owner, repo, reconcileBranchName, target); err != nil {
		return fmt.Errorf("Failed to create reconcile branch: %v", err)
	}
	log.Sugar.Debug("Created new reconcile branch from target branch")

	pr, err := createPullRequest(client, ctx, owner, repo, sourceBranch, reconcileBranchName)
	if err != nil {
		return fmt.Errorf("failed to create a draft PR: %v", err)
	}
	log.Sugar.Info("Draft pull request #%d created: %s\n", pr.GetNumber(), pr.GetHTMLURL())
	return nil
}

var checkMergeability = func(ctx context.Context, reconcileBranchName, source, target, owner, repo string, client *github.Client) (bool, error) {
	prs, _, err := client.PullRequests.List(ctx, owner, repo, nil)
	if err != nil {
		return false, err
	}
	var pr *github.PullRequest
	for _, p := range prs {
		if p.Head.GetRef() == source && p.Base.GetRef() == reconcileBranchName {
			pr = p
			break
		}
	}
	if pr != nil {
		// check if the pull request is mergable
		if pr.GetMergeable() {
			// perform the merge
			commitMessage := "Merge " + reconcileBranchName + " into " + target
			mergeRequest := &github.RepositoryMergeRequest{
				Base:          &target,
				Head:          &reconcileBranchName,
				CommitMessage: &commitMessage,
			}
			_, _, err := client.Repositories.Merge(ctx, owner, repo, mergeRequest)
			if err != nil {
				log.Sugar.Info("Successfully merged reconcile branch to target branch")
				return true, err
			}
			return false, err
		} else {
			return false, fmt.Errorf("Please re-try after resolving the merge conflicts here: %s", pr.GetURL())
		}

	} else {
		return false, fmt.Errorf("the pull request was not found")
	}
}

var handleTargetAhead = func(reconcileBranchName string, owner string, repo string, client *github.Client, ctx context.Context) (bool, error) {
	fmt.Print("The target branch has new commits, choose one of the following options:\n\n" +
		"Option 1: Merge the target branch into the reconcile branch manually and rerun command `coco reconcile`\n\n" +
		"Option 2: Automatically delete the reconcile branch and rerun the command `coco reconcile`\n\n" +
		"Enter [1] for Option 1 or [2] for Option 2: ")
	var input int
	fmt.Scanln(&input)
	switch input {
	case 1:
		fmt.Printf("\nPlease delete the branch %s and rerun the `coco reconcile` command", reconcileBranchName)
	case 2:
		return false, deleteBranch(client, ctx, owner, repo, reconcileBranchName)
	default:
		for input != 1 && input != 2 {
			fmt.Print("\nPlease choose either Option 1 or 2. Enter [1] for Option 1 or [2] for Option 2: ")
			fmt.Scanln(&input)
		}
		if input == 1 {
			fmt.Printf("\nPlease delete the branch %s and rerun the `coco reconcile` command", reconcileBranchName)
		} else if input == 2 {
			return false, deleteBranch(client, ctx, owner, repo, reconcileBranchName)
		}
	}
	return true, nil
}

var mergeBranches = func(ctx context.Context, client *github.Client, owner string, repo string, base string, head string) error {
	merge := &github.RepositoryMergeRequest{
		CommitMessage: github.String("Merge branch " + head + " into " + base),
		Base:          github.String(base),
		Head:          github.String(head),
	}
	_, _, err := client.Repositories.Merge(ctx, owner, repo, merge)
	return err
}

var authenticateWithGithub = func(ctx context.Context) (*github.Client, error) {
	token := os.Getenv("GITHUB_TOKEN")
	if token == "" {
		return nil, fmt.Errorf("GITHUB_TOKEN environment variable is not set")
	}

	return oauthClient(ctx, token)
}

var oauthClient = func(ctx context.Context, token string) (*github.Client, error) {
	ts := oauth2.StaticTokenSource(
		&oauth2.Token{AccessToken: token},
	)
	tc := oauth2.NewClient(ctx, ts)
	return github.NewClient(tc), nil
}

var getBranch = func(client *github.Client, ctx context.Context, owner string, repo string, branchName string) (*github.Branch, error) {
	branch, _, err := client.Repositories.GetBranch(ctx, owner, repo, branchName, true)
	return branch, err
}

var compareCommits = func(client *github.Client, ctx context.Context, owner string, repo string, branch1 *github.Branch, branch2 *github.Branch) (*github.CommitsComparison, error) {
	options := &github.ListOptions{}
	commits, _, err := client.Repositories.CompareCommits(
		ctx,
		owner,
		repo,
		branch1.GetCommit().GetSHA(),
		branch2.GetCommit().GetSHA(), options)
	return commits, err
}

var deleteBranch = func(client *github.Client, ctx context.Context, owner string, repo string, branchName string) error {
	warningPrompt := fmt.Sprintf("\n\nYou will lose all the changes made in the reconcile branch. Are you sure you want to delete the branch %s?\n\n", branchName) +
		"Enter [y] for Yes and [n] for No: "
	fmt.Print(warningPrompt)
	var input string
	fmt.Scanln(&input)

	if strings.ToLower(input) == "y" {
		_, err := client.Git.DeleteRef(ctx, owner, repo,
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

var getBranchRef = func(client *github.Client, ctx context.Context, owner string, repo string, branchName string) (*github.Reference, error) {
	branchRef := "refs/heads/" + branchName
	branch, _, err := client.Git.GetRef(ctx, owner, repo, branchRef)
	return branch, err
}

var createBranch = func(client *github.Client, ctx context.Context, owner string, repo string, branchName string, target *github.Reference) error {
	_, _, err := client.Git.CreateRef(ctx, owner, repo, &github.Reference{
		Ref:    github.String("refs/heads/" + branchName),
		Object: target.Object,
	})

	return err
}

var createPullRequest = func(client *github.Client, ctx context.Context, owner string, repo string, head string, base string) (*github.PullRequest, error) {
	pr, _, err := client.PullRequests.Create(ctx, owner, repo, &github.NewPullRequest{
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
