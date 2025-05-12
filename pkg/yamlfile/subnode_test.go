package yamlfile_test

import (
	"bytes"
	"errors"
	"fmt"
	"reflect"
	"strings"
	"testing"

	"github.com/SAP/configuration-tools-for-gitops/v2/pkg/testfuncs"
	"github.com/SAP/configuration-tools-for-gitops/v2/pkg/yamlfile"
)

var scenariosSelectSubNode = []scenarioSelectSubNode{
	{
		title:   "unmarshal fails",
		input:   []byte(strings.TrimSpace(`key: {`)),
		want:    "",
		wantErr: fmt.Errorf("unmarshalling failed %s", "yaml: line 1: did not find expected node content"),
	},
	{
		title: "filter by keys",
		input: []byte(strings.TrimSpace(`
k1: v1
keep:
  not: this
  only:
    not: this either
    this:
      part:
        k2: v2
        k3: v3
        arr:
          - a1
          - a2
k2: v2
`)),
		subNodeKeys: []string{"keep", "only", "this", "part"},
		want: fmt.Sprintf("%s\n", strings.TrimSpace(`
k2: v2
k3: v3
arr:
  - a1
  - a2
`)),
		wantErr: nil,
	},
	{
		title: "insert subnode",
		input: []byte(strings.TrimSpace(`
k1: v1
k2: v2
`)),
		subNodeKeys: []string{"add", "this", "part"},
		insert:      map[string]interface{}{"k1": "v1", "k2": "v2"},
		want: fmt.Sprintf("%s\n", strings.TrimSpace(`
k1: v1
k2: v2
add:
  this:
    part:
      k1: v1
      k2: v2
`)),
		wantErr: nil,
	},
	{
		title: "overwrite subnode",
		input: []byte(strings.TrimSpace(`
k1: v1
keep:
  not: this
  only:
    not: this either
    this:
      part:
        k2: v2
        k3: v3
        arr:
          - a1
          - a2
k2: v2
`)),
		subNodeKeys: []string{"keep", "only", "this", "part"},
		insert:      map[string]interface{}{"k2": "overwrite"},
		want: fmt.Sprintf("%s\n", strings.TrimSpace(`
k1: v1
keep:
  not: this
  only:
    not: this either
    this:
      part:
        k2: overwrite
k2: v2
`)),
		wantErr: nil,
	},
	{
		title: "filter out everything",
		input: []byte(strings.TrimSpace(`
k1: v1
doNotKeep:
  not: this
  only:
    not: this either
    this:
      part:
        k2: v2
        k3: v3
        arr:
          - a1
          - a2
k2: v2
`)),
		subNodeKeys: []string{"keep", "only", "this", "part"},
		want:        ``,
		wantErr:     errors.New("key not present in yaml.Node: \"keep\""),
	},
}

func TestSelectSubNode(t *testing.T) {
	testLogger()
	for _, s := range scenariosSelectSubNode {
		t.Logf("test scenario: %s\n", s.title)

		y, err := yamlfile.New(s.input)
		if s.title == "unmarshal fails" {
			testfuncs.CheckErrs(t, s.wantErr, err)
			continue
		}
		testfuncs.CheckErrs(t, nil, err)

		var res yamlfile.Yaml
		var warnings []yamlfile.Warning
		if s.insert != nil {
			warnings, err = y.Insert(s.subNodeKeys, s.insert)
			s.CheckWarnings(t, warnings)
			testfuncs.CheckErrs(t, s.wantErr, err)
			res = y
		} else {
			res, err = y.SelectSubElement(s.subNodeKeys)
			testfuncs.CheckErrs(t, s.wantErr, err)
		}

		if err == nil {
			var gotBytes bytes.Buffer
			err = res.Encode(&gotBytes, 2)
			testfuncs.CheckErrs(t, nil, err)
			if !s.CheckRes(t, gotBytes.String()) {
				scetchNodes(y.Node, []int{})
			}
		}
	}
}

type scenarioSelectSubNode struct {
	title        string
	input        []byte
	subNodeKeys  []string
	insert       interface{}
	want         string
	wantWarnings []yamlfile.Warning
	wantErr      error
}

func (s *scenarioSelectSubNode) CheckRes(t *testing.T, got string) bool {
	if s.want != got {
		t.Errorf(
			"results do not match: \nwant = \"\n%+v\"\ngot = \"\n%+v\"",
			s.want,
			got,
		)
		t.Fail()
		return false
	}
	return true
}

func (s *scenarioSelectSubNode) CheckWarnings(t *testing.T, got []yamlfile.Warning) {
	if len(got) != len(s.wantWarnings) {
		t.Errorf(
			"warnings do not match: \nwant = \"\n%+v\"\ngot = \"\n%+v\"",
			s.wantWarnings,
			got,
		)
		t.Fail()
		return
	}
	for i, w := range s.wantWarnings {
		if !reflect.DeepEqual(w.Keys, got[i].Keys) {
			t.Errorf(
				"warning keys do not match for warning %v: \nwant = \"\n%+v\"\ngot = \"\n%+v\"",
				i, w.Keys, got[i].Keys,
			)
			t.Fail()
		}
		if !reflect.DeepEqual(w.Warning, got[i].Warning) {
			t.Errorf(
				"warnings do not match for warning %v: \nwant = \"\n%+v\"\ngot = \"\n%+v\"",
				i, w.Warning, got[i].Warning,
			)
			t.Fail()
		}
	}
}
