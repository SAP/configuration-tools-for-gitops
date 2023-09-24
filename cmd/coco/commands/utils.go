package commands

import (
	"os"

	"github.com/SAP/configuration-tools-for-gitops/v2/pkg/log"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
	"github.com/spf13/viper"
)

func bindFlag(fs *pflag.FlagSet, key, flag, env string) {
	cobra.CheckErr(viper.BindPFlag(key, fs.Lookup(flag)))
	cobra.CheckErr(viper.BindEnv(key, env))
}

func failOnError(err error, command string) {
	if err != nil {
		log.Sugar.Errorf("failed in command %q: %v", command, err)
		os.Exit(1)
	}
}
