package generate

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

type scenarioMergeSort struct {
	title              string
	from               []byte
	into               []byte
	persistenceComment string
	want               resMergeSort
}

type resMergeSort struct {
	res      []byte
	warnings []yamlfile.Warning
	err      error
}

var scenariosMergeSort = []scenarioMergeSort{
	{
		title: "simple example",
		from: []byte(`
k0: o0
k1: o1
a: !stay o1
`),
		into: []byte(`
k0: n0
b: n2
`),
		persistenceComment: "stay",
		want: resMergeSort{
			res: []byte(strings.TrimLeft(`
a: !stay o1
b: n2
k0: n0
`,
				"\n")),
			warnings: []yamlfile.Warning{},
			err:      nil,
		},
	},
	{
		title: "arrays with different length",
		from: []byte(`
arr:
  - o2
  - o1
  - !stay o3
`),
		into: []byte(`
arr:
  - n2
  - n1
`),
		persistenceComment: "stay",
		want: resMergeSort{
			res: []byte(strings.TrimLeft(`
arr:
  - n2
  - n1
  - !stay o3
`,
				"\n")),
			warnings: []yamlfile.Warning{
				{
					Keys:    []string{"arr"},
					Warning: "sequence length from (3) does not match length into (2)",
				},
			},
			err: nil,
		},
	},
	{
		title: "combined example",
		from: []byte(`
k0: o0
k1:
  k10: !stay o10
  k11: o11
a: !stay o1
arr:
  - o2
  - o1
`),
		into: []byte(`
k0: n0
k1:
  k11: n1
b: n2
arr:
  - n2
  - n1
`),
		persistenceComment: "stay",
		want: resMergeSort{
			res: []byte(strings.TrimLeft(`
a: !stay o1
arr:
  - n2
  - n1
b: n2
k0: n0
k1:
  k10: !stay o10
  k11: n1
`,
				"\n")),
			warnings: []yamlfile.Warning{},
			err:      nil,
		},
	},
	{
		title:              "faulty input 1",
		from:               []byte(`k0: {`),
		into:               []byte(``),
		persistenceComment: "",
		want: resMergeSort{
			res:      []byte(""),
			warnings: []yamlfile.Warning{},
			err:      errors.New("unmarshalling failed yaml: line 1: did not find expected node content"),
		},
	},
	{
		title:              "faulty input 1",
		from:               []byte(``),
		into:               []byte(`k0: {`),
		persistenceComment: "",
		want: resMergeSort{
			res:      []byte(""),
			warnings: []yamlfile.Warning{},
			err:      errors.New("unmarshalling failed yaml: line 1: did not find expected node content"),
		},
	},
	{
		title:              "map as key",
		from:               []byte(`k: v`),
		into:               []byte(`{map: as Key}: will not work`),
		persistenceComment: "",
		want: resMergeSort{
			res:      []byte(""),
			warnings: []yamlfile.Warning{},
			err:      errors.New("merge yamlfile error: merge for non-scalar map keys is not implemented"),
		},
	},
}

func TestMergeSort(te *testing.T) {
	for _, s := range scenariosMergeSort {
		te.Logf("test scenario: %s\n", s.title)

		res, warnings, err := mergeSort(s.from, s.into, s.persistenceComment)
		testfuncs.CheckErrs(te, s.want.err, err)
		s.want.CheckRes(te, res)
		s.want.CheckWarnings(te, warnings)
	}
}

func (r resMergeSort) CheckRes(te *testing.T, res []byte) {
	if !bytes.Equal(r.res, res) {
		te.Errorf(
			"results do not match: \nwant = \"\n%+v\n\"\ngot = \"\n%+v\n\"",
			string(r.res),
			string(res),
		)
		te.Fail()
	}
}

func (r resMergeSort) CheckWarnings(te *testing.T, warn []yamlfile.Warning) {
	if len(r.warnings) != len(warn) {
		te.Errorf(
			"warnings do not match: \nwant = \"%+v\"\ngot  = \"%+v\"",
			r.warnings,
			warn,
		)
		te.Fail()
		return
	}
	for i, w := range warn {
		if !reflect.DeepEqual(w.Keys, r.warnings[i].Keys) {
			fmt.Println("array compare")
			te.Errorf(
				"warning keys do not match for %v: \nwant = \"%+v\"\ngot  = \"%+v\"",
				i,
				r.warnings[i].Keys,
				w.Keys,
			)
			te.Fail()
		}
		if strings.Compare(w.Warning, r.warnings[i].Warning) != 0 {
			te.Errorf(
				"warning does not match for %v: \nwant = \"%+v\"\ngot  = \"%+v\"",
				i,
				r.warnings[i].Warning,
				w.Warning,
			)
			te.Fail()
		}
	}
}
