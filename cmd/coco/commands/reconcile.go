package commands

import (
	"os"

	"github.com/configuration-tools-for-gitops/cmd/coco/reconcile"
	"github.com/configuration-tools-for-gitops/pkg/log"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"go.uber.org/zap"
)

var (
	owner  string
	repo   string
	dryRun bool
)

var reconcileCmd = &cobra.Command{
	Use:     "reconcile",
	Aliases: []string{"reconcile"},
	Short:   "Reconciles a target branch with source branch",
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
		client, err := reconcile.New(
			sourceBranch,
			targetBranch,
			owner,
			repo,
			viper.GetString("git-token"),
		)
		if err != nil {
			log.Sugar.Errorf("reconciliation failed with: %w", err)
			os.Exit(1)
		}

		err = client.Reconcile(dryRun)
		if err != nil {
			log.Sugar.Errorf("reconciliation failed with: %w", err)
			os.Exit(1)
		}
	},
}

//nolint:gochecknoinits // required by the cobra framework
func init() {
	if err := log.Init(logLvl, "2006-01-02T15:04:05Z07:00", true); err != nil {
		zap.S().Fatal(err)
	}
	rootCmd.AddCommand(reconcileCmd)
	reconcileCmd.PersistentFlags().StringVarP(&sourceBranch, "source", "s", "", "The souce branch to reconcile from.")
	if err := reconcileCmd.MarkPersistentFlagRequired("source"); err != nil {
		log.Sugar.Error(err)
		os.Exit(1)
	}
	reconcileCmd.PersistentFlags().StringVarP(&targetBranch, "target", "t", "", "The target branch to reconcile to.")
	if err := reconcileCmd.MarkPersistentFlagRequired("target"); err != nil {
		log.Sugar.Error(err)
		os.Exit(1)
	}
	reconcileCmd.PersistentFlags().StringVarP(&repo, "repo", "", "", "The name of the gihtub repository.")
	if err := reconcileCmd.MarkPersistentFlagRequired("repo"); err != nil {
		log.Sugar.Error(err)
		os.Exit(1)
	}
	reconcileCmd.PersistentFlags().StringVarP(&owner, "owner", "", "", "The account owner of the github repository.")
	if err := reconcileCmd.MarkPersistentFlagRequired("owner"); err != nil {
		log.Sugar.Error(err)
		os.Exit(1)
	}
	reconcileCmd.PersistentFlags().BoolVarP(&dryRun, "dry-run", "d", false, "Perform a dry-run to check for merge conflicts without making any changes.")
}
