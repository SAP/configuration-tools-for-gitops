package yamlfile_test

import (
	"bytes"
	"fmt"
	"reflect"
	"strings"
	"testing"

	"github.com/SAP/configuration-tools-for-gitops/v2/pkg/testfuncs"
	"github.com/SAP/configuration-tools-for-gitops/v2/pkg/yamlfile"
	"gopkg.in/yaml.v3"
)

var scenariosMergeNodes = []scenarioMergeNode{
	{
		title: "simple merge",
		from: []byte(strings.TrimSpace(`
key1: !HumanInput new-value
key3: new-value
`)),
		selectiveFlag: "HumanInput",
		into: []byte(strings.TrimSpace(`
key1: value
key2: value
`)),
		want: `
key1: !HumanInput new-value
key2: value
key3: new-value
`,
		wantSelective: `
key1: !HumanInput new-value
key2: value
`,
		wantErr: nil,
	},
	{
		title: "merge into empty",
		from: []byte(strings.TrimSpace(`
key1: !HumanInput new-value
key3: new-value
`)),
		selectiveFlag: "HumanInput",
		into:          []byte{},
		want: `
key1: !HumanInput new-value
key3: new-value
`,
		wantSelective: `
key1: !HumanInput new-value
`,
		wantErr: nil,
	},
	{
		title: "nested merge",
		from: []byte(strings.TrimSpace(`
k1:
  nested: !HumanInput new-value
k2:
  nested: !HumanInput new-value
  k2: remove
`)),
		selectiveFlag: "HumanInput",
		into: []byte(strings.TrimSpace(`
k1:
  nested: value
  stays: value
`)),
		want: `
k1:
  nested: !HumanInput new-value
  stays: value
k2:
  nested: !HumanInput new-value
  k2: remove
`,
		wantSelective: `
k1:
  nested: !HumanInput new-value
  stays: value
k2:
  nested: !HumanInput new-value
`,
		wantErr: nil,
	},
	{
		title: "array merge",
		from: []byte(strings.TrimSpace(`
k1:
- k1n1
- !HumanInput k1n2
k2:
  nested:
  - !HumanInput k2n1
  - k2n2
k3:
- k3n1
- !HumanInput k3n2
k4:
- k4n1
- !HumanInput k4n2
`)),
		selectiveFlag: "HumanInput",
		into: []byte(strings.TrimSpace(`
k1:
- k1o1
- k1o2
- k1o3
k2:
  nested:
  - k2o1
  stay: value
k3: []
k4:
- k4o1
- k4o2
`)),
		want: `
k1:
  - k1n1
  - !HumanInput k1n2
  - k1o3
k2:
  nested:
    - !HumanInput k2n1
    - k2n2
  stay: value
k3: [k3n1, !HumanInput k3n2]
k4:
  - k4n1
  - !HumanInput k4n2
`,
		wantSelective: `
k1:
  - k1o1
  - !HumanInput k1n2
  - k1o3
k2:
  nested:
    - !HumanInput k2n1
  stay: value
k3: [!HumanInput k3n2]
k4:
  - k4o1
  - !HumanInput k4n2
`,
		wantErr: nil,
		wantWarning: []yamlfile.Warning{
			{
				Keys:    []string{"k1"},
				Warning: "sequence length from (2) does not match length into (3)",
			},
			{
				Keys:    []string{"k2", "nested"},
				Warning: "sequence length from (2) does not match length into (1)",
			},
			{
				Keys:    []string{"k3"},
				Warning: "sequence length from (2) does not match length into (0)",
			},
		},
	},
	{
		title: "array object merge",
		from: []byte(strings.TrimSpace(`
- k3: !HumanInput NN3
  k4: NN4
- k3: !HumanInput NN5
  k7: !HumanInput NN7
  K8: NN8
- !HumanInput n
`)),
		selectiveFlag: "HumanInput",
		into: []byte(strings.TrimSpace(`
- k1: o1
  k2: o2
- k3: o3
  k4: o4
- k5: o5
  k65: o6
`)),
		want: `
- k1: o1
  k2: o2
  k3: !HumanInput NN3
  k4: NN4
- k3: !HumanInput NN5
  k4: o4
  k7: !HumanInput NN7
  K8: NN8
- !HumanInput n
`,
		wantSelective: `
- k1: o1
  k2: o2
  k3: !HumanInput NN3
- k3: !HumanInput NN5
  k4: o4
  k7: !HumanInput NN7
- !HumanInput n
`,
		wantErr: nil,
	},
	{
		title: "array in arrays merge",
		from: []byte(strings.TrimSpace(`
- - n1
  - !HumanInput n2
- - k1: !HumanInput n1
  - k2: !HumanInput n2
  - k3: n3
    k4: !HumanInput n4
`)),
		into: []byte(strings.TrimSpace(`
- - o1
  - o2
- - k1: o1
  - k22: o2
  - k3: o3
    k4: o4
`)),
		selectiveFlag: "HumanInput",
		want: `
- - n1
  - !HumanInput n2
- - k1: !HumanInput n1
  - k22: o2
    k2: !HumanInput n2
  - k3: n3
    k4: !HumanInput n4
`,
		wantSelective: `
- - o1
  - !HumanInput n2
- - k1: !HumanInput n1
  - k22: o2
    k2: !HumanInput n2
  - k3: o3
    k4: !HumanInput n4
`,
		wantErr: nil,
	},
	{
		title: "array nested object merge",
		from: []byte(strings.TrimSpace(`
- nested1: n1
- nested2: 
    - nn1: n1
    - nn2: !HumanInput n2
- nested3: !HumanInput 
    l2:
      - nn1: n1
      - nn2: n2
- nested4: !HumanInput 
    l21: n1
    l22: n2
`)),
		selectiveFlag: "HumanInput",
		into: []byte(strings.TrimSpace(`
- nested0: o0
- nested2:
    - nn1: o1
    - nn2: o2
- nested3:
    l2:
      - nn1: o1
      - nn2: o2
- nested4: 
    l21: o1
    l22: o2
`)),
		want: `
- nested0: o0
  nested1: n1
- nested2:
    - nn1: n1
    - nn2: !HumanInput n2
- nested3: !HumanInput
    l2:
      - nn1: n1
      - nn2: n2
- nested4: !HumanInput
    l21: n1
    l22: n2
`,
		wantSelective: `
- nested0: o0
- nested2:
    - nn1: o1
    - nn2: !HumanInput n2
- nested3: !HumanInput
    l2:
      - nn1: n1
      - nn2: n2
- nested4: !HumanInput
    l21: n1
    l22: n2
`,
		wantErr: nil,
	},
	{
		title: "array nested object merge 2",
		from: []byte(strings.TrimSpace(`
k1:
  - target: !HumanInput
      k1: n1
      k2: n2
    path: !HumanInput n11
`)),
		selectiveFlag: "HumanInput",
		into: []byte(strings.TrimSpace(`
k1:

`)),
		want: `
k1:
  - target: !HumanInput
      k1: n1
      k2: n2
    path: !HumanInput n11
`,
		wantSelective: `
k1:
  - target: !HumanInput
      k1: n1
      k2: n2
    path: !HumanInput n11
`,
		wantErr: nil,
		wantWarning: []yamlfile.Warning{
			{
				Keys:    []string{"k1"},
				Warning: "sequence length from (1) does not match length into (0)",
			},
		},
	},
	{
		title: "keep whole array",
		from: []byte(strings.TrimSpace(`
k1: !HumanInput
  - v1
  - v2
k2:
  - v1
  - v2
`)),
		selectiveFlag: "HumanInput",
		into: []byte(strings.TrimSpace(`
k1:
  - overwritten
  - overwritten2
k2:
  - newValue1
  - newValue2
`)),
		want: `
k1: !HumanInput
  - v1
  - v2
k2:
  - v1
  - v2
`,
		wantSelective: `
k1: !HumanInput
  - v1
  - v2
k2:
  - newValue1
  - newValue2
`,
		wantErr: nil,
	},
	{
		title: "merge map into scalar and vise versa",
		from: []byte(strings.TrimSpace(`
key1:
  nested: !HumanInput n1 # comment
key2: !HumanInput n2
`)),
		selectiveFlag: "HumanInput",
		into: []byte(strings.TrimSpace(`
key1: o1
key2:
  nested: o2
key3: o3
`)),
		want: `
key1:
  nested: !HumanInput n1 # comment
key2: !HumanInput n2
key3: o3
`,
		wantSelective: `
key1:
  nested: !HumanInput n1 # comment
key2: !HumanInput n2
key3: o3
`,
		wantErr: nil,
	},
	{
		title: "merge non existent submap",
		from: []byte(strings.TrimSpace(`
k: 
  nested: !HumanInput n1
`)),
		selectiveFlag: "HumanInput",
		into: []byte(strings.TrimSpace(`
v: 
  exists: o1
`)),
		want: `
v:
  exists: o1
k:
  nested: !HumanInput n1
`,
		wantSelective: `
v:
  exists: o1
k:
  nested: !HumanInput n1
`,
		wantErr: nil,
	},
	{
		title: "non-scalar map key",
		from: []byte(strings.TrimSpace(`
k: new-value
{k: v}: new-value
`)),
		into: []byte(strings.TrimSpace(`
k: value
`)),
		want:    ``,
		wantErr: fmt.Errorf("merge for non-scalar map keys is not implemented"),
	},
}

