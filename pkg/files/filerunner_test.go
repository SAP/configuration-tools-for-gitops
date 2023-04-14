package files_test

import (
	"fmt"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"

	"github.com/configuration-tools-for-gitops/pkg/files"
	"github.com/configuration-tools-for-gitops/pkg/testfuncs"
)

type scenario struct {
	title          string
	files          map[string][]byte
	includeFilters map[files.FilterJoin][]string
	excludeFilters map[files.FilterJoin][]string
	want           map[string]files.File
	wantErr        error
}

var scenarios = []scenario{
	{
		title: "working simple example",
		files: map[string][]byte{
			"services/A/values/env-specific/.tmpl": nil,
		},
		includeFilters: map[files.FilterJoin][]string{files.AND: {".tmpl"}},
		want: map[string]files.File{
			"${ROOT}/services/A/values/env-specific/.tmpl": {Name: ".tmpl", IsDir: false, Content: []byte{}},
		},
		wantErr: nil,
	},
	{
		title: "working exclude example",
		files: map[string][]byte{
			"services/A/values/env-specific/.tmpl": nil,
		},
		includeFilters: map[files.FilterJoin][]string{files.AND: {".tmpl"}},
		excludeFilters: map[files.FilterJoin][]string{files.AND: {".tmpl"}},
		want:           map[string]files.File{},
		wantErr:        nil,
	},
	{
		title: "working exclude example 2",
		files: map[string][]byte{
			"services/A/values/env-specific/.tmpl":  nil,
			"services/A/values/env-specific/.tmpl2": nil,
		},
		includeFilters: map[files.FilterJoin][]string{files.AND: {".tmpl"}},
		excludeFilters: map[files.FilterJoin][]string{files.AND: {".tmpl2"}},
		want: map[string]files.File{
			"${ROOT}/services/A/values/env-specific/.tmpl": {Name: ".tmpl", IsDir: false, Content: []byte{}},
		},
		wantErr: nil,
	},
	{
		title: "working include or example",
		files: map[string][]byte{
			"services/A/values/env-specific/.tmpl":  nil,
			"services/A/values/env-specific/.tmpl2": nil,
		},
		includeFilters: map[files.FilterJoin][]string{files.OR: {".tmpl", ".tmpl2"}},
		want: map[string]files.File{
			"${ROOT}/services/A/values/env-specific/.tmpl":  {Name: ".tmpl", IsDir: false, Content: []byte{}},
			"${ROOT}/services/A/values/env-specific/.tmpl2": {Name: ".tmpl2", IsDir: false, Content: []byte{}},
		},
		wantErr: nil,
	},
	{
		title: "working exclude or example",
		files: map[string][]byte{
			"services/A/values/env-specific/.tmpl":  nil,
			"services/A/values/env-specific/.tmpl2": nil,
		},
		includeFilters: map[files.FilterJoin][]string{files.OR: {".tmpl", ".tmpl2"}},
		excludeFilters: map[files.FilterJoin][]string{files.OR: {".tmpl", ".tmpl2"}},
		want:           map[string]files.File{},
		wantErr:        nil,
	},
	{
		title: "working complex example",
		files: map[string][]byte{
			"services/A/values/env-specific/.tmpl": nil,
			"services/B/overlays/.tmpl/file1":      nil,
			"services/B/overlays/.tmpl/file2":      nil,
			"services/B/overlays/.tmpl/file3":      nil,
			"folder/name.tmpl":                     nil,
			"folder/not-found":                     nil,
			"folder/other-folder/not-found":        nil,
		},
		includeFilters: map[files.FilterJoin][]string{files.AND: {".tmpl"}},
		want: map[string]files.File{
			"${ROOT}/services/A/values/env-specific/.tmpl": {Name: ".tmpl", IsDir: false, Content: []byte{}},
			"${ROOT}/folder/name.tmpl":                     {Name: "name.tmpl", IsDir: false, Content: []byte{}},
			"${ROOT}/services/B/overlays/.tmpl":            {Name: ".tmpl", IsDir: true, Content: []byte{}},
			"${ROOT}/services/B/overlays/.tmpl/file1":      {Name: "file1", IsDir: false, Content: []byte{}},
			"${ROOT}/services/B/overlays/.tmpl/file2":      {Name: "file2", IsDir: false, Content: []byte{}},
			"${ROOT}/services/B/overlays/.tmpl/file3":      {Name: "file3", IsDir: false, Content: []byte{}},
		},
		wantErr: nil,
	},
	{
		title: "multiple filters",
		files: map[string][]byte{
			"services/A/values/env-specific/.tmpl": nil,
			"services/B/overlays/.tmpl/file1":      []byte(`content file1`),
			"services/B/overlays/.tmpl/file2":      []byte(`content file2`),
			"services/B/overlays/.tmpl/file3": []byte(`multiline
content
`),
			"folder/name.tmpl":              nil,
			"folder/not-found":              nil,
			"folder/other-folder/not-found": nil,
		},
		includeFilters: map[files.FilterJoin][]string{files.AND: {".tmpl", "overlays"}},
		want: map[string]files.File{
			"${ROOT}/services/B/overlays/.tmpl": {
				Name:    ".tmpl",
				IsDir:   true,
				Content: []byte{},
			},
			"${ROOT}/services/B/overlays/.tmpl/file1": {
				Name:    "file1",
				IsDir:   false,
				Content: []byte(`content file1`),
			},
			"${ROOT}/services/B/overlays/.tmpl/file2": {
				Name:    "file2",
				IsDir:   false,
				Content: []byte(`content file2`),
			},
			"${ROOT}/services/B/overlays/.tmpl/file3": {
				Name:  "file3",
				IsDir: false,
				Content: []byte(`multiline
content
`),
			},
		},
		wantErr: nil,
	},
}

