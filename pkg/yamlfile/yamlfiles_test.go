package yamlfile_test

import (
	"bytes"
	"fmt"
	"strings"
	"testing"

	"github.com/SAP/configuration-tools-for-gitops/v2/pkg/log"
	"github.com/SAP/configuration-tools-for-gitops/v2/pkg/testfuncs"
	"github.com/SAP/configuration-tools-for-gitops/v2/pkg/yamlfile"
	"go.uber.org/zap"
	"gopkg.in/yaml.v3"
)

var scenarios = []scenarioPersistence{
	{
		title: "test",
		input: []byte(strings.TrimSpace(`
hello: !key world
a:
  - !key a1
  - a2 # key
  - a3
  - o1: !key k1
    o2: k2 # key
`)),
		filterByTagOrComment: "key",
		want: `hello: !key world
a:
  - !key a1
  - a2 # key
  - o1: !key k1
    o2: k2 # key
`,
		wantErr: nil,
	},
	{
		title: "keep slice",
		input: []byte(strings.TrimSpace(`
a: !key
  - a1
  - a2 # key
  - a3
b:
  - do not keep
  - this also not
  - c: and this neither
c:
  keep: !key
    - a
    - b
`)),
		filterByTagOrComment: "key",
		want: `a: !key
  - a1
  - a2 # key
  - a3
c:
  keep: !key
    - a
    - b
`,
		wantErr: nil,
	},

	{
		title: "simple example",
		input: []byte(strings.TrimSpace(`
key: removed
persistent: !humanOverwrite value
`)),
		filterByTagOrComment: "humanOverwrite",
		want: `persistent: !humanOverwrite value
`,
		wantErr: nil,
	},
	{
		title: "multiline example",
		input: []byte(strings.TrimSpace(`
key: removed
persistent: !humanOverwrite |
  line 1
  line 2
anotherKey: keepAsWell # humanOverwrite
`)),
		filterByTagOrComment: "humanOverwrite",
		want: `persistent: !humanOverwrite |
  line 1
  line 2
anotherKey: keepAsWell # humanOverwrite
`,
		wantErr: nil,
	},
	{
		title: "array alone",
		input: []byte(strings.TrimSpace(`
array:
  - v1
  - v2  # humanOverwrite
  - !humanOverwrite v3
  - v4  # humanOverwrite
`)),
		filterByTagOrComment: "humanOverwrite",
		want: `array:
  - v2 # humanOverwrite
  - !humanOverwrite v3
  - v4 # humanOverwrite
`,
		wantErr: nil,
	},
	{
		title:                "empty output",
		input:                []byte(strings.TrimSpace(`key: value`)),
		filterByTagOrComment: "human overwrite",
		want:                 ``,
		wantErr:              nil,
	},
	{
		title:                "empty input",
		input:                []byte{},
		filterByTagOrComment: "human overwrite",
		want:                 ``,
		wantErr:              nil,
	},
	{
		title: "nested map example",
		input: []byte(strings.TrimSpace(`
plain: value-not-persisted
notPersisted:
  nested: value
# comment
persisted0: persValue0 # human overwrite
persisted1:
  nested: persValue1  # human overwrite
  multi:
    2deep: value
array:
- v1
- v2  # human overwrite
- v3
- v4  # human overwrite
- ak: av # human overwrite
  bk: bv
`)),
		filterByTagOrComment: "human overwrite",
		want: `# comment
persisted0: persValue0 # human overwrite
persisted1:
  nested: persValue1 # human overwrite
array:
  - v2 # human overwrite
  - v4 # human overwrite
  - ak: av # human overwrite
`,
		wantErr: nil,
	},
	{
		title: "complex example",
		input: []byte(strings.TrimSpace(`
plain: value-not-persisted
notPersisted:
  nested: value
# comment
persisted0: persValue0 # human overwrite
persisted1:
  nested: persValue1  # human overwrite
  multi:
    2deep: value
    2deepPers: persValue2 # human overwrite
array:
- element0 # human overwrite
- element1
- element2 # human overwrite
nestedArray:
  array:
  - nestedEl0 # human overwrite
  - nestedEl1
  - - hello
    - keep # human overwrite
  - bla: blub
`)),
		filterByTagOrComment: "human overwrite",
		want: `# comment
persisted0: persValue0 # human overwrite
persisted1:
  nested: persValue1 # human overwrite
  multi:
    2deepPers: persValue2 # human overwrite
array:
  - element0 # human overwrite
  - element2 # human overwrite
nestedArray:
  array:
    - nestedEl0 # human overwrite
    - - keep # human overwrite
`,
		wantErr: nil,
	},
	{
		title:                "unmarshal fails",
		input:                []byte(strings.TrimSpace(`key: {`)),
		filterByTagOrComment: "human overwrite",
		want:                 "",
		wantErr:              fmt.Errorf("unmarshalling failed %s", "yaml: line 1: did not find expected node content"),
	},
	{
		title: "filtering fails",
		input: []byte(strings.TrimSpace(`
hello: &hello 'hello'
greeting:
  hello: *hello  #greeting.hello has the string value of 'hello'
`)),
		filterByTagOrComment: "human overwrite",
		want:                 "",
		wantErr:              fmt.Errorf("unmarshal yaml.AliasNode not implemented"),
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
		filterByKeys: []string{"keep", "only", "this", "part"},
		want: `keep:
  only:
    this:
      part:
        k2: v2
        k3: v3
        arr:
          - a1
          - a2
`,
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
		filterByKeys: []string{"keep", "only", "this", "part"},
		want:         ``,
		wantErr:      nil,
	},
}

func TestReadPersistentParts(t *testing.T) {
	testLogger()
	for _, s := range scenarios {
		t.Logf("test scenario: %s\n", s.title)

		y, err := yamlfile.New(s.input)
		if s.title == "unmarshal fails" {
			testfuncs.CheckErrs(t, s.wantErr, err)
			continue
		}
		testfuncs.CheckErrs(t, nil, err)

		if s.filterByTagOrComment != "" {
			err = y.FilterBy(s.filterByTagOrComment)
			testfuncs.CheckErrs(t, s.wantErr, err)
		}
		if len(s.filterByKeys) != 0 {
			err = y.FilterByKeys(s.filterByKeys)
			testfuncs.CheckErrs(t, s.wantErr, err)
		}
		if err == nil {
			var gotBytes bytes.Buffer
			err = y.Encode(&gotBytes, 2)
			testfuncs.CheckErrs(t, nil, err)
			if !s.CheckRes(t, gotBytes.String()) {
				scetchNodes(y.Node, []int{})
			}
		}
	}
}

type scenarioPersistence struct {
	title                string
	input                []byte
	filterByTagOrComment string
	filterByKeys         []string
	want                 string
	wantErr              error
}

func (s *scenarioPersistence) CheckRes(t *testing.T, got string) bool {
	if s.want != got {
		t.Errorf(
			"results do not match: \nwant = \"%+v\"\ngot = \"%+v\"",
			s.want,
			got,
		)
		t.Fail()
		return false
	}
	return true
}

func testLogger() {
	if err := log.Init(log.Debug(), "", true); err != nil {
		zap.S().Fatal(err)
	}
}

func scetchNodes(n *yaml.Node, parentLvls []int) {
	fmt.Printf("%+v: {kind: %+v, Value: %+v, Content: %+v, LineComment: %+v}\n",
		parentLvls,
		n.Kind, n.Value, n.Content, n.LineComment,
	)

	for i, e := range n.Content {
		scetchNodes(e, append(parentLvls, i))
	}
}
