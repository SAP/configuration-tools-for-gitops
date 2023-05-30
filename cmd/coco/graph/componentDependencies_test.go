package graph

import (
	"fmt"
	"reflect"
	"strings"
	"testing"

	"github.com/SAP/configuration-tools-for-gitops/pkg/testfuncs"
)

type scenarioComponentDeps struct {
	title string
	input ComponentDependencies
	want  []wantCD
}

type wantCD struct {
	keys      []string
	printYAML string
	printJSON string
	printFLAT string
	res       ComponentDependencies
	maxLevel  int
	err       error
}

var scenariosCD = []scenarioComponentDeps{
	{
		title: "minimal example 1",
		// A → + → B   E → F
		//	   ↓   ↓
		//     + → C → D
		input: ComponentDependencies{
			"A": {},
			"B": {0: {"A": true}},
			"C": {0: {"A": true, "B": true}},
			"D": {0: {"C": true}, 1: {"A": true, "B": true}},
			"E": {},
			"F": {0: {"E": true}},
		},
		want: []wantCD{
			{
				maxLevel: -1,
				keys:     []string{"A", "B", "C", "D", "E", "F"},
				res: ComponentDependencies{
					"A": {},
					"B": {0: {"A": true}},
					"C": {0: {"A": true, "B": true}},
					"D": {0: {"C": true}, 1: {"A": true, "B": true}},
					"E": {},
					"F": {0: {"E": true}},
				},
				printYAML: strings.TrimLeft(`
A: {}
B:
  0: [A]
C:
  0: [A B]
D:
  0: [C]
  1: [A B]
E: {}
F:
  0: [E]
`, "\n"),
				printJSON: `{"A":[],"B":[["A"]],"C":[["A","B"]],"D":[["C"],["A","B"]],"E":[],"F":[["E"]]}`,
				printFLAT: `A B C D E F`,
				err:       nil,
			},
			{
				maxLevel: 0,
				keys:     []string{"A", "B", "C", "D", "E", "F"},
				res: ComponentDependencies{
					"A": {},
					"B": {},
					"C": {},
					"D": {},
					"E": {},
					"F": {},
				},
				printYAML: strings.TrimLeft(`
A: {}
B: {}
C: {}
D: {}
E: {}
F: {}
`, "\n"),
				printJSON: `{"A":[],"B":[],"C":[],"D":[],"E":[],"F":[]}`,
				printFLAT: `A B C D E F`,
				err:       nil,
			},
			{
				maxLevel: 1,
				keys:     []string{"A", "B", "C", "D", "E", "F"},
				res: ComponentDependencies{
					"A": {},
					"B": {0: {"A": true}},
					"C": {0: {"A": true, "B": true}},
					"D": {0: {"C": true}},
					"E": {},
					"F": {0: {"E": true}},
				},
				printYAML: strings.TrimLeft(`
A: {}
B:
  0: [A]
C:
  0: [A B]
D:
  0: [C]
E: {}
F:
  0: [E]
`, "\n"),
				printJSON: `{"A":[],"B":[["A"]],"C":[["A","B"]],"D":[["C"]],"E":[],"F":[["E"]]}`,
				printFLAT: `A B C D E F`,
				err:       nil,
			},
			{
				maxLevel: 2,
				keys:     []string{"A", "B", "C", "D", "E", "F"},
				res: ComponentDependencies{
					"A": {},
					"B": {0: {"A": true}},
					"C": {0: {"A": true, "B": true}},
					"D": {0: {"C": true}, 1: {"A": true, "B": true}},
					"E": {},
					"F": {0: {"E": true}},
				},
				printYAML: strings.TrimLeft(`
A: {}
B:
  0: [A]
C:
  0: [A B]
D:
  0: [C]
  1: [A B]
E: {}
F:
  0: [E]
`, "\n"),
				printJSON: `{"A":[],"B":[["A"]],"C":[["A","B"]],"D":[["C"],["A","B"]],"E":[],"F":[["E"]]}`,
				printFLAT: `A B C D E F`,
				err:       nil,
			},
			{
				maxLevel: 3,
				keys:     []string{"A", "B", "C", "D", "E", "F"},
				res: ComponentDependencies{
					"A": {},
					"B": {0: {"A": true}},
					"C": {0: {"A": true, "B": true}},
					"D": {0: {"C": true}, 1: {"A": true, "B": true}},
					"E": {},
					"F": {0: {"E": true}},
				},
				printYAML: strings.TrimLeft(`
A: {}
B:
  0: [A]
C:
  0: [A B]
D:
  0: [C]
  1: [A B]
E: {}
F:
  0: [E]
`, "\n"),
				printJSON: `{"A":[],"B":[["A"]],"C":[["A","B"]],"D":[["C"],["A","B"]],"E":[],"F":[["E"]]}`,
				printFLAT: `A B C D E F`,
				err:       nil,
			},
		},
	},
	{
		title: "fail",
		// A → B
		input: ComponentDependencies{
			"A": {},
			"B": {0: {"A": true}},
		},
		want: []wantCD{
			{
				maxLevel:  0,
				keys:      []string{"A", "B"},
				res:       ComponentDependencies{"A": {}, "B": {}},
				printYAML: "",
				printJSON: "",
				printFLAT: "",
				err:       fmt.Errorf("intended failure"),
			},
		},
	},
	{
		title: "restricted output",
		input: ComponentDependencies{
			"C": {0: {"A": true, "B": true}},
			"F": {0: {"E": true}},
		},
		want: []wantCD{
			{
				maxLevel: -1,
				keys:     []string{"C", "F"},
				res: ComponentDependencies{
					"C": {0: {"A": true, "B": true}},
					"F": {0: {"E": true}},
				},
				printYAML: strings.TrimLeft(`
C:
  0: [A B]
F:
  0: [E]
`, "\n"),
				printJSON: `{"C":[["A","B"]],"F":[["E"]]}`,
				printFLAT: `A B C E F`,
				err:       nil,
			},
			{
				maxLevel: 0,
				keys:     []string{"C", "F"},
				res: ComponentDependencies{
					"C": {},
					"F": {},
				},
				printYAML: strings.TrimLeft(`
C: {}
F: {}
`, "\n"),
				printJSON: `{"C":[],"F":[]}`,
				printFLAT: `C F`,
				err:       nil,
			},
		},
	},
}

