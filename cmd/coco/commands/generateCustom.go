package commands

import (
	"fmt"

	"github.com/SAP/configuration-tools-for-gitops/v2/cmd/coco/generate"
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
				failOnError(
					fmt.Errorf("no go-template provided, please provide a template as argument"),
					"custom",
				)
			}
			return nil
		},
		Run: func(cmd *cobra.Command, args []string) {
			failOnError(
				generate.ParseTemplate(args[0], customValues, customTarget),
				"custom",
			)
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
