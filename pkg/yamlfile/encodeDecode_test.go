package yamlfile_test

import (
	"bytes"
	"reflect"
	"strings"
	"testing"

	"github.com/configuration-tools-for-gitops/pkg/testfuncs"
	"github.com/configuration-tools-for-gitops/pkg/yamlfile"
)

type scenarioDecodeEncode struct {
	title      string
	input      []byte
	wantDecode map[string]interface{}
	wantEncode string
	wantErr    error
}

var scenariosDecodeEncode = []scenarioDecodeEncode{
	{
		title: "array alone",
		input: []byte(strings.TrimSpace(`
array:
  - v1
  - v2  # human overwrite
  - v3
`)),
		wantDecode: map[string]interface{}{
			"array": []interface{}{"v1", "v2", "v3"},
		},
		wantEncode: `array:
  - v1
  - v2 # human overwrite
  - v3
`,
		wantErr: nil,
	},
	{
		title:      "empty output",
		input:      []byte{},
		wantDecode: map[string]interface{}{},
		wantEncode: ``,
		wantErr:    nil,
	},
	{
		title: "nested map example",
		input: []byte(strings.TrimSpace(`
k: v
k2:
  nested: value
# comment 
array:
- v1
- v2  # human overwrite
- v3
- ak: av # human overwrite
  bk: bv
`)),
		wantDecode: map[string]interface{}{
			"array": []interface{}{
				"v1",
				"v2",
				"v3",
				map[string]interface{}{"ak": "av", "bk": "bv"},
			},
			"k":  "v",
			"k2": map[string]interface{}{"nested": "value"},
		},
		wantEncode: `k: v
k2:
  nested: value
# comment 
array:
  - v1
  - v2 # human overwrite
  - v3
  - ak: av # human overwrite
    bk: bv
`,
		wantErr: nil,
	},
}

func TestToMap(t *testing.T) {
	testLogger()
	for _, s := range scenariosDecodeEncode {
		t.Logf("test scenario: %s\n", s.title)

		y, err := yamlfile.New(s.input)
		testfuncs.CheckErrs(t, nil, err)

		var gotMap map[string]interface{}
		err = y.Decode(&gotMap)
		testfuncs.CheckErrs(t, s.wantErr, err)

		var gotBytes bytes.Buffer
		err = y.Encode(&gotBytes, 2)
		testfuncs.CheckErrs(t, s.wantErr, err)

		if err == nil {
			s.CheckRes(t, gotMap, gotBytes.Bytes())
		}
	}
}

func (s *scenarioDecodeEncode) CheckRes(
	t *testing.T, gotMap map[string]interface{}, gotBytes []byte,
) {
	checkDecode(t, s.wantDecode, gotMap)
	checkEncode(t, s.wantEncode, gotBytes)
}

func checkDecode(t *testing.T, want, got map[string]interface{}) {
	if len(want) == 0 && len(got) == 0 &&
		reflect.TypeOf(want) == reflect.TypeOf(got) {
		return
	}
	if !reflect.DeepEqual(want, got) {
		t.Errorf(
			"results do not match: \nwant = \"%+v\"\ngot  = \"%+v\"",
			want,
			got,
		)
		t.Fail()
	}
}

func checkEncode(t *testing.T, want string, got []byte) {
	if want == "" && len(got) == 0 {
		return
	}
	if want != string(got) {
		t.Errorf(
			"results do not match: \nwant = \"%+v\"\ngot = \"%+v\"",
			want,
			string(got),
		)
		t.Fail()
	}
}
