package commands

import (
	"os"

	"github.com/configuration-tools-for-gitops/cmd/coco/reconcile"
	"github.com/configuration-tools-for-gitops/pkg/log"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var (
	owner  string
	repo   string
	dryRun bool
)

var reconcileCmd = &cobra.Command{
	Use:     "reconcile",
	Aliases: []string{"r"},
	Short:   "Reconciles a target branch with source branch",
	Long: `The command is intended to reconcile a target branch with a source branch
	 by merging them. The reconciling process involves creating a new branch with the 
	 name "reconcile/{target_branch}," where {target_branch} is the name of the 
	 target branch, merging the source branch into the target branch, and 
	 pushing the result to the remote repository`,
	PreRun: func(cmd *cobra.Command, args []string) {
		if viper.GetString("git-token") == "" {
			cobra.CheckErr(
				"environment variable \"GITHUB_TOKEN\" must be set for the \"dependencies\" command.",
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
		err := reconcile.Reconcile(
			sourceBranch,
			targetBranch,
			owner,
			repo,
			viper.GetString("git-token"),
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
	reconcileCmd.PersistentFlags().StringVarP(&sourceBranch, "source-branch", "s", "", "The souce branch to reconcile from.")
	if err := reconcileCmd.MarkFlagRequired("source-branch"); err != nil {
		log.Sugar.Error(err)
		os.Exit(1)
	}
	reconcileCmd.PersistentFlags().StringVarP(&targetBranch, "target-branch", "t", "", "The target branch to reconcile to.")
	if err := reconcileCmd.MarkFlagRequired("target-branch"); err != nil {
		log.Sugar.Error(err)
		os.Exit(1)
	}
	reconcileCmd.PersistentFlags().StringVarP(&repo, "repo", "r", "", "The name of the gihtub repository.")
	registerFlag(repo, "repo", "GITHUB_REPOSITORY")
	if err := reconcileCmd.MarkFlagRequired("repo"); err != nil {
		log.Sugar.Error(err)
		os.Exit(1)
	}
	reconcileCmd.PersistentFlags().StringVarP(&owner, "owner", "o", "", "The account owner of the github repository.")
	registerFlag(repo, "owner", "REPOSITORY_OWNER")
	if err := reconcileCmd.MarkFlagRequired("owner"); err != nil {
		log.Sugar.Error(err)
		os.Exit(1)
	}
	reconcileCmd.PersistentFlags().BoolVarP(&dryRun, "dry-run", "d", false, "Perform a dry-run to check for merge conflicts without making any changes.")
}
