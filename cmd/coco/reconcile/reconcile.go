package reconcile

import (
	"context"
	"fmt"
	"time"

	"github.com/configuration-tools-for-gitops/pkg/githubclient"
)

func Reconcile(sourceBranch, targetBranch, owner, repo, token string, dryRun bool) error {

	timeout := 100 * time.Millisecond
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	reconcileBranchName := fmt.Sprintf("reconcile/%s-%s", sourceBranch, targetBranch)

	// Authenticate with Github
	// target is base and source is head
	client, err := newGithubClient(token, owner, repo, targetBranch, sourceBranch, reconcileBranchName, ctx)
	if err != nil {
		return fmt.Errorf("failed to authenticate with Github: %v", err)
	}

	return client.Merge(dryRun)
}

var newGithubClient = func(token, owner, repo, base, head, reconcileBranchName string, ctx context.Context) (*githubclient.Github, error) {
	return githubclient.New(token, owner, repo, base, head, reconcileBranchName, ctx)
}
