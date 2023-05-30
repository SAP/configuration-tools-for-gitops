package dependencies

import (
	"fmt"
	"os"
	"reflect"
	"testing"

	"github.com/SAP/configuration-tools-for-gitops/cmd/coco/graph"
	"github.com/SAP/configuration-tools-for-gitops/pkg/files"
	"github.com/SAP/configuration-tools-for-gitops/pkg/log"
	"github.com/SAP/configuration-tools-for-gitops/pkg/testfuncs"
	"go.uber.org/zap"
	"gopkg.in/yaml.v3"
)

type graphTest struct {
	title string
	input input
	want  want
}

type input struct {
	depFileName string
	files       map[string][]byte
	mock        mockGraph
}

type want struct {
	res graph.ComponentDependencies
	err error
}

var graphScenarios = []graphTest{
	{
		title: "happy path empty",
		input: input{
			depFileName: "dependencies.yaml",
			files:       map[string][]byte{"someOtherFile": {}},
			mock: mockGraph{
				rf: nil,
				un: nil,
				rd: nil,
			},
		},
		want: want{
			res: graph.ComponentDependencies{},
			err: nil,
		},
	},
	{
		title: "happy path no dependencies",
		input: input{
			depFileName: "dependencies.yaml",
			files: map[string][]byte{
				"component-1/dependencies.yaml": []byte(string(`
name: component-1
dependencies:
`)),
				"component-2/dependencies.yaml": []byte(string(`
name: component-2
dependencies:
`)),
			},
			mock: mockGraph{
				rf: nil,
				un: nil,
				rd: nil,
			},
		},
		want: want{
			res: graph.ComponentDependencies{
				"component-1": {},
				"component-2": {},
			},
			err: nil,
		},
	},
	{
		title: "happy path with dependencies",
		input: input{
			depFileName: "dependencies.yaml",
			files: map[string][]byte{
				"component-1/dependencies.yaml": []byte(string(`
name: component-1
dependencies:
- component-2
- unknown-component
`)),
				"component-2/dependencies.yaml": []byte(string(`
name: component-2
dependencies:
`)),
			},
			mock: mockGraph{
				rf: nil,
				un: nil,
				rd: nil,
			},
		},
		want: want{
			res: graph.ComponentDependencies{
				"component-1":       {},
				"component-2":       {0: {"component-1": true}},
				"unknown-component": {0: {"component-1": true}},
			},
			err: nil,
		},
	},
	{
		title: "folder selected",
		input: input{
			depFileName: "folder",
			files:       map[string][]byte{"folder/otherFile": {}},
			mock: mockGraph{
				rf: nil,
				un: nil,
				rd: nil,
			},
		},
		want: want{
			res: graph.ComponentDependencies{"": {}},
			err: nil,
		},
	},
	{
		title: "error in readDeps",
		input: input{
			depFileName: "dependencies.yaml",
			files:       map[string][]byte{},
			mock: mockGraph{
				rf: nil,
				un: nil,
				rd: fmt.Errorf("fail in readDeps"),
			},
		},
		want: want{
			res: graph.ComponentDependencies{},
			err: fmt.Errorf("fail in readDeps"),
		},
	},
	{
		title: "error in readFile",
		input: input{
			depFileName: "file",
			files:       map[string][]byte{"file": {}},
			mock: mockGraph{
				rf: fmt.Errorf("fail in readFile"),
				un: nil,
				rd: nil,
			},
		},
		want: want{
			res: graph.ComponentDependencies{},
			err: fmt.Errorf("fail in readFile"),
		},
	},
	{
		title: "error in unmarshal",
		input: input{
			depFileName: "file",
			files:       map[string][]byte{"file": {}},
			mock: mockGraph{
				rf: nil,
				un: fmt.Errorf("fail in unmarshal"),
				rd: nil,
			},
		},
		want: want{
			res: graph.ComponentDependencies{},
			err: fmt.Errorf("fail in unmarshal"),
		},
	},
}

func TestGraph(t *testing.T) {
	if err := log.Init(log.Debug(), "", true); err != nil {
		zap.S().Fatal(err)
	}
	for _, s := range graphScenarios {
		t.Logf("test scenario: %s\n", s.title)
		s.Test(t)
	}
}

func (g *graphTest) Test(t *testing.T) {
	testDir, err := g.setup()
	if err != nil {
		t.Logf("unable to create test dir tree: %v\n", err)
		t.FailNow()
	}
	defer testDir.Cleanup(t)

	got, _, err := Graph(testDir.Path(), g.input.depFileName)
	testfuncs.CheckErrs(t, g.want.err, err)
	checkRes(t, g.want.res, got)
}

func (g *graphTest) setup() (testfuncs.TestDir, error) {
	readFile = g.input.mock.readFile
	unmarshal = g.input.mock.unmarshal
	dependencies = g.input.mock.readDeps

	return testfuncs.PrepareTestDirTree(g.input.files)
}

func checkRes(t *testing.T, want, got graph.ComponentDependencies) {
	if !reflect.DeepEqual(got, want) {
		t.Errorf(
			"results do not match: \nwant = \"%+v\"\ngot =  \"%+v\"",
			want, got,
		)
		t.Fail()
	}
}

type mockGraph struct {
	rf error
	un error
	rd error
}

func (m mockGraph) readFile(f string) ([]byte, error) {
	if m.rf != nil {
		return nil, m.rf
	}
	return os.ReadFile(f)
}

func (m mockGraph) unmarshal(in []byte, out interface{}) error {
	if m.un != nil {
		return m.un
	}
	return yaml.Unmarshal(in, out)
}

func (m mockGraph) readDeps(path, depFileName string) (*files.Files, error) {
	if m.rd != nil {
		return nil, m.rd
	}
	return deps(path, depFileName)
}
