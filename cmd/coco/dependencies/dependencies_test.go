package dependencies

import (
	"fmt"
	"testing"

	"github.com/configuration-tools-for-gitops/cmd/coco/graph"
	"github.com/configuration-tools-for-gitops/pkg/git"
	"github.com/configuration-tools-for-gitops/pkg/log"
	"github.com/configuration-tools-for-gitops/pkg/testfuncs"
	"go.uber.org/zap"
)

const (
	notUsed = "notUsed"
)

type scenarioDependency struct {
	title string
	input inputDeps
	want  wantDeps
}

type inputDeps struct {
	graphDepth            int
	diffFiles             []string
	componentPaths        map[string]string
	componentDependencies graph.ComponentDependencies
	errors                errors
}

type errors struct {
	repo      error
	checkout  error
	mergeBase error
	diffPaths error
	graph     error
}

type wantDeps struct {
	res graph.ComponentDependencies
	err error
}

var scenariosDependencies = []scenarioDependency{
	{
		title: "happy path",
		input: inputDeps{
			graphDepth: -1,
			diffFiles:  []string{"folder/path/c1/hello.sh"},
			componentPaths: map[string]string{
				"c1": "folder/path/c1",
				"c2": "c2",
			},
			componentDependencies: graph.ComponentDependencies{
				"c1": {},
				"c2": {0: {"c1": true}},
			},
			errors: errors{},
		},
		want: wantDeps{
			res: map[string]graph.WeightedDeps{"c1": {}},
			err: nil,
		},
	},
	{
		title: "happy path 2",
		input: inputDeps{
			graphDepth: -1,
			diffFiles:  []string{"folder/path/c2/subpath/hello.sh"},
			componentPaths: map[string]string{
				"c1": "folder/path/c1",
				"c2": "folder/path/c2",
			},
			componentDependencies: graph.ComponentDependencies{
				"c1": {},
				"c2": {0: {"c1": true}},
			},
			errors: errors{},
		},
		want: wantDeps{
			res: map[string]graph.WeightedDeps{"c2": {0: {"c1": true}}},
			err: nil,
		},
	},
	{
		title: "happy path multi components",
		input: inputDeps{
			graphDepth: -1,
			diffFiles: []string{
				"c2/subpath/hello.sh",
				"folder/c3/f",
			},
			componentPaths: map[string]string{
				"c1": "folder/path/c1",
				"c2": "c2",
				"c3": "folder/c3",
			},
			componentDependencies: graph.ComponentDependencies{
				"c1": {},
				"c2": {0: {"c1": true}},
				"c3": {0: {"c2": true}, 1: {"c1": true}},
			},
			errors: errors{},
		},
		want: wantDeps{
			res: map[string]graph.WeightedDeps{
				"c2": {0: {"c1": true}},
				"c3": {0: {"c2": true}, 1: {"c1": true}},
			},
			err: nil,
		},
	},
	{
		title: "happy path restrict depth",
		input: inputDeps{
			graphDepth: 1,
			diffFiles: []string{
				"c2/subpath/hello.sh",
				"folder/c3/f",
			},
			componentPaths: map[string]string{
				"c1": "folder/path/c1",
				"c2": "c2",
				"c3": "folder/c3",
			},
			componentDependencies: graph.ComponentDependencies{
				"c1": {},
				"c2": {0: {"c1": true}},
				"c3": {0: {"c2": true}, 1: {"c1": true}},
			},
			errors: errors{},
		},
		want: wantDeps{
			res: map[string]graph.WeightedDeps{
				"c2": {0: {"c1": true}},
				"c3": {0: {"c2": true}},
			},
			err: nil,
		},
	},
	{
		title: "no match",
		input: inputDeps{
			graphDepth:            -1,
			diffFiles:             []string{"folder/c1/notSelected"},
			componentPaths:        map[string]string{"c1": "folder/path/c1"},
			componentDependencies: graph.ComponentDependencies{"c1": {}},
			errors:                errors{},
		},
		want: wantDeps{
			res: map[string]graph.WeightedDeps{},
			err: nil,
		},
	},
	{
		title: "no match in subpath",
		input: inputDeps{
			graphDepth: -1,
			diffFiles:  []string{"folder/path/c2/subpath/hello.sh"},
			componentPaths: map[string]string{
				"c1": "folder/path/c1",
				"c2": "c2",
			},
			componentDependencies: graph.ComponentDependencies{
				"c1": {},
				"c2": {0: {"c1": true}},
			},
			errors: errors{},
		},
		want: wantDeps{
			res: map[string]graph.WeightedDeps{},
			err: nil,
		},
	},
	{
		title: "error repo",
		input: inputDeps{
			graphDepth: -1,
			diffFiles:  []string{"folder/path/c1/hello.sh"},
			componentPaths: map[string]string{
				"c1": "folder/path/c1",
			},
			componentDependencies: graph.ComponentDependencies{
				"c1": {},
			},
			errors: errors{
				checkout: fmt.Errorf("induced failure"),
			},
		},
		want: wantDeps{
			res: map[string]graph.WeightedDeps{},
			err: fmt.Errorf("induced failure"),
		},
	},
	{
		title: "error repo",
		input: inputDeps{
			graphDepth: -1,
			diffFiles:  []string{"folder/path/c1/hello.sh"},
			componentPaths: map[string]string{
				"c1": "folder/path/c1",
			},
			componentDependencies: graph.ComponentDependencies{
				"c1": {},
			},
			errors: errors{
				mergeBase: fmt.Errorf("induced failure"),
			},
		},
		want: wantDeps{
			res: map[string]graph.WeightedDeps{},
			err: fmt.Errorf("induced failure"),
		},
	},
	{
		title: "error repo",
		input: inputDeps{
			graphDepth: -1,
			diffFiles:  []string{"folder/path/c1/hello.sh"},
			componentPaths: map[string]string{
				"c1": "folder/path/c1",
			},
			componentDependencies: graph.ComponentDependencies{
				"c1": {},
			},
			errors: errors{
				diffPaths: fmt.Errorf("induced failure"),
			},
		},
		want: wantDeps{
			res: map[string]graph.WeightedDeps{},
			err: fmt.Errorf("induced failure"),
		},
	},
	{
		title: "error repo",
		input: inputDeps{
			graphDepth: -1,
			diffFiles:  []string{"folder/path/c1/hello.sh"},
			componentPaths: map[string]string{
				"c1": "folder/path/c1",
			},
			componentDependencies: graph.ComponentDependencies{
				"c1": {},
			},
			errors: errors{
				graph: fmt.Errorf("induced failure"),
			},
		},
		want: wantDeps{
			res: map[string]graph.WeightedDeps{},
			err: fmt.Errorf("induced failure"),
		},
	},
}

