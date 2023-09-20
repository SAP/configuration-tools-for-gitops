package reconcile

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/SAP/configuration-tools-for-gitops/v2/pkg/github"
	"github.com/SAP/configuration-tools-for-gitops/v2/pkg/log"
	"github.com/go-git/go-billy/v5/memfs"
	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing"
	"github.com/go-git/go-git/v5/storage/memory"
	"go.uber.org/zap"
)

type scenario struct {
	title                 string
	sourceBranch          BranchConfig
	targetBranch          BranchConfig
	owner                 string
	repo                  string
	expectedErr           error
	reconcileBranchExists bool
	targetAhead           bool
	mergeSuccessful       bool
	reconcileMergable     bool
	manualMerge           bool
	falseInput            bool
	force                 bool
}

var (
	timeout = 5 * time.Minute
)

var scenarios = []scenario{
	{
		title:                 "successful merge",
		sourceBranch:          BranchConfig{Name: "feature", Remote: "https://github.com/<remote-url>.git"},
		targetBranch:          BranchConfig{Name: "main", Remote: "https://github.com/<remote-url>.git"},
		owner:                 "test",
		repo:                  "repo",
		expectedErr:           nil,
		mergeSuccessful:       true,
		reconcileBranchExists: false,
	},
	{
		title:                 "target and source have different remotes",
		sourceBranch:          BranchConfig{Name: "feature", Remote: "https://github.com/<source-remote-url>.git"},
		targetBranch:          BranchConfig{Name: "main", Remote: "https://github.com/<target-remote-url>.git"},
		owner:                 "test",
		repo:                  "repo",
		expectedErr:           nil,
		mergeSuccessful:       true,
		reconcileBranchExists: false,
	},
	{
		title:                 "successful merge with reconcile branch exists",
		sourceBranch:          BranchConfig{Name: "feature", Remote: "https://github.com/<remote-url>.git"},
		targetBranch:          BranchConfig{Name: "main", Remote: "https://github.com/<remote-url>.git"},
		owner:                 "test",
		repo:                  "repo",
		expectedErr:           nil,
		mergeSuccessful:       true,
		reconcileBranchExists: true,
	},
	{
		title:                 "unsuccessful merge",
		sourceBranch:          BranchConfig{Name: "feature", Remote: "https://github.com/<remote-url>.git"},
		targetBranch:          BranchConfig{Name: "main", Remote: "https://github.com/<remote-url>.git"},
		owner:                 "test",
		repo:                  "repo",
		expectedErr:           fmt.Errorf("merge conflicts detected"),
		mergeSuccessful:       false,
		reconcileBranchExists: false,
	},
	{
		title:                 "default unsuccessful merge with no reconcile branch",
		sourceBranch:          BranchConfig{Name: "feature", Remote: "https://github.com/<remote-url>.git"},
		targetBranch:          BranchConfig{Name: "main", Remote: "https://github.com/<remote-url>.git"},
		owner:                 "test",
		repo:                  "repo",
		expectedErr:           nil,
		mergeSuccessful:       false,
		reconcileBranchExists: false,
	},
	// // need to add mergability check in the pkg
	{
		title:                 "default unsuccessful merge with a reconcile branch & target not ahead",
		sourceBranch:          BranchConfig{Name: "feature", Remote: "https://github.com/<remote-url>.git"},
		targetBranch:          BranchConfig{Name: "main", Remote: "https://github.com/<remote-url>.git"},
		owner:                 "test",
		repo:                  "repo",
		expectedErr:           nil,
		mergeSuccessful:       false,
		reconcileBranchExists: true,
		targetAhead:           false,
		reconcileMergable:     true,
	},
	{
		title:                 "default unsuccessful merge with a reconcile branch & target not ahead & no pr exists",
		sourceBranch:          BranchConfig{Name: "feature2", Remote: "https://github.com/<remote-url>.git"},
		targetBranch:          BranchConfig{Name: "main", Remote: "https://github.com/<remote-url>.git"},
		owner:                 "test",
		repo:                  "repo",
		expectedErr:           nil,
		mergeSuccessful:       false,
		reconcileBranchExists: true,
		targetAhead:           false,
		reconcileMergable:     true,
	},
	{
		title:                 "default unsuccessful merge with a reconcile branch & target is ahead & manualmerge false",
		sourceBranch:          BranchConfig{Name: "feature", Remote: "https://github.com/<remote-url>.git"},
		targetBranch:          BranchConfig{Name: "main", Remote: "https://github.com/<remote-url>.git"},
		owner:                 "test",
		repo:                  "repo",
		expectedErr:           nil,
		mergeSuccessful:       false,
		reconcileBranchExists: true,
		targetAhead:           true,
		manualMerge:           false,
	},
	{
		title:                 "default unsuccessful merge with a reconcile branch & target is ahead & force",
		sourceBranch:          BranchConfig{Name: "feature", Remote: "https://github.com/<remote-url>.git"},
		targetBranch:          BranchConfig{Name: "main", Remote: "https://github.com/<remote-url>.git"},
		owner:                 "test",
		repo:                  "repo",
		expectedErr:           nil,
		mergeSuccessful:       false,
		reconcileBranchExists: true,
		targetAhead:           true,
		manualMerge:           false,
		force:                 true,
	},
	{
		title:                 "default unsuccessful merge with a reconcile branch & target is ahead & manualmerge true",
		sourceBranch:          BranchConfig{Name: "feature", Remote: "https://github.com/<remote-url>.git"},
		targetBranch:          BranchConfig{Name: "main", Remote: "https://github.com/<remote-url>.git"},
		owner:                 "test",
		repo:                  "repo",
		expectedErr:           nil,
		mergeSuccessful:       false,
		reconcileBranchExists: true,
		targetAhead:           true,
		manualMerge:           true,
		falseInput:            false,
	},
	{
		title:                 "default unsuccessful merge with a reconcile branch & target is ahead & false input",
		sourceBranch:          BranchConfig{Name: "feature", Remote: "https://github.com/<remote-url>.git"},
		targetBranch:          BranchConfig{Name: "main", Remote: "https://github.com/<remote-url>.git"},
		owner:                 "test",
		repo:                  "repo",
		expectedErr:           fmt.Errorf("input must be in [y yes]: illegal input"),
		mergeSuccessful:       false,
		reconcileBranchExists: true,
		targetAhead:           true,
		manualMerge:           false,
		falseInput:            true,
	},
}

