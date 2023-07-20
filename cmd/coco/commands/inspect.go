package commands

import (
	"fmt"
	"github.com/SAP/configuration-tools-for-gitops/cmd/coco/inputfile"
	"github.com/SAP/configuration-tools-for-gitops/pkg/yamlfile"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"gopkg.in/yaml.v3"
	"os"
)

func newInspect() *cobra.Command {
	var c = &cobra.Command{
		Use:   "inspect",
		Short: "show the current coco configuration",
		Long:  `Returns the configuration of coco as well as the default configuration options.`,
		Run: func(cmd *cobra.Command, args []string) {
			cfg, err := yaml.Marshal(viper.AllSettings())
			cobra.CheckErr(err)
			fmt.Println(string(cfg))
		},
	}
	c.AddCommand(NewInspectValues())
	return c
}

func NewInspectValues() *cobra.Command {
	return &cobra.Command{
		Use:   "values",
		Short: "options for environment values",
		Long:  `Returns the default configuration options.`,
		Run: func(cmd *cobra.Command, args []string) {
			docsYaml, err := yamlfile.NewFromInterface(yamlfile.DocOutput(inputfile.Coco{}))
			cobra.CheckErr(err)
			cobra.CheckErr(docsYaml.Encode(os.Stdout, 2))
		},
	}
}
