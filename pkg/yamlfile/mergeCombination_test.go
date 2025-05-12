package yamlfile

import (
	"fmt"
	"testing"

	"github.com/SAP/configuration-tools-for-gitops/v2/pkg/testfuncs"
	"gopkg.in/yaml.v3"
)

var mergeCombinations = []mergeCombinationTest{
	{
		title:   "scalar to scalar",
		from:    yaml.ScalarNode,
		into:    yaml.ScalarNode,
		want:    scalar2scalar,
		wantErr: nil,
	},
	{
		title:   "scalar to map",
		from:    yaml.ScalarNode,
		into:    yaml.MappingNode,
		want:    scalar2x,
		wantErr: nil,
	},
	{
		title:   "mapping to scalar",
		from:    yaml.MappingNode,
		into:    yaml.ScalarNode,
		want:    map2scalar,
		wantErr: nil,
	},
	{
		title:   "mapping to mapping",
		from:    yaml.MappingNode,
		into:    yaml.MappingNode,
		want:    map2map,
		wantErr: nil,
	},
	{
		title:   "mapping to sequence",
		from:    yaml.MappingNode,
		into:    yaml.SequenceNode,
		want:    map2sequence,
		wantErr: nil,
	},
	{
		title:   "mapping to x",
		from:    yaml.MappingNode,
		into:    yaml.DocumentNode,
		want:    scalar2scalar,
		wantErr: fmt.Errorf("merge combination from 4 (yaml.MappingNode) into 1 not supported"),
	},
	{
		title:   "sequence to scalar",
		from:    yaml.SequenceNode,
		into:    yaml.ScalarNode,
		want:    sequence2scalar,
		wantErr: nil,
	},
	{
		title:   "sequence to mapping",
		from:    yaml.SequenceNode,
		into:    yaml.MappingNode,
		want:    sequence2map,
		wantErr: nil,
	},
	{
		title:   "sequence to sequence",
		from:    yaml.SequenceNode,
		into:    yaml.SequenceNode,
		want:    sequence2sequence,
		wantErr: nil,
	},
	{
		title:   "sequence to x",
		from:    yaml.SequenceNode,
		into:    yaml.DocumentNode,
		want:    sequence2scalar,
		wantErr: fmt.Errorf("merge combination from 2 (yaml.SequenceNode) into 1 not supported"),
	},
	{
		title:   "document to document",
		from:    yaml.DocumentNode,
		into:    yaml.DocumentNode,
		want:    document2document,
		wantErr: nil,
	},
	{
		title:   "document to x",
		from:    yaml.DocumentNode,
		into:    yaml.ScalarNode,
		want:    scalar2scalar,
		wantErr: fmt.Errorf("merge combination from 1 (yaml.DocumentNode) into 8 not supported"),
	},
	{
		title:   "alias to x",
		from:    yaml.AliasNode,
		into:    yaml.ScalarNode,
		want:    scalar2scalar,
		wantErr: fmt.Errorf("merge combination from 16 (yaml.AliasNode) into 8 not supported"),
	},
	{
		title:   "x to x",
		from:    0,
		into:    0,
		want:    scalar2scalar,
		wantErr: fmt.Errorf("merge combination from 0 into 0 not supported"),
	},
}

func TestMergeCombination(t *testing.T) {
	// testLogger()
	for _, s := range mergeCombinations {
		t.Logf("test scenario: %s\n", s.title)

		got, err := mergeCombination(s.from, s.into)
		testfuncs.CheckErrs(t, s.wantErr, err)
		if err == nil {
			s.CheckRes(t, got)
		}
	}
}

type mergeCombinationTest struct {
	title   string
	from    yaml.Kind
	into    yaml.Kind
	want    mergeType
	wantErr error
}

func (m mergeCombinationTest) CheckRes(t *testing.T, got mergeType) {
	if m.want != got {
		t.Errorf(
			"results do not match: \nwant = \"\n%+v\"\ngot = \"\n%+v\"",
			m.want,
			got,
		)
		t.Fail()
	}
}
