package commands

import (
	"os"

	"github.com/configuration-tools-for-gitops/cmd/coco/reconcile"
	"github.com/configuration-tools-for-gitops/pkg/log"
	"github.com/spf13/cobra"
)

var (
	ownerName string
	repoName  string
	dryRun    bool
	localOnly bool
	botMode   bool
)

var reconcileCmd = &cobra.Command{
	Use:     "reconcile",
	Aliases: []string{"r"},
	Short:   "Reconciles a target branch with source branch",
	Long:    `This command ...`,
	Run: func(cmd *cobra.Command, args []string) {
		err := reconcile.StartReconcilition(
			sourceBranch,
			targetBranch,
			ownerName,
			repoName,
			dryRun,
		)
		if err != nil {
			log.Sugar.Errorf("reconciliation failed with: %s", err)
			os.Exit(1)
		}

	},
}

//nolint:gochecknoinits // required by the cobra framework
func init() {
	reconcileCmd.Flags().StringVar(&sourceBranch, "source", "", "The souce branch to reconcile from.")
	reconcileCmd.Flags().StringVar(&targetBranch, "target", "", "The target branch to reconcile to.")
	reconcileCmd.Flags().StringVar(&repoName, "repoName", "", "The name of the repository.")
	reconcileCmd.Flags().StringVar(&ownerName, "ownerName", "", "The name of the owner of the repository.")
	reconcileCmd.Flags().BoolVar(&dryRun, "dry-run", false, "Perform a dry-run to check for merge conflicts without making any changes.")
}
