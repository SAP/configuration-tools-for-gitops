package reconcile

import (
	"context"
	"fmt"
	"testing"

	"github.com/google/go-github/v51/github"
)

type scenario struct {
	title                 string
	sourceBranch          string
	targetBranch          string
	owner                 string
	repo                  string
	dryRun                bool
	expectedErr           error
	reconcileBranchExists bool
	targetAhead           bool
	mergeSuccessful       bool
	reconcileMergable     bool
}

var scenarios = []scenario{
	{
		title:                 "missing source branch",
		sourceBranch:          "",
		targetBranch:          "main",
		owner:                 "test",
		repo:                  "repo",
		dryRun:                false,
		expectedErr:           fmt.Errorf("source and target branches must be specified"),
		mergeSuccessful:       true,
		reconcileBranchExists: false,
	},
	{
		title:                 "missing target branch",
		sourceBranch:          "feature",
		targetBranch:          "",
		owner:                 "test",
		repo:                  "repo",
		dryRun:                false,
		expectedErr:           fmt.Errorf("source and target branches must be specified"),
		mergeSuccessful:       true,
		reconcileBranchExists: false,
	},
	{
		title:                 "missing owner name",
		sourceBranch:          "feature",
		targetBranch:          "main",
		owner:                 "",
		repo:                  "repo",
		dryRun:                false,
		expectedErr:           fmt.Errorf("owner name and repository name must be specified"),
		mergeSuccessful:       true,
		reconcileBranchExists: false,
	},
	{
		title:                 "missing repo name",
		sourceBranch:          "feature",
		targetBranch:          "main",
		owner:                 "test",
		repo:                  "",
		dryRun:                false,
		expectedErr:           fmt.Errorf("owner name and repository name must be specified"),
		mergeSuccessful:       true,
		reconcileBranchExists: false,
	},
	{
		title:                 "dry run mode with successful merge",
		sourceBranch:          "feature",
		targetBranch:          "main",
		owner:                 "test",
		repo:                  "repo",
		dryRun:                true,
		expectedErr:           nil,
		mergeSuccessful:       true,
		reconcileBranchExists: false,
	},
	{
		title:                 "dry run mode with unsuccessful merge",
		sourceBranch:          "feature",
		targetBranch:          "main",
		owner:                 "test",
		repo:                  "repo",
		dryRun:                true,
		expectedErr:           fmt.Errorf("merge conflicts detected"),
		mergeSuccessful:       false,
		reconcileBranchExists: false,
	},
	{
		title:                 "default unsuccessful merge with no reconcile branch",
		sourceBranch:          "feature",
		targetBranch:          "main",
		owner:                 "test",
		repo:                  "repo",
		dryRun:                false,
		expectedErr:           nil,
		mergeSuccessful:       false,
		reconcileBranchExists: false,
	},
	// // need to add mergability check in the pkg
	{
		title:                 "default unsuccessful merge with a reconcile branch & target not ahead",
		sourceBranch:          "feature",
		targetBranch:          "main",
		owner:                 "test",
		repo:                  "repo",
		dryRun:                false,
		expectedErr:           nil,
		mergeSuccessful:       false,
		reconcileBranchExists: true,
		targetAhead:           false,
		reconcileMergable:     true,
	},
	{
		title:                 "default unsuccessful merge with a reconcile branch & target is ahead",
		sourceBranch:          "feature",
		targetBranch:          "main",
		owner:                 "test",
		repo:                  "repo",
		dryRun:                false,
		expectedErr:           nil,
		mergeSuccessful:       false,
		reconcileBranchExists: true,
		targetAhead:           true,
	},
}

func TestReconcilition(t *testing.T) {
	token := "dummy_token_1234567890"
	getBranchRef = func(client *github.Client, ctx context.Context, owner, repo, branchName string) (*github.Reference, error) {
		return &github.Reference{}, nil
	}
	createBranch = func(client *github.Client, ctx context.Context, owner, repo, branchName string, target *github.Reference) error {
		return nil
	}
	handleTargetAhead = func(reconcileBranchName, owner, repo string, client *github.Client, ctx context.Context) (bool, error) {
		return true, nil
	}
	checkMergeability = func(ctx context.Context, reconcileBranchName, source, target, owner, repo string, client *github.Client) (bool, error) {
		return true, nil
	}
	for _, tt := range scenarios {
		t.Run(tt.title, func(t *testing.T) {
			createPullRequest = func(client *github.Client, ctx context.Context, owner, repo, head, base string) (*github.PullRequest, error) {
				prNumber := 1
				url := fmt.Sprintf("www.github.com/%s/%s/pulls/%v", tt.owner, tt.repo, prNumber)
				return &github.PullRequest{
					Number:  &prNumber,
					HTMLURL: &url,
				}, nil
			}
			compareCommits = func(client *github.Client, ctx context.Context, owner, repo string, branch1, branch2 *github.Branch) (*github.CommitsComparison, error) {
				var aheadBy int
				if tt.targetAhead {
					aheadBy = 2
				} else {
					aheadBy = 0
				}
				return &github.CommitsComparison{
					//head is ahead of base
					AheadBy: &aheadBy,
				}, nil
			}
			getBranch = func(client *github.Client, ctx context.Context, owner string, repo string, branchName string) (*github.Branch, error) {
				dummySHA := "dd0b557d0696d2e1b8a1cf9de6b3c6d3a3a8a8f9"
				return &github.Branch{
					Name: &tt.targetBranch,
					Commit: &github.RepositoryCommit{
						SHA: &dummySHA,
					},
				}, nil
			}
			mergeBranches = func(ctx context.Context, client *github.Client, owner string, repo string, targetBranch string, sourceBranch string) error {
				if tt.mergeSuccessful {
					return nil
				} else {
					return fmt.Errorf("Merge conflict")
				}
			}
			err := Reconcile(tt.sourceBranch, tt.targetBranch, tt.owner, tt.repo, token, tt.dryRun)
			if err != nil && err.Error() != tt.expectedErr.Error() {
				t.Errorf("unexpected error: got %v, want %v", err, tt.expectedErr)
			}
		})
	}
}
