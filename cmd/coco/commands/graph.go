package commands

import (
	"fmt"
	"os"

	"github.com/SAP/configuration-tools-for-gitops/v2/cmd/coco/dependencies"
	"github.com/SAP/configuration-tools-for-gitops/v2/cmd/coco/graph"
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
				failOnError(fmt.Errorf("illegal format %q", rawFormat), "graph")
			}
		},
		Run: func(cmd *cobra.Command, args []string) {
			deps, _, err := dependencies.Graph(
				viper.GetString(gitPathKey),
				viper.GetString(componentCfg),
			)
			failOnError(err, "graph")
			deps.Print(os.Stdout, format)
		},
	}
}
