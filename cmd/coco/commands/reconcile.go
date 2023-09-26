package commands

import (
	"context"
	"fmt"
	"net/url"
	"time"

	"github.com/SAP/configuration-tools-for-gitops/v2/cmd/coco/reconcile"
	"github.com/SAP/configuration-tools-for-gitops/v2/pkg/log"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var (
	owner          string
	repositoryName string
	sourceRemote   string
	targetRemote   string
	forceReconcile bool
)

var (
	timeout = 5 * time.Minute
)

func newReconcile() *cobra.Command {
	var c = &cobra.Command{
		Use:   "reconcile",
		Short: "Reconciles a target branch with source branch",
		Long: `The command is intended to reconcile a target branch with a source branch
		by merging them. The reconciling process involves creating a new branch with the
		name "reconcile/{target_branch}," where {target_branch} is the name of the
		target branch, merging the source branch into the target branch, and
		pushing the result to the remote repository`,
		PreRun: func(cmd *cobra.Command, args []string) {
			if !viper.IsSet("git-token") {
				cobra.CheckErr(`environment variable "GITHUB_TOKEN" must be set`)
			}
			missingParams := []string{}
			if sourceBranch == "" {
				missingParams = append(missingParams, "source")
			}
			if targetBranch == "" {
				missingParams = append(missingParams, "target")
			}
			if owner == "" {
				missingParams = append(missingParams, "owner")
			}
			if repositoryName == "" {
				missingParams = append(missingParams, "repository")
			}
			if !viper.IsSet(gitURLKey) {
				missingParams = append(missingParams, "git-url")
			}
			if len(missingParams) != 0 {
				failOnError(fmt.Errorf("the CLI parameters %v must be set", missingParams), "reconcile")
			}
		},
		Run: func(cmd *cobra.Command, args []string) {
			ctx, cancel := context.WithTimeout(context.Background(), timeout)
			defer cancel()
			githubBaseURL, err := url.Parse(viper.GetString(gitURLKey))
			failOnError(err, "reconcile")

			client, err := reconcile.New(
				ctx,
				owner,
				repositoryName,
				viper.GetString("git-token"),
				fmt.Sprintf("https://%s", githubBaseURL.Hostname()),
				reconcile.BranchConfig{Name: targetBranch, Remote: targetRemote},
				reconcile.BranchConfig{Name: sourceBranch, Remote: sourceRemote},
				log.Sugar,
			)
			failOnError(err, "reconcile")

			failOnError(client.Reconcile(forceReconcile), "reconcile")
		},
	}

	c.PersistentFlags().StringVarP(&sourceBranch, "source", "s", "", "The source branch to reconcile from.")
	failOnError(c.MarkPersistentFlagRequired("source"), "reconcile")

	c.PersistentFlags().StringVar(&sourceRemote, "source-remote", "origin", `The URL for the source branch.
	Can be left out incase the source and target branches are in the same repository.`)

	c.PersistentFlags().StringVarP(&targetBranch, "target", "t", "", "The target branch to reconcile to.")
	failOnError(c.MarkPersistentFlagRequired("target"), "reconcile")

	c.PersistentFlags().StringVar(&targetRemote, "target-remote", "origin", `The URL for the target branch.
	Can be left out incase the source and target branches are in the same repository.`)

	c.PersistentFlags().StringVar(&repositoryName, "repo", "", "The name of the TARGET github repository.")
	failOnError(
		c.PersistentFlags().MarkDeprecated("repo", `please use "repository" flag instead.`),
		"reconcile",
	)
	c.PersistentFlags().StringVar(&repositoryName, "repository", "", "The name of the TARGET github repository.")
	failOnError(c.MarkPersistentFlagRequired("repository"), "reconcile")

	c.PersistentFlags().StringVar(&owner, "owner", "", "The account owner of the TARGET github repository.")
	failOnError(c.MarkPersistentFlagRequired("owner"), "reconcile")

	c.Flags().BoolVar(
		&forceReconcile, "force", false,
		`Allows coco to forcefully deletes the reconcile branch if required.`,
	)
	return c
}
