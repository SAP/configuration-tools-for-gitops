package githubclient

import (
	"context"
	"fmt"
	"strings"

	"github.com/configuration-tools-for-gitops/pkg/log"
	"github.com/google/go-github/v51/github"
	"golang.org/x/oauth2"
)

type Github struct {
	client              *github.Client
	ctx                 context.Context
	owner               string
	repo                string
	base                string
	head                string
	reconcileBranchName string
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
		base,
		head,
		reconcileBranchName,
	}, nil
}

func (gh *Github) Merge(dryRun bool) error {
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

func (gh *Github) handleMergeConflict(dryRun bool) error {
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

func (gh *Github) handleExistingReconcileBranch(reconcileBranch *github.Branch) (bool, error) {
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

func (gh *Github) handleNewReconcileBranch() error {
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

func (gh *Github) checkMergeability() (bool, error) {
	prs, _, err := gh.client.PullRequests.List(gh.ctx, gh.owner, gh.repo, nil)
	if err != nil {
		return false, err
	}
	var pr *github.PullRequest
	for _, p := range prs {
		if p.Head.GetRef() == gh.head && p.Base.GetRef() == gh.reconcileBranchName {
			pr = p
			break
		}
	}
	if pr != nil {
		// check if the pull request is mergable
		if pr.GetMergeable() {
			// perform the merge
			commitMessage := "Merge " + gh.reconcileBranchName + " into " + gh.base
			mergeRequest := &github.RepositoryMergeRequest{
				Base:          &gh.base,
				Head:          &gh.reconcileBranchName,
				CommitMessage: &commitMessage,
			}
			_, _, err := gh.client.Repositories.Merge(gh.ctx, gh.owner, gh.repo, mergeRequest)
			if err != nil {
				log.Sugar.Infof("Successfully merged %s to %s", gh.reconcileBranchName, gh.base)
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

func (gh *Github) handleTargetAhead() (bool, error) {
	fmt.Print("The target branch has new commits, choose one of the following options:\n\n" +
		"Option 1: Merge the target branch into the reconcile branch manually and rerun command `coco reconcile`\n\n" +
		"Option 2: Automatically delete the reconcile branch and rerun the command `coco reconcile`\n\n" +
		"Enter [1] for Option 1 or [2] for Option 2: ")
	var input int
	fmt.Scanln(&input)
	switch input {
	case 1:
		fmt.Printf("\nPlease delete the branch %s and rerun the `coco reconcile` command", gh.reconcileBranchName)
	case 2:
		return false, gh.deleteBranch(gh.reconcileBranchName)
	default:
		for input != 1 && input != 2 {
			fmt.Print("\nPlease choose either Option 1 or 2. Enter [1] for Option 1 or [2] for Option 2: ")
			fmt.Scanln(&input)
		}
		if input == 1 {
			fmt.Printf("\nPlease delete the branch %s and rerun the `coco reconcile` command", gh.reconcileBranchName)
		} else if input == 2 {
			return false, gh.deleteBranch(gh.reconcileBranchName)
		}
	}
	return true, nil
}

func (gh *Github) mergeBranches() error {
	merge := &github.RepositoryMergeRequest{
		CommitMessage: github.String("Merge branch " + gh.head + " into " + gh.base),
		Base:          github.String(gh.base),
		Head:          github.String(gh.head),
	}
	_, _, err := gh.client.Repositories.Merge(gh.ctx, gh.owner, gh.repo, merge)
	return err
}

func (gh *Github) getBranch(branchName string) (*github.Branch, error) {
	branch, _, err := gh.client.Repositories.GetBranch(gh.ctx, gh.owner, gh.repo, branchName, true)
	return branch, err
}

func (gh *Github) compareCommits(branch1 *github.Branch, branch2 *github.Branch) (*github.CommitsComparison, error) {
	options := &github.ListOptions{}
	commits, _, err := gh.client.Repositories.CompareCommits(
		gh.ctx,
		gh.owner,
		gh.repo,
		branch1.GetCommit().GetSHA(),
		branch2.GetCommit().GetSHA(), options)
	return commits, err
}

func (gh *Github) deleteBranch(branchName string) error {
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

func (gh *Github) getBranchRef(branchName string) (*github.Reference, error) {
	branchRef := "refs/heads/" + branchName
	branch, _, err := gh.client.Git.GetRef(gh.ctx, gh.owner, gh.repo, branchRef)
	return branch, err
}

func (gh *Github) createBranch(branchName string, target *github.Reference) error {
	_, _, err := gh.client.Git.CreateRef(gh.ctx, gh.owner, gh.repo, &github.Reference{
		Ref:    github.String("refs/heads/" + branchName),
		Object: target.Object,
	})

	return err
}

func (gh *Github) createPullRequest(head, base string) (*github.PullRequest, error) {
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
