package reconcile

import (
	"context"
	"fmt"
	"testing"

	"github.com/SAP/configuration-tools-for-gitops/pkg/github"
	"github.com/SAP/configuration-tools-for-gitops/pkg/log"
	"go.uber.org/zap"
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
	if err := log.Init(log.Debug(), "", true); err != nil {
		zap.S().Fatal(err)
	}
	for _, tt := range scenarios {
		t.Run(tt.title, func(t *testing.T) {
			newGithubClient = func(token, owner, repo string, ctx context.Context) (githubClient, error) {
				return github.NewMock(
					token,
					owner,
					repo,
					ctx,
					tt.reconcileBranchExists,
					tt.targetAhead,
					tt.mergeSuccessful,
				)
			}
			client, err := New(tt.sourceBranch, tt.targetBranch, tt.owner, tt.repo, token)
			if err != nil && err.Error() != tt.expectedErr.Error() {
				t.Errorf("unexpected error: got %v, want %v", err, tt.expectedErr)
			}
			err = client.Reconcile(tt.dryRun)
			if err != nil && err.Error() != tt.expectedErr.Error() {
				t.Errorf("unexpected error: got %v, want %v", err, tt.expectedErr)
			}
		})
	}
}