func TestReconcilition(t *testing.T) {
	token := "dummy_token_1234567890"
	if err := log.Init(log.Debug(), "", true); err != nil {
		zap.S().Fatal(err)
	}

	printTerminal = func(msg string) {
	}

	gitPull = func(worktree *git.Worktree, o *git.PullOptions) error {
		return nil
	}
	gitFetch = func(repo *git.Repository, o *git.FetchOptions) error {
		return nil
	}
	gitPush = func(repo *git.Repository, o *git.PushOptions) error {
		return nil
	}

	for _, tt := range scenarios {
		t.Run(tt.title, func(t *testing.T) {
			remoteName := fmt.Sprintf("reconcile/source/%s", tt.sourceBranch.Name)
			gitClone = func(path string, isBare bool, o *git.CloneOptions) (*git.Repository, error) {
				r, err := git.InitWithOptions(memory.NewStorage(), memfs.New(), git.InitOptions{
					DefaultBranch: "refs/heads/foo",
				})
				if err != nil {
					return nil, err
				}
				ref := plumbing.NewHashReference(
					plumbing.ReferenceName(fmt.Sprintf("refs/remotes/%s/%s", remoteName, tt.sourceBranch.Name)),
					plumbing.NewHash("dummytest"))
				return r, r.Storer.SetReference(ref)
			}
			gitOpen = func(path string) (*git.Repository, error) {
				r, err := git.InitWithOptions(memory.NewStorage(), memfs.New(), git.InitOptions{
					DefaultBranch: "refs/heads/foo",
				})
				if err != nil {
					return nil, err
				}
				ref := plumbing.NewHashReference(
					plumbing.ReferenceName(fmt.Sprintf("refs/remotes/%s/%s", remoteName, tt.sourceBranch.Name)),
					plumbing.NewHash("dummytest"))
				return r, r.Storer.SetReference(ref)
			}
			githubClient = func(ctx context.Context, stoken, owner, repo, baseURL string,
				isEnterprise bool) (github.Interface, error) {
				return github.Mock(
					owner, repo,
					tt.reconcileBranchExists,
					tt.targetAhead,
					tt.mergeSuccessful,
				)
			}
			confirmed = func() (bool, error) {
				if tt.falseInput {
					return false, fmt.Errorf("illegal input")
				}
				if tt.manualMerge {
					return false, nil
				}
				return true, nil
			}
			ctx, cancel := context.WithTimeout(context.Background(), timeout)
			defer cancel()
			client, err := New(
				ctx, tt.owner, tt.repo, token, "", tt.targetBranch, tt.sourceBranch, log.Sugar,
			)
			if err != nil && err.Error() != tt.expectedErr.Error() {
				t.Errorf("unexpected error: got %q, want %q", err, tt.expectedErr)
			}
			err = client.Reconcile(tt.force)
			if err != nil && err.Error() != tt.expectedErr.Error() {
				t.Errorf("unexpected error: got %q, want %q", err, tt.expectedErr)
			}
		})
	}
}