func TestReadList(t *testing.T) {
	for _, s := range scenarios {
		t.Logf("test scenario: %s\n", s.title)
		s.Test(t)
	}
}

func (s scenario) Test(t *testing.T) {
	tmpDir, err := prepareTestDirTree(s.files)
	if err != nil {
		t.Logf("unable to create test dir tree: %v\n", err)
		t.FailNow()
	}
	defer os.RemoveAll(tmpDir)

	fileRunner := files.New(tmpDir).
		Include(files.AND, s.includeFilters[files.AND]).
		Include(files.OR, s.includeFilters[files.OR]).
		Exclude(files.AND, s.excludeFilters[files.AND]).
		Exclude(files.OR, s.excludeFilters[files.OR])

	gotRaw, err := fileRunner.Execute()
	testfuncs.CheckErrs(t, s.wantErr, err)
	got := cleanRes(tmpDir, gotRaw.Content())
	wantList := removeContent(s.want)
	if !reflect.DeepEqual(wantList, got) {
		t.Errorf("List results do not match: \nwant = \"%+v\"\ngot =  \"%+v\"", wantList, got)
		t.Fail()
	}

	gotRawRead, err := fileRunner.ReadContent().Execute()
	testfuncs.CheckErrs(t, s.wantErr, err)
	gotRead := cleanRes(tmpDir, gotRawRead.Content())
	// TestRead
	if !reflect.DeepEqual(s.want, gotRead) {
		t.Errorf("Read results do not match: \nwant = \"%+v\"\ngot =  \"%+v\"", s.want, gotRead)
		t.Fail()
	}
}

func cleanRes(basedir string, gotRaw map[string]files.File) map[string]files.File {
	got := make(map[string]files.File, len(gotRaw))
	for name, file := range gotRaw {
		got[strings.Replace(name, basedir, "${ROOT}", 1)] = files.File{
			Name: file.Name, IsDir: file.IsDir, Content: file.Content,
		}
	}
	return got
}

func removeContent(wantRaw map[string]files.File) map[string]files.File {
	want := make(map[string]files.File, len(wantRaw))
	for name, file := range wantRaw {
		want[name] = files.File{Name: file.Name, IsDir: file.IsDir, Content: []byte{}}
	}
	return want
}

func prepareTestDirTree(fs map[string][]byte) (string, error) {
	tmpDir, err := os.MkdirTemp("", "")
	if err != nil {
		return "", fmt.Errorf("error creating temp directory: %v\n", err)
	}

	for name, content := range fs {
		fSlice := strings.Split(name, "/")
		fileName := fSlice[len(fSlice)-1]
		filePath := strings.Join(fSlice[:len(fSlice)-1], "/")

		if err := os.MkdirAll(filepath.Join(tmpDir, filePath), 0755); err != nil {
			os.RemoveAll(tmpDir)
			return "", fmt.Errorf("failed to create dir %s: %s", filePath, err)
		}

		if err := files.Write(
			filepath.Join(tmpDir, filePath, fileName),
			0666, content,
		); err != nil {
			os.RemoveAll(tmpDir)
			return "", err
		}
	}
	return tmpDir, nil
}
