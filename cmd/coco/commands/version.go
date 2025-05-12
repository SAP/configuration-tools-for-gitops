package commands

import (
	"fmt"

	"github.com/SAP/configuration-tools-for-gitops/v2/pkg/version"
	"github.com/spf13/cobra"
)

func newVersion() *cobra.Command {
	return &cobra.Command{
		Use:   "version",
		Short: "coco version",
		Long: `The version command returns version and additional information about 
the coco binary`,
		Run: func(cmd *cobra.Command, args []string) {
			fmt.Printf("%+v\n", version.ReadAll())
		},
	}
}
