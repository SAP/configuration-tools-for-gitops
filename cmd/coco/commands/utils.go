package commands

import (
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
	"github.com/spf13/viper"
)

func bindFlag(fs *pflag.FlagSet, key, flag, env string) {
	cobra.CheckErr(viper.BindPFlag(key, fs.Lookup(flag)))
	cobra.CheckErr(viper.BindEnv(key, env))
}
