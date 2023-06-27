package commands

import (
	"os"

	"github.com/SAP/configuration-tools-for-gitops/cmd/coco/generate"
	"github.com/SAP/configuration-tools-for-gitops/pkg/log"
	"github.com/spf13/cobra"
)

var (
	customTarget string
	customValues []string
)

func newGenerateCustom() *cobra.Command {
	// generateCmd represents the generate command
	var c = &cobra.Command{
		Use:   "custom",
		Short: "custom allows render a custom provided template with custom provided values",

		Args: func(cmd *cobra.Command, args []string) error {
			if len(args) < 1 {
				log.Sugar.Errorf("no go-template provided, please provide a template as argument")
				os.Exit(1)
			}
			return nil
		},
		Run: func(cmd *cobra.Command, args []string) {
			err := generate.ParseTemplate(args[0], customValues, customTarget)
			if err != nil {
				log.Sugar.Errorf("generate failed: %s", err)
				os.Exit(1)
			}
		},
	}

	c.Flags().StringVar(
		&customTarget, "target", "",
		"target file for the custom template result",
	)
	c.Flags().StringSliceVar(
		&customValues, "value", []string{},
		"value files for rendering a custom template",
	)
	c.MarkFlagsRequiredTogether("value", "target")

	return c
}
