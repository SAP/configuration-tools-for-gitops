package commands

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/SAP/configuration-tools-for-gitops/cmd/coco/generate"
	"github.com/SAP/configuration-tools-for-gitops/pkg/log"
	"github.com/SAP/configuration-tools-for-gitops/pkg/version"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"go.uber.org/zap"
)

var (
	environmentFilter []string
	valuesFolders     []string
	excludeFolders    []string
	tmplIdentifier    string
	persistenceFlag   string
	takeControl       bool
)

// generateCmd represents the generate command
var generateCmd = &cobra.Command{
	Use:     "generate",
	Aliases: []string{"gen"},
	Short:   "generate allows to run file-generation over the gitops repository",
	Long: `
The generate command governs all aspects of non-sensitive file-generation
in the gitops repository.
`,
	Args: func(cmd *cobra.Command, args []string) error {
		if len(args) < 1 {
			log.Sugar.Info("no service specified - generate globally")
			return nil
		}
		return nil
	},
	Run: func(cmd *cobra.Command, args []string) {

		basepath := viper.GetString(gitPath)
		err := generate.Generate(
			basepath,
			tmplIdentifier,
			persistenceFlag,
			version.ReadAll(),
			cleanValuePaths(valuesFolders, basepath),
			environmentFilter,
			args,
			excludeFolders,
			logLvl,
			takeControl,
		)
		if err != nil {
			log.Sugar.Errorf("generate failed: %s", err)
			os.Exit(1)
		}
	},
}

//nolint:gochecknoinits // required by the cobra framework
func init() {
	rootCmd.AddCommand(generateCmd)

	generateCmd.PersistentFlags().StringSliceVarP(
		&environmentFilter,
		"env-filter", "e",
		[]string{},
		"restrict the command to one or more environments",
	)

	generateCmd.Flags().StringSliceVarP(
		&valuesFolders,
		"values", "v",
		[]string{"values"},
		"folder that contains all value files used for rendering templates",
	)
	generateCmd.Flags().StringSliceVarP(
		&excludeFolders,
		"exclude", "x",
		[]string{},
		"folders that will be excluded from file generation",
	)
	generateCmd.Flags().StringVarP(
		&tmplIdentifier,
		"templates", "t",
		".tmpl",
		"pattern in folder or file names that identifies templates for rendering",
	)
	generateCmd.Flags().StringVar(
		&persistenceFlag,
		"keep-lines",
		"HumanInput",
		`the value of this parameter governs which lines in generated files will not
be overwritten by coco. Per default, all lines with the comment "# HumanInput"
or the yaml tag "!HumanInput" will not be overwritten.`,
	)
	generateCmd.Flags().BoolVar(
		&takeControl,
		"take-control",
		false,
		`if this flag is set, coco takes control over all generated files regardless 
of the version in the generated files`,
	)
	if err := generateCmd.Flags().MarkDeprecated("take-control", "please use \"--force\" instead"); err != nil {
		zap.L().Sugar().Errorf("flag %q not found: %v", "take-control", err)
		os.Exit(1)
	}
	generateCmd.Flags().BoolVar(
		&takeControl,
		"force",
		false,
		`if this flag is set, coco forcefully regenerats all files regardless of 
the version in the generated files`,
	)
}

func cleanValuePaths(valuesFolders []string, basepath string) []string {
	res := make([]string, 0, len(valuesFolders))
	for _, f := range valuesFolders {
		if filepath.IsAbs(f) {
			res = append(res, f)
			continue
		}
		res = append(res,
			fmt.Sprintf("%s%s", filepath.Join(basepath, f), string(os.PathSeparator)),
		)
	}
	return res
}
