package commands

import (
	"context"
	"os"
	"time"

	"github.com/SAP/configuration-tools-for-gitops/cmd/coco/reconcile"
	"github.com/SAP/configuration-tools-for-gitops/pkg/log"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

func newSecret() *cobra.Command {
	var c = &cobra.Command{
		Use:   "secret",
		Short: "Command root for handling sealed secrets",
		Long:  `The `,
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
			client, err := reconcile.New(
				sourceBranch,
				targetBranch,
				owner,
				repo,
				viper.GetString("git-token"),
				ctx,
			)
			if err != nil {
				log.Sugar.Errorf("reconciliation failed with: %w", err)
				os.Exit(1)
			}

			err = client.Reconcile()
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
	return c
}
