package graph

import (
	"reflect"
	"testing"
)

type scenario struct {
	title              string
	components         DownToUp
	wantDownstreamDeps ComponentDependencies
}

var scenarios = []scenario{
	{
		title: "minimal example 1",
		// A → + → B   E → F
		//	   ↓   ↓
		//     + → C → D
		components: DownToUp{
			"A": {"B": true, "C": true},
			"B": {"C": true},
			"C": {"D": true},
			"D": {},
			"E": {"F": true},
			"F": {},
		},
		wantDownstreamDeps: ComponentDependencies{
			"A": {},
			"B": {0: {"A": true}},
			"C": {0: {"A": true, "B": true}},
			"D": {0: {"C": true}, 1: {"A": true, "B": true}},
			"E": {},
			"F": {0: {"E": true}},
		},
	},
	{
		title: "minimal example 2",
		// A → + → B
		//	   ↓   ↓
		//     + → C → D → E → F
		components: DownToUp{
			"A": {"B": true, "C": true},
			"B": {"C": true},
			"C": {"D": true},
			"D": {"E": true},
			"E": {"F": true},
			"F": {},
		},
		wantDownstreamDeps: ComponentDependencies{
			"A": {},
			"B": {0: {"A": true}},
			"C": {0: {"A": true, "B": true}},
			"D": {0: {"C": true}, 1: {"A": true, "B": true}},
			"E": {0: {"D": true}, 1: {"C": true}, 2: {"A": true, "B": true}},
			"F": {0: {"E": true}, 1: {"D": true}, 2: {"C": true}, 3: {"A": true, "B": true}},
		},
	},
	{
		title: "cyclic dependency",
		// A → B
		// ↑   ↓
		// + ← +
		components: DownToUp{
			"A": {"B": true},
			"B": {"A": true},
		},
		wantDownstreamDeps: ComponentDependencies{
			"A": {0: {"B": true}},
			"B": {0: {"A": true}},
		},
	},
}

func TestGenerateGraph(t *testing.T) {
	for _, s := range scenarios {
		t.Logf("test scenario: %s\n", s.title)
		orderedDependencies := GenerateUpToDown(s.components)
		s.CheckRes(t, orderedDependencies)
	}
}

func (s scenario) CheckRes(t *testing.T, got ComponentDependencies) {
	if !reflect.DeepEqual(s.wantDownstreamDeps, got) {
		t.Errorf("results do not match: \nwant = \"%+v\"\ngot =  \"%+v\"", s.wantDownstreamDeps, got)
		t.Fail()
	}
}
