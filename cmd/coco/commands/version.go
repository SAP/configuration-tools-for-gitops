package commands

import (
	"fmt"

	"github.com/SAP/configuration-tools-for-gitops/pkg/version"
	"github.com/spf13/cobra"
)

// generateCmd represents the generate command
var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "coco version",
	Long: `The version command returns version and additional information about 
the coco binary`,
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Printf("%+v\n", version.ReadAll())
	},
}

//nolint:gochecknoinits // required by the cobra framework
func init() {
	rootCmd.AddCommand(versionCmd)
}
