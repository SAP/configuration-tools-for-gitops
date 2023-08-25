package files

import (
	"errors"
	"io/fs"
	"os"
	"path/filepath"
	"reflect"
	"testing"

	"github.com/SAP/configuration-tools-for-gitops/v2/pkg/testfuncs"
)

type scenarioFiles struct {
	title  string
	input  input
	output output
}

type input struct {
	path              string
	permissions       fs.FileMode
	content           []byte
	additionalContent []byte
}

type output struct {
	content []byte
	err     error
}

var scenariosFiles = []scenarioFiles{
	{
		title: "working simple example",
		input: input{
			path:        "file",
			permissions: 0777,
			content: []byte(`hello >

`),
			additionalContent: []byte(`world`),
		},
		output: output{
			content: []byte(`hello >

world`),
			err: nil,
		},
	},
	{
		title: "fail",
		input: input{
			path:              "file",
			permissions:       0777,
			content:           []byte(``),
			additionalContent: []byte(``),
		},
		output: output{
			content: []byte(``),
			err:     errors.New("fail in createOpen"),
		},
	},
}

func TestReadWrite(t *testing.T) {
	for _, s := range scenariosFiles {
		t.Logf("test scenario: %s\n", s.title)
		s.Test(t)
	}
}

func (s *scenarioFiles) Test(t *testing.T) {
	tmpDir, err := testfuncs.PrepareTestDirTree(map[string][]byte{})
	if err != nil {
		t.Logf("unable to create test dir tree: %v\n", err)
		t.FailNow()
	}
	defer cleanup(t, tmpDir)

	p := filepath.Join(tmpDir.Path(), s.input.path)

	t.Log("check Write")
	if s.output.err != nil {
		createOpen = fail{s.output.err}.co
	}
	err = Write(p, s.input.permissions, s.input.content)
	testfuncs.CheckErrs(t, s.output.err, err)
	content, err := Read(p)
	testfuncs.CheckErrs(t, nil, err)
	s.checkRes(t, "Write", s.input.content, content)

	t.Log("check Open")
	f, err := Open(p)
	testfuncs.CheckErrs(t, s.output.err, err)
	if f != nil {
		_, err = f.Write(s.input.additionalContent)
		testfuncs.CheckErrs(t, nil, err)
		content, err = Read(p)
		testfuncs.CheckErrs(t, nil, err)
		s.checkRes(t, "Open", s.output.content, content)
	}

	t.Log("check Write for erasing content")
	err = Write(p, s.input.permissions, []byte{})
	testfuncs.CheckErrs(t, s.output.err, err)
	content, err = Read(p)
	testfuncs.CheckErrs(t, nil, err)
	s.checkRes(t, "Write for erasing content", []byte{}, content)

	t.Log("check WriteOpen")
	f, err = WriteOpen(p, s.input.permissions, s.input.content)
	testfuncs.CheckErrs(t, s.output.err, err)
	if f != nil {
		_, err = f.Write(s.input.additionalContent)
		testfuncs.CheckErrs(t, nil, err)
		content, err = Read(p)
		testfuncs.CheckErrs(t, nil, err)
		s.checkRes(t, "WriteOpen", s.output.content, content)
	}
}

func (s *scenarioFiles) checkRes(t *testing.T, function string, want, got []byte) {
	if !reflect.DeepEqual(got, want) {
		t.Errorf("Test %s: file content does not match: \nwant = \"%+v\"\ngot =  \"%+v\"",
			function, string(want), string(got))
		t.Fail()
	}
}

type fail struct {
	err error
}

func (f fail) co(path string, flag int, permissions fs.FileMode) (*os.File, error) {
	return nil, f.err
}

func cleanup(t *testing.T, tmpDir testfuncs.TestDir) {
	createOpen = co
	tmpDir.Cleanup(t)
}
