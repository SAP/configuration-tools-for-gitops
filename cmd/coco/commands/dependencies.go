package commands

import (
	"io"
	"os"
	"path/filepath"

	"github.com/SAP/configuration-tools-for-gitops/cmd/coco/dependencies"
	"github.com/SAP/configuration-tools-for-gitops/cmd/coco/graph"
	"github.com/SAP/configuration-tools-for-gitops/pkg/log"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

const (
	allAllowed = 0777
)

var (
	outputFile, sourceBranch, targetBranch, rawFormat string

	format     graph.OutputFormat
	graphDepth int
)

func newDependencies() *cobra.Command {
	var c = &cobra.Command{
		Use:     "dependencies",
		Aliases: []string{"deps"},
		Short:   "Returns structured information which components and dependencies are affected by a change in git",
		Long: `The dependencies command finds all components and their downstream dependencies
	that are affected by a change from a source to a target commit.
	This is done by constructing the full dependency graph of upstream components
	(defined implicitly via the direct upstream dependencies given in the
	dependencies.yaml file of each component) and then inverting this graph.
	In addition all components that have changed between the source and the target
	commit are identified. Combining the dependency graph and the changed components
	gives the complete structure of change-affected components.`,
		Args: func(cmd *cobra.Command, args []string) error {
			return nil
		},
		PreRun: func(cmd *cobra.Command, args []string) {
			if viper.GetString("git-token") == "" {
				cobra.CheckErr(
					"environment variable \"GITHUB_TOKEN\" must be set for the \"dependencies\" command.",
				)
			}
			var ok bool
			format, ok = graph.CastOutputFormat(rawFormat)
			if !ok {
				log.Sugar.Errorf("illegal format %q", rawFormat)
				os.Exit(1)
			}
		},
		Run: func(cmd *cobra.Command, args []string) {
			changedDeps, err := dependencies.ChangeAffectedComponents(
				viper.GetString(gitURLKey),
				viper.GetString(gitRemoteKey),
				viper.GetString("git-token"),
				viper.GetString(gitPathKey),
				viper.GetString(componentCfg),
				sourceBranch,
				targetBranch,
				graphDepth,
				overWriteGitDepth, // viper.GetInt(gitDepth),
				logLvl,
			)
			if err != nil {
				log.Sugar.Errorf("dependency failed with: %s", err)
				os.Exit(1)
			}

			writeTo, err := writeTarget(viper.GetString(gitPathKey), outputFile)
			if err != nil {
				log.Sugar.Errorf("dependency failed with: %s", err)
				os.Exit(1)
			}
			changedDeps.Print(writeTo, format)
		},
	}

	c.AddCommand(newGraph())

	c.Flags().IntVar(
		&graphDepth, "depth", -1,
		`maximum depth for which downstream dependencies will be returned:
	-1: all dependencies
	0: only the components themselves
	1: components and direct dependencies
		`,
	)
	c.Flags().StringVar(
		&outputFile, "output", "",
		`specify an output filename where the results are stored. If empty results will
	be sent to stdout.`,
	)
	c.Flags().StringVarP(
		&sourceBranch, "source-branch", "s", "",
		`source branch for evaluating changed components`,
	)
	cobra.CheckErr(c.MarkFlagRequired("source-branch"))
	c.Flags().StringVarP(
		&targetBranch, "target-branch", "t", "main",
		`target branch for evaluating changed components`,
	)
	cobra.CheckErr(c.MarkFlagRequired("target-branch"))

	c.PersistentFlags().StringVarP(&rawFormat, "format", "f", "yaml", "output format [yaml,json,flat]")

	return c
}

func writeTarget(basePath, outputFile string) (io.Writer, error) {
	if outputFile == "" {
		return os.Stdout, nil
	}

	var absPath string
	if filepath.IsAbs(outputFile) {
		absPath = outputFile
	} else {
		absPath = filepath.Join(basePath, outputFile)
	}

	if err := os.MkdirAll(filepath.Dir(absPath), allAllowed); err != nil {
		return nil, err
	}
	return os.Create(absPath)
}
