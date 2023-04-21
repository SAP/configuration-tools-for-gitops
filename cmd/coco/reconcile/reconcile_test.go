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
	ownerName             string
	repoName              string
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
		ownerName:             "test",
		repoName:              "repo",
		dryRun:                false,
		expectedErr:           fmt.Errorf("source and target branches must be specified"),
		mergeSuccessful:       true,
		reconcileBranchExists: false,
	},
	{
		title:                 "missing target branch",
		sourceBranch:          "feature",
		targetBranch:          "",
		ownerName:             "test",
		repoName:              "repo",
		dryRun:                false,
		expectedErr:           fmt.Errorf("source and target branches must be specified"),
		mergeSuccessful:       true,
		reconcileBranchExists: false,
	},
	{
		title:                 "missing owner name",
		sourceBranch:          "feature",
		targetBranch:          "main",
		ownerName:             "",
		repoName:              "repo",
		dryRun:                false,
		expectedErr:           fmt.Errorf("owner name and repository name must be specified"),
		mergeSuccessful:       true,
		reconcileBranchExists: false,
	},
	{
		title:                 "missing repo name",
		sourceBranch:          "feature",
		targetBranch:          "main",
		ownerName:             "test",
		repoName:              "",
		dryRun:                false,
		expectedErr:           fmt.Errorf("owner name and repository name must be specified"),
		mergeSuccessful:       true,
		reconcileBranchExists: false,
	},
	{
		title:                 "dry run mode with successful merge",
		sourceBranch:          "feature",
		targetBranch:          "main",
		ownerName:             "test",
		repoName:              "repo",
		dryRun:                true,
		expectedErr:           nil,
		mergeSuccessful:       true,
		reconcileBranchExists: false,
	},
	{
		title:                 "dry run mode with unsuccessful merge",
		sourceBranch:          "feature",
		targetBranch:          "main",
		ownerName:             "test",
		repoName:              "repo",
		dryRun:                true,
		expectedErr:           fmt.Errorf("merge conflicts detected"),
		mergeSuccessful:       false,
		reconcileBranchExists: false,
	},
	{
		title:                 "default unsuccessful merge with no reconcile branch",
		sourceBranch:          "feature",
		targetBranch:          "main",
		ownerName:             "test",
		repoName:              "repo",
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
		ownerName:             "test",
		repoName:              "repo",
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
		ownerName:             "test",
		repoName:              "repo",
		dryRun:                false,
		expectedErr:           nil,
		mergeSuccessful:       false,
		reconcileBranchExists: true,
		targetAhead:           true,
	},
}

func TestReconcilition(t *testing.T) {
	token := "dummy_token_1234567890"
	getBranchRef = func(client *github.Client, ctx context.Context, ownerName, repoName, branchName string) (*github.Reference, error) {
		return &github.Reference{}, nil
	}
	createBranch = func(client *github.Client, ctx context.Context, ownerName, repoName, branchName string, target *github.Reference) error {
		return nil
	}
	handleTargetAhead = func(reconcileBranchName, ownerName, repoName string, client *github.Client, ctx context.Context) (bool, error) {
		return true, nil
	}
	checkMergeability = func(ctx context.Context, reconcileBranchName, source, target, ownerName, repoName string, client *github.Client) (bool, error) {
		return true, nil
	}
	for _, tt := range scenarios {
		t.Run(tt.title, func(t *testing.T) {
			createPullRequest = func(client *github.Client, ctx context.Context, ownerName, repoName, head, base string) (*github.PullRequest, error) {
				prNumber := 1
				url := fmt.Sprintf("www.github.com/%s/%s/pulls/%v", tt.ownerName, tt.repoName, prNumber)
				return &github.PullRequest{
					Number:  &prNumber,
					HTMLURL: &url,
				}, nil
			}
			compareCommits = func(client *github.Client, ctx context.Context, ownerName, repoName string, branch1, branch2 *github.Branch) (*github.CommitsComparison, error) {
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
			getBranch = func(client *github.Client, ctx context.Context, ownerName string, repoName string, branchName string) (*github.Branch, error) {
				dummySHA := "dd0b557d0696d2e1b8a1cf9de6b3c6d3a3a8a8f9"
				return &github.Branch{
					Name: &tt.targetBranch,
					Commit: &github.RepositoryCommit{
						SHA: &dummySHA,
					},
				}, nil
			}
			getBranchList = func(client *github.Client, ctx context.Context, ownerName string, repoName string) ([]*github.Branch, error) {
				if tt.reconcileBranchExists {
					reconcileBranchName := fmt.Sprintf("reconcile/%s-%s", tt.sourceBranch, tt.targetBranch)
					dummySHA := "a1b2c3d4e5f6g7h8i9j0k1l2m3n4o5p6q7r8s9t0"
					branchList := []*github.Branch{
						{
							Name: &reconcileBranchName,
							Commit: &github.RepositoryCommit{
								SHA: &dummySHA,
							},
						},
					}
					return branchList, nil
				} else {
					return []*github.Branch{}, nil
				}
			}
			mergeBranches = func(ctx context.Context, client *github.Client, ownerName string, repoName string, merge *github.RepositoryMergeRequest) error {
				if tt.mergeSuccessful {
					return nil
				} else {
					return fmt.Errorf("Merge conflict")
				}
			}
			err := Reconcile(tt.sourceBranch, tt.targetBranch, tt.ownerName, tt.repoName, token, tt.dryRun)
			if err != nil && err.Error() != tt.expectedErr.Error() {
				t.Errorf("unexpected error: got %v, want %v", err, tt.expectedErr)
			}
		})
	}
}