func TestComponentDependenciesFunctions(t *testing.T) {
	for _, s := range scenariosCD {
		for _, it := range s.want {
			t.Logf("test scenario: %s - depth: %v\n", s.title, it.maxLevel)

			it.CheckKeys(t, s.input.Keys())

			res := s.input.MaxDepth(it.maxLevel)
			it.CheckRes(t, res)

			printYAML := testWriter{}
			if it.err != nil {
				printYAML = testWriter{fail: it.err}
			}
			err := res.Print(&printYAML, YAML)
			testfuncs.CheckErrs(t, it.err, err)

			printJSON := testWriter{}
			if it.err != nil {
				printJSON = testWriter{fail: it.err}
			}
			err = res.Print(&printJSON, JSON)
			testfuncs.CheckErrs(t, it.err, err)

			printFLAT := testWriter{}
			if it.err != nil {
				printFLAT = testWriter{fail: it.err}
			}
			err = res.Print(&printFLAT, FLAT)
			testfuncs.CheckErrs(t, it.err, err)

			it.CheckPrints(t, string(printYAML.data), string(printJSON.data), string(printFLAT.data))
		}
	}
}

func (w *wantCD) CheckKeys(t *testing.T, got []string) {
	if !reflect.DeepEqual(w.keys, got) {
		t.Errorf("keys do not match: \nwant = %+v\ngot =  %+v",
			w.keys, got)
		t.Fail()
	}
}

func (w *wantCD) CheckPrints(t *testing.T, gotYAML, gotJSON, gotFLAT string) {
	if w.printYAML != gotYAML {
		t.Errorf("printed yaml results do not match: \nwant = \"\n%+v\"\ngot =  \"\n%+v\"",
			w.printYAML, gotYAML)
		t.Fail()
	}
	if w.printJSON != gotJSON {
		t.Errorf("printed json results do not match: \nwant = \"\n%+v\"\ngot =  \"\n%+v\n\"",
			w.printJSON, gotJSON)
		t.Fail()
	}
	if w.printFLAT != gotFLAT {
		t.Errorf("printed flat results do not match: \nwant = \"\n%+v\"\ngot =  \"\n%+v\n\"",
			w.printFLAT, gotFLAT)
		t.Fail()
	}
}

func (w *wantCD) CheckRes(t *testing.T, got ComponentDependencies) {
	if !reflect.DeepEqual(w.res, got) {
		t.Errorf("max-depth results do not match: \nwant = %+v\ngot =  %+v",
			w.res, got)
		t.Fail()
	}
}

type testWriter struct {
	data []byte
	fail error
}

func (t *testWriter) Write(p []byte) (n int, err error) {
	if t.fail != nil {
		return 0, t.fail
	}
	t.data = append(t.data, p...)
	return len(p), nil
}
