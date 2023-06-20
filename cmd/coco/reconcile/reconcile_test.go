package reconcile

import (
	"context"
	"fmt"
	"testing"
	"time"

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
		title:                 "dry run mode with successful merge",
		sourceBranch:          "feature",
		targetBranch:          "main",
		owner:                 "test",
		repo:                  "repo",
		expectedErr:           nil,
		mergeSuccessful:       true,
		reconcileBranchExists: false,
	},
	{
		title:                 "dry run mode with successful merge",
		sourceBranch:          "feature",
		targetBranch:          "main",
		owner:                 "test",
		repo:                  "repo",
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
		expectedErr:           nil,
		mergeSuccessful:       false,
		reconcileBranchExists: true,
		targetAhead:           false,
		reconcileMergable:     true,
	},
	{
		title:                 "default unsuccessful merge with a reconcile branch & target is ahead & manualmerge false",
		sourceBranch:          "feature",
		targetBranch:          "main",
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
		sourceBranch:          "feature",
		targetBranch:          "main",
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
		sourceBranch:          "feature",
		targetBranch:          "main",
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
		sourceBranch:          "feature",
		targetBranch:          "main",
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

	for _, tt := range scenarios {
		t.Run(tt.title, func(t *testing.T) {
			githubClient = func(token, owner, repo, baseURL string, ctx context.Context,
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
				} else {
					return true, nil
				}
			}
			ctx, cancel := context.WithTimeout(context.Background(), timeout)
			defer cancel()
			client, err := New(tt.sourceBranch, tt.targetBranch, tt.owner, tt.repo, token, "", false, ctx)
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