func TestChangeAffectedComponents(t *testing.T) {
	if err := log.Init(log.Debug(), "", true); err != nil {
		zap.S().Fatal(err)
	}
	for _, s := range scenariosDependencies {
		t.Logf("test scenario: %s\n", s.title)
		repo = s.input.repo
		mergeBase = s.input.mergeBase
		diffPaths = s.input.diffPaths
		graphh = s.input.graph

		got, err := ChangeAffectedComponents(
			notUsed, notUsed, notUsed, notUsed, notUsed, notUsed, notUsed, s.input.graphDepth, 0,
			log.Debug(),
		)
		testfuncs.CheckErrs(t, s.want.err, err)
		checkRes(t, s.want.res, got)
	}
}

func (m *inputDeps) repo(
	path, url, token, remote, defaultBranch string, maxDepth int, logLvl log.Level,
) (gitRepo, error) {
	if m.errors.repo != nil {
		return nil, m.errors.repo
	}
	return m, nil
}
func (m *inputDeps) Checkout(branch string, force bool) (res *git.Tree, err error) {
	return nil, m.errors.checkout
}
func (m *inputDeps) mergeBase(*git.Tree, *git.Tree) (*git.Tree, error) {
	return nil, m.errors.mergeBase
}
func (m *inputDeps) diffPaths(*git.Tree, *git.Tree) ([]string, error) {
	return m.diffFiles, m.errors.diffPaths
}
func (m *inputDeps) graph(string, string) (graph.ComponentDependencies, map[string]string, error) {
	return m.componentDependencies, m.componentPaths, m.errors.graph
}
