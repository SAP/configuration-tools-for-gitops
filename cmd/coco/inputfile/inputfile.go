package inputfile

import (
	"fmt"

	"github.com/SAP/configuration-tools-for-gitops/v2/pkg/files"
	"github.com/SAP/configuration-tools-for-gitops/v2/pkg/maputils"
	"gopkg.in/yaml.v3"
)

// Coco struct contains keys for both components and environments.
// Unused fields are set nil by the unmarshal function
//
//nolint:lll // no linebreaks available for struct tags
type Coco struct {
	Type         ConfigType `yaml:"type" doc:"msg=type of the configuration file,req,o=environment,o=component"`
	Values       []string   `yaml:"values" doc:"msg=list relative paths to config files, req=for environments only"`
	Name         string     `yaml:"name" doc:"msg=name of component or environment,req"`
	Dependencies []string   `yaml:"dependencies" doc:"msg=list of components that this component depends on, req=for components only"`
}

// Types of config files.
// Needs to be maintained manually together with the corresponding check functions.
type ConfigType string

const (
	COMPONENT   ConfigType = "component"
	ENVIRONMENT ConfigType = "environment"
)

var AllConfigTypes = map[ConfigType]bool{COMPONENT: true, ENVIRONMENT: true}

// Receives a file path and reads the byte content into a Coco struct
// File should be a yaml containing at least a valid type key
func Load(file string) (Coco, error) {
	content, err := files.Read(file)
	if err != nil {
		return Coco{}, err
	}
	res := Coco{}
	err = yaml.Unmarshal(content, &res)
	if err != nil {
		return Coco{}, err
	}

	if _, ok := AllConfigTypes[res.Type]; !ok {
		allConfigTypesOrdered := maputils.KeysSorted(AllConfigTypes)

		return Coco{}, fmt.Errorf(
			"unsupported type: %q, available options: %+v",
			res.Type,
			allConfigTypesOrdered,
		)
	}
	return res, nil
}

func FindAll(
	basepath, configFileName string, includeOr, includeAnd, exclude []string,
) (map[string]files.File, error) {
	includeAnd = append(includeAnd, configFileName)
	fileRunner := files.New(basepath).
		Include(files.OR, includeOr).
		Include(files.AND, includeAnd).
		Exclude(files.OR, exclude).
		ReadContent()
	vFiles, err := fileRunner.Execute()
	if err != nil {
		return nil, err
	}
	return vFiles.Content(), nil
}

func (c *Coco) IsComponent() bool {
	return c.Type == COMPONENT
}

func (c *Coco) IsEnvironment() bool {
	return c.Type == ENVIRONMENT
}