func TestMergeNodes(t *testing.T) {
	for _, s := range scenariosMergeNodes {
		t.Logf("test scenario: %s\n", s.title)

		into, err := yamlfile.New(s.into)
		testfuncs.CheckErrs(t, nil, err)
		from, err := yamlfile.New(s.from)
		testfuncs.CheckErrs(t, nil, err)
		warnings, err := into.Merge(from)
		testfuncs.CheckErrs(t, s.wantErr, err)
		s.CheckWarnings(t, warnings)
		if err == nil {
			s.CheckRes(t, into, s.want)
		}
		if t.Failed() {
			scetchNodes(from.Node, []int{})
		}
	}
}

func TestMergeBytesNodes(t *testing.T) {
	for _, s := range append(
		scenariosMergeNodes,
		scenarioMergeNode{
			title:   "unmarshal fails",
			from:    []byte(strings.TrimSpace(`key: {`)),
			into:    []byte(strings.TrimSpace(`key: value`)),
			want:    "",
			wantErr: fmt.Errorf("unmarshalling failed %s", "yaml: line 1: did not find expected node content"),
		},
	) {
		t.Logf("test scenario: %s\n", s.title)

		into, err := yamlfile.New(s.into)
		testfuncs.CheckErrs(t, nil, err)
		warnings, err := into.MergeBytes(s.from)
		testfuncs.CheckErrs(t, s.wantErr, err)
		s.CheckWarnings(t, warnings)
		if err == nil {
			s.CheckRes(t, into, s.want)
		}
		if t.Failed() {
			scetchNodes(into.Node, []int{})
		}
	}
}

