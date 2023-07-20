package generate

import (
	"bytes"
	"fmt"
	"github.com/SAP/configuration-tools-for-gitops/cmd/coco/inputfile"
	"github.com/SAP/configuration-tools-for-gitops/pkg/maputils"
	"github.com/spf13/viper"
	"reflect"
	"sort"
	"testing"

	"github.com/SAP/configuration-tools-for-gitops/pkg/testfuncs"
	"gopkg.in/yaml.v3"
)

type scenarioValueFiles struct {
	title          string
	includeFilters []string
	excludeFilters []string
	files          map[string][]byte
	wantFiles      map[string][]byte
	wantErr        error
}

var (
	allConfigTypes = sortConfigTypes(maputils.Keys(inputfile.AllConfigTypes))
	configFileName = "coco.yaml"
)

var scenariosValueFiles = []scenarioValueFiles{
	{
		title:          "working general example",
		includeFilters: []string{"${BASEPATH}/values/"},
		excludeFilters: []string{".tmpl"},
		files: map[string][]byte{
			"services/A/values/env-specific": nil,
			"folder/name.tmpl":               nil,
			"values/.tmpl/test":              nil,
			"values/name.tmpl":               nil,
			"values/env1/file1.yaml":         []byte(`k1: v1`),
			"values/env1/file2.yaml": []byte(`
k2: v2
k22: v22
`),
			"values2/env2/file3": nil,
			"values2/env2/file4": nil,
			"values/env2/file3.yaml": []byte(`
k3: v3
k33: v33
`),
			"values/env2/file4.yaml": []byte(`k4: v4`),
			"values/env1/coco.yaml": []byte(`
type: environment
name: name1
values: 
  - file1
  - file2
`),
			"values/env2/coco.yaml": []byte(`
type: environment
name: name2
values: 
  - file3
`),
		},
		wantFiles: map[string][]byte{
			"name1": []byte(`
k1: v1
k2: v2
k22: v22
`),
			"name2": []byte(`
k3: v3
k33: v33
`),
		},
		wantErr: nil,
	},
	{
		title:          "working general example",
		includeFilters: []string{"${BASEPATH}/values/", "${BASEPATH}/values2/"},
		excludeFilters: []string{".tmpl"},
		files: map[string][]byte{
			"services/A/values/env-specific": nil,
			"folder/name.tmpl":               nil,
			"values/.tmpl/test":              nil,
			"values/name.tmpl":               nil,
			"values/env1/file1.yaml":         []byte(`k1: v1`),
			"values/env1/file2.yaml": []byte(`
k2: v2
k22: v22
`),
			"values/env2/file3": nil,
			"values/env2/file4": nil,
			"values2/env/file3.yaml": []byte(`
k3: v3
k33: v33
`),
			"values2/env/file4.yaml": []byte(`k4: v4`),
			"values/env1/coco.yaml": []byte(`
type: environment
name: name1
values: 
  - file1
  - file2
`),
			"values2/env/coco.yaml": []byte(`
type: environment
name: name2
values: 
  - file3
`),
		},
		wantFiles: map[string][]byte{
			"name1": []byte(`
k1: v1
k2: v2
k22: v22
`),
			"name2": []byte(`
k3: v3
k33: v33
`),
		},
		wantErr: nil,
	},
	{
		title:          "Unsupported coco type",
		includeFilters: []string{"${BASEPATH}/values/", "${BASEPATH}/values2/"},
		excludeFilters: []string{".tmpl"},
		files: map[string][]byte{
			"values/coco.yaml": []byte(`
type: unsupportedType
`),
		},
		wantFiles: nil,
		wantErr: fmt.Errorf(
			"unsupported type: %q, available options: %+v",
			"unsupportedType",
			allConfigTypes,
		),
	},
}

func TestFindValueFiles(t *testing.T) {
	fmt.Println(viper.GetString("component.cfg"))
	for _, s := range scenariosValueFiles {
		t.Logf("test scenario: %s\n", s.title)
		s.Test(t)
	}
}

func (s *scenarioValueFiles) Test(t *testing.T) {
	td, err := testfuncs.PrepareTestDirTree(s.files)
	if err != nil {
		t.Logf("unable to create test dir tree: %v\n", err)
		t.FailNow()
	}
	defer td.Cleanup(t)
	tmpDir := td.Path()

	got, err := readValueFiles(tmpDir, configFileName, s.includeFilters, []string{}, s.excludeFilters)
	testfuncs.CheckErrs(t, s.wantErr, err)

	s.CheckRes(t, tmpDir, got)
}

func (s *scenarioValueFiles) CheckRes(t *testing.T, basedir string, got map[string]interface{}) {
	var expected map[string]interface{}
	if len(s.wantFiles) > 0 {
		expected = make(map[string]interface{}, len(s.wantFiles))

	}

	for name, rawValues := range s.wantFiles {
		d := yaml.NewDecoder(bytes.NewReader(rawValues))
		var values interface{}
		if err := d.Decode(&values); err != nil {
			t.Fatalf("failed to decode want: %v", err)
		}
		expected[name] = values
	}
	if !reflect.DeepEqual(expected, got) {
		t.Errorf("results do not match: \nwant = \"%+v\"\ngot =  \"%+v\"",
			expected,
			got,
		)
		t.Fail()
	}
}

func sortConfigTypes(a []inputfile.ConfigType) []inputfile.ConfigType {
	sort.Slice(a, func(i, j int) bool {
		return a[i] < a[j]
	})
	return a
}
