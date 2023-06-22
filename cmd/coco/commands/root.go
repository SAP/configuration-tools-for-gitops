package commands

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/SAP/configuration-tools-for-gitops/pkg/git"
	"github.com/SAP/configuration-tools-for-gitops/pkg/log"
	"github.com/SAP/configuration-tools-for-gitops/pkg/version"
	"github.com/spf13/cobra"
	"go.uber.org/zap"

	"github.com/spf13/viper"
)

var (
	componentCfgFile, cfgFile, gp, gu string
	remote, defaultBranch             string
	mD                                int
	logLvl                            log.Level
)

const (
	gitPathKey   = "git.path"
	gitURLKey    = "git.URL"
	gitRemoteKey = "git.remote"
	gitDepthKey  = "git.depth"
	componentCfg = "component.cfg"
	// cannot restrict checkout depth due to upstream bug
	// (see https://github.com/go-git/go-git/issues/328 for issue tracking)
	overWriteGitDepth = 0
)

func Execute() {
	cmd := newRoot()
	cobra.OnInitialize(initConfig)
	cobra.CheckErr(cmd.Execute())
}

func newRoot() *cobra.Command {
	var c = &cobra.Command{
		Use:   "coco",
		Short: "CLI to interact with the gitops repository",
		Long: `coco is a CLI to interact with a gitops repository and shall provide
	various solutions, ranging from file-generation over the calculation of
	dependency trees to various interactions with git and github.`,
		// Uncomment the following line if your bare application
		// has an action associated with it:
		// Run: func(cmd *cobra.Command, args []string) { },
	}

	c.AddCommand(newVersion())
	c.AddCommand(newDependencies())
	c.AddCommand(newGenerate())
	c.AddCommand(newInspect())
	c.AddCommand(newReconcile())

	c.PersistentFlags().StringVar(
		&cfgFile, "config", "", "config file (default $HOME/.coco)",
	)
	bindFlag(c.PersistentFlags(), "cfgfile", "config", "COCO_CFG")

	c.PersistentFlags().StringVarP(
		&componentCfgFile, "component-cfg", "c", "coco.yaml",
		`name of the component-specific configuration file`,
	)
	bindFlag(c.PersistentFlags(), componentCfg, "component-cfg", "COCO_COMPONENT_CFG")

	c.PersistentFlags().VarP(
		&logLvl, "loglvl", "l",
		fmt.Sprintf("sets the log level of the application - key or value of %v", logLvl.AllLevels()),
	)
	bindFlag(c.PersistentFlags(), "loglvl", "loglvl", "LOGLEVEL")

	c.PersistentFlags().StringVarP(
		&gp, "git-path", "p", "", "path where the configuration repository locally resides",
	)
	bindFlag(c.PersistentFlags(), gitPathKey, "git-path", "GIT_PATH")

	c.PersistentFlags().StringVarP(
		&gu, "git-url", "u", "", "git URL of the configuration repository",
	)
	bindFlag(c.PersistentFlags(), gitURLKey, "git-url", "GIT_URL")

	c.PersistentFlags().StringVarP(
		&remote, "git-remote", "r", "origin",
		"remote branch to compare against for changed components",
	)
	bindFlag(c.PersistentFlags(), gitRemoteKey, "git-remote", "GIT_REMOTE")

	c.PersistentFlags().StringVarP(
		&defaultBranch, "git-defaultbranch", "b", "main", "default branch",
	)
	bindFlag(c.PersistentFlags(), "git.defaultBranch", "git-defaultbranch", "GIT_DEFAULT_BRANCH")

	c.PersistentFlags().IntVar(
		&mD, "git-depth", 0,
		`[NOT IN USE (upstream bug: see https://github.com/go-git/go-git/issues/328 for issue tracking)]
	max checkout depth of the git repository`,
	)
	bindFlag(c.PersistentFlags(), gitDepthKey, "git-depth", "GIT_DEPTH")

	cobra.CheckErr(viper.BindEnv("git-token", "GITHUB_TOKEN"))

	return c
}

// initConfig reads in config file and ENV variables if set.
func initConfig() {
	if err := log.Init(logLvl, "2006-01-02T15:04:05Z07:00", true); err != nil {
		zap.S().Fatal(err)
	}

	v := version.ReadAll()
	log.Sugar.Debugf("build details: %+v", v)

	viper.SetConfigType("yaml")
	if cfgFile != "" {
		viper.SetConfigFile(cfgFile)
	} else {
		home, err := os.UserHomeDir()
		if err != nil {
			zap.S().Fatal(err)
		}
		viper.AddConfigPath(home)
		viper.SetConfigName(".coco")
		viper.SetDefault("cfgfile", filepath.Join(home, ".coco"))
	}

	viper.AutomaticEnv() // read in environment variables that match

	// If a config file is found, read it in.
	if err := viper.ReadInConfig(); err == nil {
		log.Sugar.Debugf("Using config file: %s", viper.ConfigFileUsed())
	}

	if viper.GetString(gitPathKey) == "" {
		path, err := os.Getwd()
		if err != nil {
			zap.S().Fatal(err)
		}
		log.Sugar.Debug("git.path not set - using default ", zap.String("git.path", path))
		viper.Set(gitPathKey, path)
	}

	if viper.GetInt(gitDepthKey) != 0 {
		log.Sugar.Warnf(
			"%s cannot be used at the moment due to upstream bug (https://github.com/go-git/go-git/issues/328)",
			gitDepthKey,
		)
	}

	if ok := consistentGitSetup(
		viper.GetString(gitPathKey),
		viper.GetString(gitURLKey),
		viper.GetString("git-token"),
		viper.GetString(gitRemoteKey),
		viper.GetString("git.defaultBranch"),
		overWriteGitDepth, // viper.GetInt(gitDepth),
		logLvl,
	); !ok {
		os.Exit(1)
	}

	if logLvl.Is(log.Debug()) {
		zf := []interface{}{"all settings: "}
		for k, v := range viper.AllSettings() {
			if k == "git-token" {
				zf = append(zf, zap.String(k, "<is-set>"))
			} else {
				zf = append(zf, zap.String(k, fmt.Sprintf("%+v", v)))
			}
		}
		log.Log.Sugar().Debug(zf...)
	}
}

func consistentGitSetup(
	path, url, token, remote, defaultBranch string, gitDepth int, logLvl log.Level,
) bool {
	if remote == "" {
		return true
	}
	if remote == "origin" && url == "" {
		log.Sugar.Debug(
			"git.remote and git.URL appear to be default: ",
			"functionality relying on external git calls may not work.",
		)
		return true
	}

	if url == "" {
		log.Sugar.Errorf(
			"\"git.remote\" is set but \"git.URL\" is missing.\n%s",
			provideBy(gitURLKey, "--git-url", "git.URL", "GIT_URL"),
		)
		return false
	}
	if _, err := git.New(
		path, url, token, remote, defaultBranch, gitDepth, logLvl,
	); err != nil {
		log.Sugar.Errorf(
			"failed to validate repository in path \"%s\": %s", path, err,
		)
		return false
	}
	return true
}

func provideBy(name, par, config, env string) string {
	return fmt.Sprintf(
		"    Provide \"%s\" in either of the following ways:"+
			"\n      - CLI parameter        (%s)"+
			"\n      - coco config         (%s)"+
			"\n      - environment variable (%s)",
		name, par, config, env,
	)
}
