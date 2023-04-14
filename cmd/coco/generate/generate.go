package generate

import (
	"fmt"

	"github.com/configuration-tools-for-gitops/pkg/log"
	"github.com/configuration-tools-for-gitops/pkg/version"
)

var (
	renderer func(
		string, []template, map[string]interface{},
		chan<- renderReport, log.Level, string, *version.Version, bool,
	) = render
)

// Generate is the main entry function for the generate package which governs
// file generation from templates as described in the ./readme.md.
// Inputs:
//   - basepath: root from where template files will be identified in all sub folders
//   - templateIdentifier: substring in a filename that identifies this file as a template
//   - persistenceFlag: identifier in generated yaml files that this line shall not be overwritten
//   - clusterValues: folder in which value files for file generation are located
//   - envFilters: filters down the list of environments for which file generation is performed
//   - folderFilters: filters down the list of template locations to specific sub-folders
//   - version: coco version (for comparisons with the version in the existing generated files)
//   - takeControl: overwrite to do file generation also on files that have a different version
//   - logLvl: specifies the log level that will be used
func Generate(
	basepath, templateIdentifier, persistenceFlag string,
	v *version.Version,
	clusterValues, envFilters, folderFilters, excludeFolders []string,
	logLvl log.Level, takeControl bool,
) error {
	tmpls, err := findTemplates(basepath, templateIdentifier, folderFilters, excludeFolders)
	if err != nil {
		return err
	}

	vals, err := readValueFiles(
		basepath,
		clusterValues,
		envFilters,
		[]string{templateIdentifier},
	)
	if err != nil {
		return err
	}

	reports := make(chan renderReport, len(tmpls))

	// All template folders (or files) that have been found are rendered concurrently.
	// Each concurrent process renders the template(s) for all specified environments
	// (from the value files).
	for name, tmpl := range tmpls {
		go renderer(name, tmpl, vals, reports, logLvl, persistenceFlag, v, takeControl)
	}
	return reportResults(reports)
}

// renderReport holds the aggregated result report of a render function call
// If non-nil, it contains either warnings or error messages.
type renderReport struct {
	items []logItem
}

type logItem struct {
	Msg string
	// Level: Debug(-1), Info(0), Warn(1), Error(2), DPanic(3), Panic(4), Fatal(5)
	Level   log.Level
	Context log.Context
}

// reportResults waits for the reports of all concurrent render function calls
// and evaluates them. All results are sent to the logger and if the log level
// is at Error level (2) or higher the reporter returns an error to the caller.
func reportResults(reports chan renderReport) error {
	foundReports := []renderReport{}

	for i := 0; i < cap(reports); i++ {
		r := <-reports
		if len(r.items) > 0 {
			foundReports = append(foundReports, r)
		}
	}
	close(reports)

	if len(foundReports) > 0 {
		errorsFound := false
		for _, r := range foundReports {
			for _, i := range r.items {
				i.Context.Log(i.Msg, i.Level)
				if i.Level.AsInt() >= 2 {
					errorsFound = true
				}
			}
		}
		if errorsFound {
			return fmt.Errorf("%d rendering errors encountered", len(foundReports))
		}
	}
	return nil
}
