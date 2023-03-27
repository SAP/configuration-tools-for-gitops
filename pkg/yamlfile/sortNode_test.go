package yamlfile_test

import (
	"bytes"
	"fmt"
	"strings"
	"testing"

	"github.com/configuration-tools-for-gitops/pkg/testfuncs"
	"github.com/configuration-tools-for-gitops/pkg/yamlfile"
)

var scenariosSort = []scenarioSort{
	{
		title: "simple example",
		input: []byte(strings.TrimSpace(`
z: v1
a: v2
`)),
		want: `
a: v2
z: v1
`,
	},
	{
		title: "example with arrays",
		input: []byte(strings.TrimSpace(`
z: v1
a: v2
b:
  - z
  - a
`)),
		want: `
a: v2
b:
  - z
  - a
z: v1
`,
	},
	{
		title: "nexted arrays",
		input: []byte(strings.TrimSpace(`
- z: v1
  a: v2
- a:
    z: v3
    a: v4
`)),
		want: `
- a: v2
  z: v1
- a:
    a: v4
    z: v3
`,
	},
}

func TestSort(t *testing.T) {
	for _, s := range scenariosSort {
		t.Logf("test scenario: %s\n", s.title)

		y, err := yamlfile.New(s.input)
		testfuncs.CheckErrs(t, nil, err)

		y.Sort()
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

type scenarioSort struct {
	title string
	input []byte
	want  string
}

func (s scenarioSort) CheckRes(t *testing.T, got string) bool {
	wantPrepared := fmt.Sprintf("%v\n", strings.TrimSpace(s.want))
	if wantPrepared != got {
		t.Errorf(
			"results do not match: \nwant = \"\n%+v\"\ngot = \"\n%+v\"",
			wantPrepared,
			got,
		)
		t.Fail()
		return false
	}
	return true
}
