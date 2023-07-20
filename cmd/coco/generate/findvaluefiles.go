package generate

import (
	"bytes"
	"github.com/SAP/configuration-tools-for-gitops/cmd/coco/inputfile"
	"gopkg.in/yaml.v3"
	"path/filepath"
)

func readValueFiles(
	basepath, configFileName string,
	includeOr, includeAnd, exclude []string,
) (map[string]interface{}, error) {
	valueFiles, err := inputfile.FindAll(basepath, configFileName, includeOr, includeAnd, exclude)
	if err != nil {
		return nil, err
	}
	res := make(map[string]interface{}, len(valueFiles))
	var excludeFiles []string
	for path, file := range valueFiles {
		if file.IsDir {
			continue
		}

		var coco inputfile.Coco
		coco, err := inputfile.Load(path)
		if err != nil {
			return nil, err
		}

		if !coco.IsEnvironment() {
			continue
		}

		dir := filepath.Dir(path)
		var valueFilesForEnv []string
		for _, v := range coco.Values {
			valueFilesForEnv = append(valueFilesForEnv, dir+"/"+v+".yaml")
		}

		merged, err := mergeValues(valueFilesForEnv)
		excludeFiles = append(excludeFiles, dir)
		var renderedData bytes.Buffer
		if err = merged.Encode(&renderedData, 2); err != nil {
			return nil, err
		}

		var finalValues interface{}
		finalValuesDecoder := yaml.NewDecoder(bytes.NewReader(renderedData.Bytes()))

		if err = finalValuesDecoder.Decode(&finalValues); err != nil {
			return nil, err
		}

		res[coco.Name] = finalValues
	}
	return res, nil
}
