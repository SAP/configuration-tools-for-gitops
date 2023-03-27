package yamlfile_test

import (
	"errors"
	"reflect"
	"testing"

	"github.com/configuration-tools-for-gitops/pkg/testfuncs"
	"github.com/configuration-tools-for-gitops/pkg/yamlfile"
)

type scenarioMerge struct {
	title  string
	input1 interface{}
	input2 interface{}
	want   interface{}
	err    error
}

var scenariosMerge = []scenarioMerge{
	{
		title: "simple merge",
		input1: map[string]interface{}{
			"key":      "value",
			"key-base": map[string]interface{}{"keynested": "nested-value"},
			"key-base-complex": map[string]interface{}{
				"nested": "nested-value",
				"multi": map[string]interface{}{
					"2deep":      "value",
					"2deep-stay": "value",
				},
			},
		},
		input2: map[string]interface{}{
			"new-key": "value",
			"key":     "value-overwrite",
			"key-base-complex": map[string]interface{}{
				"nested": "nested-value-overwrite",
				"multi": map[string]interface{}{
					"2deep": "value-overwrite",
				},
			},
		},
		want: map[string]interface{}{
			"new-key":  "value",
			"key":      "value-overwrite",
			"key-base": map[string]interface{}{"keynested": "nested-value"},
			"key-base-complex": map[string]interface{}{
				"nested": "nested-value-overwrite",
				"multi": map[string]interface{}{
					"2deep":      "value-overwrite",
					"2deep-stay": "value",
				},
			},
		},
		err: nil,
	},
	{
		title:  "no input 1",
		input1: nil,
		input2: map[string]interface{}{"key": "value"},
		want:   map[string]interface{}{"key": "value"},
		err:    nil,
	},
	{
		title:  "no input 2",
		input1: map[string]interface{}{"key": "value"},
		input2: "hello",
		want:   "hello",
		err:    nil,
	},
	{
		title:  "default case",
		input1: []float64{6.1},
		input2: map[string]interface{}{"key": "value"},
		want:   nil,
		err:    errors.New("type []float64 not implemented for merging"),
	},
	{
		title:  "incompatible types 1",
		input1: []string{"hello"},
		input2: map[string]interface{}{"key": "value"},
		want:   nil,
		err:    errors.New("cannot merge types []string and map[string]interface {}"),
	},
	{
		title:  "incompatible types 2",
		input1: []interface{}{"hello"},
		input2: []string{"val"},
		want:   nil,
		err:    errors.New("cannot merge types []interface {} and []string"),
	},
	{
		title:  "incompatible types 3",
		input1: map[string]interface{}{"key": "value"},
		input2: []string{"val"},
		want:   nil,
		err:    errors.New("cannot merge types map[string]interface {} and []string"),
	},
	{
		title: "complex merge",
		input1: map[string]interface{}{
			"key":      "value",
			"key-base": map[string]interface{}{"keynested": "nested-value"},
			"key-base-complex": map[string]interface{}{
				"nested": "nested-value",
				"multi": map[string]interface{}{
					"2deep":      "value",
					"2deep-stay": "value",
				},
			},
			"array": []string{"element0", "element1", "element2"},
			"nestedArray": map[string]interface{}{
				"array": []interface{}{
					"nestedEl0",
					"nestedEl1",
					[]string{"hello"},
					map[string]interface{}{"key": "value"},
				},
			},
		},
		input2: map[string]interface{}{
			"key": "value-overwrite",
			"key-base-complex": map[string]interface{}{
				"nested": "nested-value-overwrite",
				"multi": map[string]interface{}{
					"2deep": "value-overwrite",
				},
			},
			"array": []string{"added-element"},
			"nestedArray": map[string]interface{}{
				"array": []interface{}{
					"added-element",
					[]string{"added-array"},
					map[string]interface{}{"added": "map"},
				},
			},
		},
		want: map[string]interface{}{
			"key":      "value-overwrite",
			"key-base": map[string]interface{}{"keynested": "nested-value"},
			"key-base-complex": map[string]interface{}{
				"nested": "nested-value-overwrite",
				"multi": map[string]interface{}{
					"2deep":      "value-overwrite",
					"2deep-stay": "value",
				},
			},
			"array": []string{"element0", "element1", "element2", "added-element"},
			"nestedArray": map[string]interface{}{
				"array": []interface{}{
					"nestedEl0",
					"nestedEl1",
					[]string{"hello"},
					map[string]interface{}{"key": "value"},
					"added-element",
					[]string{"added-array"},
					map[string]interface{}{"added": "map"},
				},
			},
		},
		err: nil,
	},
}

func (s *scenarioMerge) CheckRes(t *testing.T, got interface{}) {
	if !reflect.DeepEqual(s.want, got) {
		t.Errorf(
			"results do not match: \nwant = \"%+v\"\ngot = \"%+v\"",
			s.want,
			got,
		)
		t.Fail()
	}
}

func TestMerge(t *testing.T) {
	testLogger()
	for _, s := range scenariosMerge {
		t.Logf("test scenario: %s\n", s.title)

		got, err := yamlfile.Merge(s.input1, s.input2)
		testfuncs.CheckErrs(t, s.err, err)

		s.CheckRes(t, got)
	}
}
