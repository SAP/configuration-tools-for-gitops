package generate

import (
	"bytes"
	"reflect"
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
			"values/file1.yaml":              []byte(`k1: v1`),
			"values/file2": []byte(`
k2: v2
k22: v22
`),
			"values2/file3": nil,
			"values2/file4": nil,
			"values3/file3": nil,
			"values3/file4": nil,
		},
		wantFiles: map[string][]byte{
			"file1": []byte(`k1: v1`),
			"file2": []byte(`
k2: v2
k22: v22
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
			"values/file1":                   []byte(`k1: v1`),
			"values/file2": []byte(`
k2: v2
k22: v22
`),
			"values2/file3": []byte(`k3: v3`),
			"values2/file4": []byte(`k4: v4`),
			"values3/file3": nil,
			"values3/file4": nil,
		},
		wantFiles: map[string][]byte{
			"file1": []byte(`k1: v1`),
			"file2": []byte(`
k2: v2
k22: v22
`),
			"file3": []byte(`k3: v3`),
			"file4": []byte(`k4: v4`),
		},
		wantErr: nil,
	},
}

func TestFindValueFiles(t *testing.T) {
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

	got, err := readValueFiles(tmpDir, s.includeFilters, []string{}, s.excludeFilters)
	testfuncs.CheckErrs(t, s.wantErr, err)

	s.CheckRes(t, tmpDir, got)
}

func (s *scenarioValueFiles) CheckRes(t *testing.T, basedir string, got map[string]interface{}) {
	expected := make(map[string]interface{}, len(s.wantFiles))

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
