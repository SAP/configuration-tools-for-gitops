package generate

import (
	"path/filepath"

	"github.com/SAP/configuration-tools-for-gitops/cmd/coco/inputfile"
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
	for path, file := range valueFiles {
		if file.IsDir {
			continue
		}

		coco, err := inputfile.Load(path)
		if err != nil {
			return nil, err
		}

		if !coco.IsEnvironment() {
			continue
		}

		dir := filepath.Dir(path)
		valueFilesForEnv := make([]string, 0, len(coco.Values))
		for _, v := range coco.Values {
			valueFilesForEnv = append(valueFilesForEnv, filepath.Join(dir, v))
		}

		merged, err := mergeValues(valueFilesForEnv)
		if err != nil {
			return nil, err
		}

		var finalValues interface{}
		err = merged.Decode(&finalValues)
		if err != nil {
			return nil, err
		}
		res[coco.Name] = finalValues

		res[coco.Name] = finalValues
	}
	return res, nil
}
