package generate

import (
	"fmt"
	"path/filepath"
	"reflect"
	"strings"
	"testing"

	"github.com/SAP/configuration-tools-for-gitops/pkg/testfuncs"
)

const tmplIdentifier = ".tmpl"

type scenario struct {
	title         string
	files         map[string][]byte
	inclFilters   []string
	exclFilters   []string
	wantTemplates map[string][]template
	wantErr       error
}

var scenarios = []scenario{
	{
		title: "working simple example",
		files: map[string][]byte{
			"services/A/values/env-specific/.tmpl": nil,
			"services/B/overlays/.tmpl/file1":      nil,
			"services/B/overlays/.tmpl/file2":      nil,
			"services/B/overlays/.tmpl/file3":      nil,
			"folder/name.tmpl":                     nil,
			"folder/not-found":                     nil,
			"folder/other-folder/not-found":        nil,
		},
		wantTemplates: map[string][]template{
			"services/A/values/env-specific": {
				{
					source:     "services/A/values/env-specific/.tmpl",
					basepath:   "services/A/values/env-specific",
					namePrefix: "",
					subpath:    "",
				},
			},
			"services/B/overlays": {
				{
					source:     "services/B/overlays/.tmpl/file1",
					basepath:   "services/B/overlays",
					namePrefix: "",
					subpath:    "/file1",
				},
				{
					source:     "services/B/overlays/.tmpl/file2",
					basepath:   "services/B/overlays",
					namePrefix: "",
					subpath:    "/file2",
				},
				{
					source:     "services/B/overlays/.tmpl/file3",
					basepath:   "services/B/overlays",
					namePrefix: "",
					subpath:    "/file3",
				},
			},
			"folder": {
				{
					source:     "folder/name.tmpl",
					basepath:   "folder",
					namePrefix: "name",
					subpath:    "",
				},
			},
		},
		wantErr: nil,
	},
	{
		title: "working nested and filtered example",
		files: map[string][]byte{
			"services/B/overlays/.tmpl/file1":             nil,
			"services/B/overlays/.tmpl/sub/sub/sub/file2": nil,
			"services/B/overlays/.tmpl/sub/sub/file3":     nil,
			"folder/not-found":                            nil,
			"folder/other-folder/not-found":               nil,
			"folder/.tmpl":                                nil,
		},
		inclFilters: []string{"B/overlays"},
		wantTemplates: map[string][]template{
			"services/B/overlays": {
				{
					source:     "services/B/overlays/.tmpl/file1",
					basepath:   "services/B/overlays",
					namePrefix: "",
					subpath:    "/file1",
				},
				{
					source:     "services/B/overlays/.tmpl/sub/sub/file3",
					basepath:   "services/B/overlays",
					namePrefix: "",
					subpath:    "/sub/sub/file3",
				},
				{
					source:     "services/B/overlays/.tmpl/sub/sub/sub/file2",
					basepath:   "services/B/overlays",
					namePrefix: "",
					subpath:    "/sub/sub/sub/file2",
				},
			},
		},
		wantErr: nil,
	},
	{
		title: "working exclude filter",
		files: map[string][]byte{
			"A/.tmpl":       nil,
			"B/.tmpl/file1": nil,
			"tm/name.tmpl":  nil,
		},
		exclFilters: []string{"tm/"},
		wantTemplates: map[string][]template{
			"A": {
				{
					source:     "A/.tmpl",
					basepath:   "A",
					namePrefix: "",
					subpath:    "",
				},
			},
			"B": {
				{
					source:     "B/.tmpl/file1",
					basepath:   "B",
					namePrefix: "",
					subpath:    "/file1",
				},
			},
		},
		wantErr: nil,
	},
}

func TestFindTemplates(t *testing.T) {
	for _, s := range scenarios {
		t.Logf("test scenario: %s\n", s.title)
		s.Test(t)
	}
}

func (s *scenario) Test(t *testing.T) {
	td, err := testfuncs.PrepareTestDirTree(s.files)
	if err != nil {
		t.Logf("unable to create test dir tree: %v\n", err)
		t.FailNow()
	}
	defer td.Cleanup(t)
	tmpDir := td.Path()

	t.Logf("temporary directory for test: %s", tmpDir)

	got, err := findTemplates(tmpDir, tmplIdentifier, s.inclFilters, s.exclFilters)
	testfuncs.CheckErrs(t, s.wantErr, err)

	s.CheckRes(t, tmpDir, got)
}

func (s *scenario) CheckRes(t *testing.T, basedir string, got map[string][]template) {
	expected := make(map[string][]template, len(s.wantTemplates))
	for name, tmpls := range s.wantTemplates {
		expectedTmpls := make([]template, 0, len(tmpls))
		for _, t := range tmpls {
			expectedTmpls = append(expectedTmpls, template{
				source:     filepath.Join("${BASEPATH}", t.source),
				basepath:   filepath.Join("${BASEPATH}", t.basepath),
				namePrefix: t.namePrefix,
				subpath:    t.subpath,
			})
		}
		expected[fmt.Sprintf("${BASEPATH}/%s", name)] = expectedTmpls
	}
	gotClean := make(map[string][]template, len(got))
	for name, tmpls := range got {
		gotTmplClean := make([]template, 0, len(tmpls))
		for _, t := range tmpls {
			gotTmplClean = append(gotTmplClean, template{
				source:     strings.Replace(t.source, basedir, "${BASEPATH}", 1),
				basepath:   strings.Replace(t.basepath, basedir, "${BASEPATH}", 1),
				namePrefix: t.namePrefix,
				subpath:    t.subpath,
			})
		}
		gotClean[strings.Replace(name, basedir, "${BASEPATH}", 1)] = gotTmplClean
	}
	if !reflect.DeepEqual(expected, gotClean) {
		t.Errorf(
			"results do not match: \nwant = \"%+v\"\ngot =  \"%+v\"",
			expected,
			gotClean,
		)
		t.Fail()
	}
}
