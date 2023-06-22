package commands

import (
	"context"
	"fmt"
	"net/url"
	"os"
	"time"

	"github.com/SAP/configuration-tools-for-gitops/cmd/coco/reconcile"
	"github.com/SAP/configuration-tools-for-gitops/pkg/log"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var (
	owner          string
	repo           string
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
			if viper.GetString("git-token") == "" {
				cobra.CheckErr(
					"environment variable \"GITHUB_TOKEN\" must be set for the \"reconcile\" command.",
				)
			}
		},
		Run: func(cmd *cobra.Command, args []string) {
			if sourceBranch == "" || targetBranch == "" {
				log.Sugar.Errorf("source and target branches must be specified")
				os.Exit(1)
			}

			if owner == "" || repo == "" {
				log.Sugar.Errorf("owner name and repository name must be specified")
				os.Exit(1)
			}
			ctx, cancel := context.WithTimeout(context.Background(), timeout)
			defer cancel()
			githubBaseURL, err := url.Parse(gitURL)
			if err != nil {
				log.Sugar.Errorf("reconciliation failed with: %w", err)
			}
			var client *reconcile.ReconcileClient
			client, err = reconcile.New(
				sourceBranch,
				targetBranch,
				owner,
				repo,
				viper.GetString("git-token"),
				fmt.Sprintf("https://%s", githubBaseURL.Hostname()),
				ctx,
			)
			if err != nil {
				log.Sugar.Errorf("reconciliation failed with: %w", err)
				os.Exit(1)
			}

			err = client.Reconcile(forceReconcile)
			if err != nil {
				log.Sugar.Errorf("reconciliation failed with: %w", err)
				os.Exit(1)
			}
		},
	}

	c.PersistentFlags().StringVarP(&sourceBranch, "source", "s", "", "The souce branch to reconcile from.")
	if err := c.MarkPersistentFlagRequired("source"); err != nil {
		log.Sugar.Error(err)
		os.Exit(1)
	}
	c.PersistentFlags().StringVarP(&targetBranch, "target", "t", "", "The target branch to reconcile to.")
	if err := c.MarkPersistentFlagRequired("target"); err != nil {
		log.Sugar.Error(err)
		os.Exit(1)
	}
	c.PersistentFlags().StringVarP(&repo, "repo", "", "", "The name of the gihtub repository.")
	if err := c.MarkPersistentFlagRequired("repo"); err != nil {
		log.Sugar.Error(err)
		os.Exit(1)
	}
	c.PersistentFlags().StringVarP(&owner, "owner", "", "", "The account owner of the github repository.")
	if err := c.MarkPersistentFlagRequired("owner"); err != nil {
		log.Sugar.Error(err)
		os.Exit(1)
	}
	c.Flags().BoolVar(
		&forceReconcile, "force", false,
		`Allows coco to forcefully deletes the reconcile branch if required.`,
	)
	return c
}
