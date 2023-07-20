package yamlfile_test

import (
	"github.com/SAP/configuration-tools-for-gitops/pkg/yamlfile"
	"reflect"
	"testing"
)

type scenarioDoc struct {
	title      string
	input      interface{}
	wantOutput interface{}
}

var scenariosDocs = []scenarioDoc{
	{
		title: "struct example",
		input: struct {
			Key1 string `doc:"msg=this is a description,default=hello"`
			Key2 bool   `yaml:"custom_key" doc:"req"`
			Key3 int    `doc:"req=by this"`
			Key4 int    `doc:"o=0,o=1,req"`
		}{},
		wantOutput: map[string]interface{}{
			"Key1":       `this is a description (string, default:"hello")`,
			"Key3":       `(int) REQUIRED:"by this"`,
			"custom_key": `(bool) REQUIRED`,
			"Key4":       `(int, options:[0,1]) REQUIRED`,
		},
	},
	{
		title: "map example",
		input: map[string]struct {
			Key1 string `doc:"msg=this is a description,default=hello"`
		}{},
		wantOutput: map[string]interface{}{
			"string": map[string]interface{}{
				"Key1": `this is a description (string, default:"hello")`,
			},
		},
	},
	{
		title: "slice example",
		input: []struct {
			Key1 string `doc:"msg=this is a description,default=hello"`
		}{},
		wantOutput: []interface{}{
			map[string]interface{}{
				"Key1": `this is a description (string, default:"hello")`,
			},
		},
	},
	{
		title: "pointer example",
		input: struct {
			Key1 *string `doc:"msg=this is a description,default=hello"`
			Key2 *struct {
				A *map[string]string `doc:"msg=no deeper resolution"`
			}
			Key3 *[]struct {
				B string
				C []string `doc:"msg=no deeper resolution"`
			}
		}{},
		wantOutput: map[string]interface{}{
			"Key1": `this is a description (string, default:"hello")`,
			"Key2": map[string]interface{}{"A": "no deeper resolution (map[string]string)"},
			"Key3": []interface{}{map[string]interface{}{
				"B": "(string)",
				"C": "no deeper resolution ([]string)",
			}},
		},
	},
	{
		title: "nested structs",
		input: struct {
			Root struct {
				L11 string `doc:"req"`
				L12 struct {
					L2 string `doc:"req"`
				}
			}
		}{},
		wantOutput: map[string]interface{}{
			"Root": map[string]interface {
			}{
				"L11": "(string) REQUIRED",
				"L12": map[string]interface{}{
					"L2": "(string) REQUIRED",
				},
			},
		},
	},
	{
		title: "nested map and slice",
		input: struct {
			Root struct {
				Slice []struct {
					S bool `doc:"req"`
				}
				Map map[string]struct{ M string }
			}
		}{},
		wantOutput: map[string]interface{}{
			"Root": map[string]interface {
			}{
				"Slice": []interface{}{map[string]interface{}{"S": "(bool) REQUIRED"}},
				"Map": map[string]interface{}{
					"string": map[string]interface{}{"M": "(string)"},
				},
			},
		},
	},
	{
		title: "ignore keys with doc tag -",
		input: struct {
			Key1       string `doc:"msg=this is a description,default=hello"`
			NotIgnored string
			Ignored    string `yaml:"ignored" doc:"-"`
		}{},
		wantOutput: map[string]interface{}{
			"Key1":       `this is a description (string, default:"hello")`,
			"NotIgnored": `(string)`,
		},
	},
	{
		title: "wrong doc string format",
		input: struct {
			Key1 string `doc:"irrelevant content,msg="`
		}{},
		wantOutput: map[string]interface{}{
			"Key1": `(string)`,
		},
	},
}

func TestDocOutput(t *testing.T) {
	for _, s := range scenariosDocs {
		t.Logf("test scenario: %s\n", s.title)
		res := yamlfile.DocOutput(s.input)
		s.Check(t, res)
	}
}

func (s *scenarioDoc) Check(t *testing.T, got interface{}) {
	if !reflect.DeepEqual(s.wantOutput, got) {
		t.Errorf(
			"results do not match: \nwant = \"%+v\"\ngot =  \"%+v\"",
			s.wantOutput,
			got,
		)
		t.Fail()
	}
}
