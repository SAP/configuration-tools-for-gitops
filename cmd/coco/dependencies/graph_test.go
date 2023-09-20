package dependencies

import (
	"fmt"
	"reflect"
	"testing"

	"github.com/SAP/configuration-tools-for-gitops/v2/cmd/coco/inputfile"

	"github.com/SAP/configuration-tools-for-gitops/v2/cmd/coco/graph"
	"github.com/SAP/configuration-tools-for-gitops/v2/pkg/files"
	"github.com/SAP/configuration-tools-for-gitops/v2/pkg/log"
	"github.com/SAP/configuration-tools-for-gitops/v2/pkg/testfuncs"
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
			depFileName: "coco.yaml",
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
			depFileName: "coco.yaml",
			files: map[string][]byte{
				"component-1/coco.yaml": []byte(string(`
type: component
name: component-1
dependencies:
`)),
				"component-2/coco.yaml": []byte(string(`
type: component
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
			depFileName: "coco.yaml",
			files: map[string][]byte{
				"component-1/coco.yaml": []byte(string(`
type: component
name: component-1
dependencies:
- component-2
- unknown-component
`)),
				"component-2/coco.yaml": []byte(string(`
type: component
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
			depFileName: "coco.yaml",
			files: map[string][]byte{
				"coco.yaml": []byte(`
type: component
dependencies:
`),
				"folder/otherFile": {}},
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
			depFileName: "coco.yaml",
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
		title: "error in unmarshal",
		input: input{
			depFileName: "coco.yaml",
			files: map[string][]byte{"coco.yaml": []byte(`
type: component
name: component-1
dependencies:
  - dep1`),
				"dep1": {}},
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

func (m mockGraph) unmarshal(in []byte, out interface{}) error {
	if m.un != nil {
		return m.un
	}
	return yaml.Unmarshal(in, out)
}

func (m mockGraph) readDeps(
	path,
	depFileName string,
	includeOr,
	includeAnd,
	exclude []string,
) (map[string]files.File, error) {
	if m.rd != nil {
		return nil, m.rd
	}
	return inputfile.FindAll(path, depFileName, includeOr, includeAnd, exclude)
}
