package dependencies

import (
	"path/filepath"

	g "github.com/SAP/configuration-tools-for-gitops/cmd/coco/graph"
	"github.com/SAP/configuration-tools-for-gitops/pkg/files"
	"github.com/SAP/configuration-tools-for-gitops/pkg/log"
	"gopkg.in/yaml.v3"
)

var (
	unmarshal    func([]byte, interface{}) error            = yaml.Unmarshal
	dependencies func(string, string) (*files.Files, error) = deps
)

func Graph(path, depFileName string) (
	graph g.ComponentDependencies, components map[string]string, err error) {
	c := log.Context{"path": path, "dependency-file": depFileName}
	allDeps, components, err := constructGraph(path, depFileName)
	if logErr(c, err) {
		return g.ComponentDependencies{}, nil, err
	}
	return g.GenerateUpToDown(allDeps), components, nil
}

func logErr(c log.Context, err error) bool {
	if err != nil {
		c.NewError(err, log.Error()).Log(1)
		return true
	}
	return false
}

func constructGraph(path, depFileName string) (
	downToUp g.DownToUp, componentPaths map[string]string, err error,
) {
	depFiles, err := dependencies(path, depFileName)
	if err != nil {
		return
	}
	fs := depFiles.Content()

	downToUp = make(g.DownToUp, len(fs))
	componentPaths = make(map[string]string, len(fs))

	for p, f := range fs {
		if f.IsDir {
			continue
		}

		var df depFile
		if err := unmarshal(f.Content, &df); err != nil {
			return downToUp, componentPaths, err
		}

		depMap := make(map[string]bool, len(df.Dependencies))
		for _, d := range df.Dependencies {
			depMap[d] = true
		}
		downToUp[df.Name] = depMap

		relPath := relativeComponentPath(p, path)
		if df.Name != filepath.Base(relPath) {
			log.Sugar.Warnf(
				"component name \"%s\" and its folder name \"%s\" differ",
				df.Name, relPath,
			)
		}
		componentPaths[df.Name] = relPath
	}
	return downToUp, componentPaths, nil
}

func relativeComponentPath(p, path string) string {
	cleanedPath := filepath.Dir(p[len(path):])
	if filepath.IsAbs(cleanedPath) {
		return cleanedPath[1:]
	}
	return cleanedPath
}

type depFile struct {
	Name         string   `yaml:"name"`
	Dependencies []string `yaml:"dependencies"`
}

func deps(path, depFileName string) (*files.Files, error) {
	return files.New(path).
		Include(files.AND, []string{depFileName}).
		ReadContent().
		Execute()
}
