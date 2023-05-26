package commands

import (
	"fmt"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"gopkg.in/yaml.v3"
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
	return c
}
