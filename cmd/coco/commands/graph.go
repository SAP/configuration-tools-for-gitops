package commands

import (
	"os"

	"github.com/SAP/configuration-tools-for-gitops/cmd/coco/dependencies"
	"github.com/SAP/configuration-tools-for-gitops/cmd/coco/graph"
	"github.com/SAP/configuration-tools-for-gitops/pkg/log"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

func newGraph() *cobra.Command {
	return &cobra.Command{
		Use:     "graph",
		Aliases: []string{"g"},
		Short:   "Returns the downstream dependencies for all components in the repository",
		Long: `This command constructs all components that depend on any given component.
	This is done by constructing the full dependency graph of upstream components
	(defined implicitly via the direct upstream dependencies given in the 
	dependencies.yaml file of each component) and then inverting this graph.
	The output gives per component the weighted list of downstream dependencies,
	where the weight corresponds to the number of connections to reach the downstream
	from the upstream component.
	`,
		PreRun: func(cmd *cobra.Command, args []string) {
			var ok bool
			format, ok = graph.CastOutputFormat(rawFormat)
			if !ok {
				log.Sugar.Errorf("illegal format %q", rawFormat)
				os.Exit(1)
			}
		},
		Run: func(cmd *cobra.Command, args []string) {
			deps, _, err := dependencies.Graph(
				viper.GetString(gitPath),
				viper.GetString(componentCfg),
			)
			if err != nil {
				log.Sugar.Errorf("graph failed with: %s", err)
				os.Exit(1)
			}
			deps.Print(os.Stdout, format)
		},
	}
}