func TestMergeSelective(t *testing.T) {
	for _, s := range scenariosMergeNodes {
		t.Logf("test scenario: %s\n", s.title)

		into, err := yamlfile.New(s.into)
		testfuncs.CheckErrs(t, nil, err)
		from, err := yamlfile.New(s.from)
		testfuncs.CheckErrs(t, nil, err)
		warnings, err := into.MergeSelective(from, s.selectiveFlag)
		testfuncs.CheckErrs(t, s.wantErr, err)
		s.CheckWarnings(t, warnings)
		if err == nil {
			s.CheckRes(t, into, s.wantSelective)
		}
		if t.Failed() {
			scetchNodes(from.Node, []int{})
			fmt.Println("----------------------")
			scetchNodes(into.Node, []int{})
		}
	}
}

type scenarioMergeNode struct {
	title         string
	from          []byte
	into          []byte
	selectiveFlag string
	want          string
	wantSelective string
	wantErr       error
	wantWarning   []yamlfile.Warning
}

func (s *scenarioMergeNode) CheckRes(t *testing.T, got yamlfile.Yaml, want string) {
	var gotBytes bytes.Buffer
	e := yaml.NewEncoder(&gotBytes)
	e.SetIndent(2)
	err := e.Encode(got.Node)
	if err != nil {
		t.Errorf("could not encode result: %v", err)
		t.FailNow()
	}
	e.Close()
	wantPrepared := fmt.Sprintf("%v\n", strings.TrimSpace(want))
	if wantPrepared != gotBytes.String() {
		t.Errorf(
			"results do not match: \nwant = \"\n%+v\"\ngot = \"\n%+v\"",
			wantPrepared,
			gotBytes.String(),
		)
		t.Fail()
	}
}

func (s *scenarioMergeNode) CheckWarnings(t *testing.T, got []yamlfile.Warning) {
	if len(s.wantWarning) != len(got) {
		t.Errorf(
			"warnings do not match: \nwant = \"%+v\"\ngot = \"%+v\"",
			s.wantWarning,
			got,
		)
		t.Fail()
	}
	for i, want := range s.wantWarning {
		if !reflect.DeepEqual(want, got[i]) {
			t.Errorf(
				"warnings do not match: \nwant = \"%+v\"\ngot = \"%+v\"",
				s.wantWarning,
				got,
			)
			t.Fail()
		}
	}
}
