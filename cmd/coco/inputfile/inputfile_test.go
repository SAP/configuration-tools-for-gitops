package inputfile

import (
	"fmt"
	"reflect"
	"strings"
	"testing"

	"github.com/SAP/configuration-tools-for-gitops/pkg/log"
	"github.com/SAP/configuration-tools-for-gitops/pkg/maputils"
	"github.com/SAP/configuration-tools-for-gitops/pkg/testfuncs"
	"go.uber.org/zap"
)

type inputfile struct {
	title   string
	input   map[string][]byte
	want    []Coco
	wantErr error
}

type inputFindAll struct {
	title            string
	configFileName   string
	input            map[string][]byte
	includeOrFilters []string
	includeAndFilter []string
	excludeFilter    []string
	want             map[string][]byte
	wantErr          error
}

var inputfiles = []inputfile{
	{
		title: "General working example for environment",
		input: map[string][]byte{
			"coco": []byte(`
type: environment
name: name1
values:
  - file1
  - file2
`),
		},
		want: []Coco{
			{
				Type:   ENVIRONMENT,
				Name:   "name1",
				Values: []string{"file1", "file2"},
			},
		},
		wantErr: nil,
	},
	{
		title: "General working example for component",
		input: map[string][]byte{
			"coco": []byte(`
type: component
name: component1
dependencies:
  - dep1
  - dep2
`),
		},
		want: []Coco{
			{
				Type:         COMPONENT,
				Name:         "component1",
				Dependencies: []string{"dep1", "dep2"},
			},
		},
		wantErr: nil,
	},
	{
		title: "Unsupported type",
		input: map[string][]byte{
			"coco.yaml": []byte(`
type: unsupportedType
`),
		},
		want: []Coco{{}},
		wantErr: fmt.Errorf(
			"unsupported type: %q, available options: %+v",
			"unsupportedType",
			maputils.KeysSorted(AllConfigTypes),
		),
	},
}

var inputsFindAll = []inputFindAll{
	{
		title:          "general working examples with different config types and default configFileName",
		configFileName: "coco.yaml",
		input: map[string][]byte{
			"values/env1/coco.yaml": []byte(`
type: environment
name: env1
values: 
  - v1.yaml
  - v11.yaml`),
			"values/env2/coco.yaml": []byte(`
type: environment
name: env2
values: 
  - v2.yaml
  - v22.yaml`),
			"dependencies/coco.yaml": []byte(`
type: component
name: comp1
dependencies:
  - dep1
`),
		},
		includeOrFilters: nil,
		includeAndFilter: nil,
		excludeFilter:    nil,
		want: map[string][]byte{
			"/values/env1/coco.yaml": []byte(`
type: environment
name: env1
values: 
  - v1.yaml
  - v11.yaml`),
			"/values/env2/coco.yaml": []byte(`
type: environment
name: env2
values: 
  - v2.yaml
  - v22.yaml`),
			"/dependencies/coco.yaml": []byte(`
type: component
name: comp1
dependencies:
  - dep1
`),
		},
		wantErr: nil,
	},
	{
		title:          "general working example including only specified configFileNames",
		configFileName: "custom.yaml",
		input: map[string][]byte{
			"values/env1/coco.yaml": []byte(`
type: environment
name: env1
values: 
  - v1.yaml
  - v11.yaml`),
			"values/env2/custom.yaml": []byte(`
type: environment
name: env2
values: 
  - v2.yaml
  - v22.yaml`),
			"dependencies/coco.yaml": []byte(`
type: component
name: comp1
dependencies:
  - dep1
`),
		},
		includeOrFilters: nil,
		includeAndFilter: nil,
		excludeFilter:    nil,
		want: map[string][]byte{
			"/values/env2/custom.yaml": []byte(`
type: environment
name: env2
values: 
  - v2.yaml
  - v22.yaml`),
		},
		wantErr: nil,
	},
}

func TestInputfile(t *testing.T) {
	if err := log.Init(log.Debug(), "", true); err != nil {
		zap.S().Fatal(err)
	}
	for _, i := range inputfiles {
		t.Logf("test scenario: %s\n", i.title)
		i.Test(t)
	}
	for _, i := range inputsFindAll {
		t.Logf("test scenario: %s\n", i.title)
		i.FindTest(t)
	}
}

func (i *inputFindAll) FindTest(t *testing.T) {
	td, err := testfuncs.PrepareTestDirTree(i.input)
	if err != nil {
		t.Logf("unable to create test dir tree: %v\n", err)
		t.FailNow()
	}
	defer td.Cleanup(t)
	tmpDir := td.Path()
	files, err := FindAll(tmpDir, i.configFileName, i.includeOrFilters, i.includeAndFilter, i.excludeFilter)
	if err != nil {
		t.Error(err)
		t.FailNow()
	}
	var res = map[string][]byte{}
	for p, v := range files {
		res[strings.Replace(p, tmpDir, "", 1)] = v.Content
	}
	if !reflect.DeepEqual(res, i.want) {
		t.Errorf("results do not match: \nwant = \"%+v\"\ngot =  \"%+v\"",
			i.want,
			res,
		)
		t.Fail()
	}
}

func (i inputfile) Test(t *testing.T) {
	td, err := testfuncs.PrepareTestDirTree(i.input)
	if err != nil {
		t.Logf("unable to create test dir tree: %v\n", err)
		t.FailNow()
	}
	defer td.Cleanup(t)
	tmpDir := td.Path()
	var cocoStructs []Coco
	for filename := range i.input {
		coco, err := Load(tmpDir + "/" + filename)
		testfuncs.CheckErrs(t, i.wantErr, err)
		cocoStructs = append(cocoStructs, coco)
	}
	if !reflect.DeepEqual(cocoStructs, i.want) {
		t.Errorf("results do not match: \nwant = \"%+v\"\ngot =  \"%+v\"",
			i.want,
			cocoStructs,
		)
		t.Fail()
	}
}
