package generate

import (
	"bytes"
	"path/filepath"
	"strings"

	"github.com/SAP/configuration-tools-for-gitops/pkg/files"
	"gopkg.in/yaml.v3"
)

func readValueFiles(
	basepath string,
	includeOr, includeAnd, exclude []string,
) (map[string]interface{}, error) {
	fileRunner := files.New(basepath).
		Include(files.OR, includeOr).
		Include(files.AND, includeAnd).
		Exclude(files.OR, exclude).
		ReadContent()

	if len(includeAnd) > 0 {
		fileRunner = fileRunner.Include(files.AND, includeAnd)
	}
	vFiles, err := fileRunner.Execute()
	if err != nil {
		return nil, err
	}
	valueFiles := vFiles.Content()
	res := make(map[string]interface{}, len(valueFiles))
	for path, file := range valueFiles {
		if file.IsDir {
			continue
		}
		d := yaml.NewDecoder(bytes.NewReader(file.Content))
		var values interface{}
		err := d.Decode(&values)
		if err != nil {
			return nil, err
		}
		res[strings.Replace(filepath.Base(path), ".yaml", "", 1)] = values
	}
	return res, nil
}
