package generate

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"io/fs"
	"os"
	"path/filepath"
	"reflect"
	"regexp"
	"strconv"
	"strings"
	"testing"
	gotemplate "text/template"

	"github.com/configuration-tools-for-gitops/pkg/log"
	"github.com/configuration-tools-for-gitops/pkg/testfuncs"
	"github.com/configuration-tools-for-gitops/pkg/version"
	"github.com/configuration-tools-for-gitops/pkg/yamlfile"
	"gopkg.in/yaml.v3"
)

var selectTest = ""

type scenarioRender struct {
	title string
	i     renderInput
	m     mock
	o     renderOutput
}

type renderInput struct {
	templates          []template
	templateContent    [][]byte
	alreadyPresent     map[string][]byte
	values             map[string][]byte
	persistenceComment string
	version            string
	takeControl        bool
}

type renderOutput struct {
	want       map[string][]byte
	wantReport []logItem
}

var scenariosRender = []scenarioRender{
	{
		title: "test generation of correct filenames",
		m: mock{
			mockMergeSort: true,
		},
		i: renderInput{
			templates:       []template{{"path/X/.tmpl", "path/X", "", ""}},
			templateContent: [][]byte{content(``)},
			alreadyPresent:  map[string][]byte{},
			values: map[string][]byte{
				"c1": content(``),
				"c2": content(``),
			},
			version: "99.99.99",
		},
		o: renderOutput{
			want: map[string][]byte{
				"path/X/c1.yaml": defaultFileContent,
				"path/X/c2.yaml": defaultFileContent,
			},
		},
	},
	{
		title: "test template rendering",
		i: renderInput{
			templates: []template{{"path/.tmpl", "path", "", ""}},
			templateContent: [][]byte{content(`
constant: const-value
key: !yamlFlag {{.value1}}
{{- if eq .ifKey "parse" }}
parse:
	conditional: {{.value2}} # inline comment
{{- end }}
testTrim: {{ trimSuffix "hello world" " world" }}
testJoin: {{ join "-" "hello" "world" "2" }}
testQuote: {{ "hello" | quote }}
testSplitSelect: {{ split "hello-world-2" "-" | select 1 }}
testSplitSelectEmpty: {{ split "hello-world-2" "-" | select 10 | quote }}
	`)},
			values: map[string][]byte{
				"c1": content(`
value1: fromValues-1
value2: fromValues-2
ifKey: parse
`),
			},
			version: "99.99.99",
		},
		m: mock{
			mockMergeSort: true,
			w: wantInput{
				into: content(`
constant: const-value
key: !yamlFlag fromValues-1
parse:
	conditional: fromValues-2 # inline comment
testTrim: hello
testJoin: hello-world-2
testQuote: "hello"
testSplitSelect: world
testSplitSelectEmpty: ""
`),
			},
		},
		o: renderOutput{
			want: map[string][]byte{
				"path/c1.yaml": defaultFileContent,
			},
		},
	},
	{
		title: "test non-yaml rendering",
		i: renderInput{
			templates: []template{{"path/.tmpl/nonYamlFile", "path", "", "nonYamlFile"}},
			templateContent: [][]byte{content(`
VAR_1={{ join "-" "hello" "world" "2" }}
VAR_2={{ .value1 }}
	`)},
			values: map[string][]byte{
				"c1": content(`
value1: fromValues-1
`),
			},
			version: "99.99.99",
		},
		o: renderOutput{
			want: map[string][]byte{
				"path/c1/nonYamlFile": content(`
# Code generated by CLI 'coco generate ...' (version: 99.99); DO NOT EDIT.

VAR_1=hello-world-2
VAR_2=fromValues-1
`),
			},
		},
	},
	{
		title: "template parsing fails",
		i: renderInput{
			templates:       []template{{"path/.tmpl", "path", "", ""}},
			templateContent: [][]byte{content(`fail: {{ doesNotExist }}`)},
			values:          map[string][]byte{"c1": content(``)},
			version:         "99.99.99",
		},
		m: mock{mockMergeSort: true},
		o: renderOutput{
			wantReport: []logItem{
				{
					Msg:   `template: {{.TmpDir}}/path/.tmpl:1: function "doesNotExist" not defined`,
					Level: log.Error(),
					Context: map[string]interface{}{
						"error":      `template: {{.TmpDir}}/path/.tmpl:1: function "doesNotExist" not defined`,
						"go-routine": "template parsing fails",
						"template":   `{{.TmpDir}}/path/.tmpl`,
					},
				},
			},
		},
	},
	{
		title: "template rendering fails",
		i: renderInput{
			templates:       []template{{"path/.tmpl", "path", "", ""}},
			templateContent: [][]byte{content(``)},
			values:          map[string][]byte{"c1": content(``)},
			version:         "99.99.99",
		},
		m: mock{
			mockMergeSort: true,
			mockParser:    true,
			o:             mockOutput{err: errors.New("rendering error")},
		},
		o: renderOutput{
			wantReport: []logItem{
				{
					Msg:   `rendering error`,
					Level: log.Error(),
					Context: map[string]interface{}{
						"error":      `rendering error`,
						"file":       `{{.TmpDir}}/path/c1.yaml`,
						"go-routine": "template rendering fails",
						"template":   `{{.TmpDir}}/path/.tmpl`,
						"values":     "",
					},
				},
			},
		},
	},
	{
		title: "test warnings",
		i: renderInput{
			templates:          []template{{"path/.tmpl", "path", "", ""}},
			templateContent:    [][]byte{content(``)},
			values:             map[string][]byte{"c1": content(``)},
			persistenceComment: "",
			version:            "99.99.99",
		},
		m: mock{
			mockMergeSort: true,
			o: mockOutput{
				warnings: []yamlfile.Warning{
					{
						Keys:    []string{"k1"},
						Warning: "first warning",
					},
					{
						Keys:    []string{"k1", "k2"},
						Warning: "second warning",
					},
				},
				err: nil,
			},
		},
		o: renderOutput{
			wantReport: []logItem{
				{
					Msg:   "first warning",
					Level: log.Warn(),
					Context: map[string]interface{}{
						"file":       `{{.TmpDir}}/path/c1.yaml`,
						"go-routine": "test warnings",
						"keys":       []string{"k1"},
						"template":   `{{.TmpDir}}/path/.tmpl`,
						"values":     "",
					},
				},
				{
					Msg:   "second warning",
					Level: log.Warn(),
					Context: map[string]interface{}{
						"file":       `{{.TmpDir}}/path/c1.yaml`,
						"go-routine": "test warnings",
						"keys":       []string{"k1", "k2"},
						"template":   `{{.TmpDir}}/path/.tmpl`,
						"values":     "",
					},
				},
			},
		},
	},
	{
		title: "yamlProcessor fails",
		i: renderInput{
			templates:          []template{{"path/.tmpl", "path", "", ""}},
			templateContent:    [][]byte{content(``)},
			values:             map[string][]byte{"c1": content(``)},
			persistenceComment: "",
			version:            "99.99.99",
		},
		m: mock{
			mockMergeSort: true,
			o: mockOutput{
				err: errors.New("intended failure"),
			},
		},
		o: renderOutput{
			wantReport: []logItem{
				{
					Msg:   "intended failure",
					Level: log.Error(),
					Context: map[string]interface{}{
						"error":      "intended failure",
						"file":       `{{.TmpDir}}/path/c1.yaml`,
						"go-routine": "yamlProcessor fails",
						"template":   `{{.TmpDir}}/path/.tmpl`,
						"values":     "",
					},
				},
			},
		},
	},
	{
		title: "compatible versions",
		i: renderInput{
			templates:       []template{{"path/.tmpl", "path", "", ""}},
			templateContent: [][]byte{content(`minimal: template`)},
			values:          map[string][]byte{"c1": content(``)},
			version:         "99.99.99",
			takeControl:     false,
			alreadyPresent: map[string][]byte{
				"path/c1.yaml": content(fmt.Sprintf(genFileHeader, "99", "98")),
			},
		},
		m: mock{mockMergeSort: true},
		o: renderOutput{
			want: map[string][]byte{
				"path/c1.yaml": content(fmt.Sprint(
					fmt.Sprintf(genFileHeader, "99", "99"),
					"mocked mergeSort",
				)),
			},
		},
	},
	{
		title: "no change -> no updated version string",
		i: renderInput{
			templates:       []template{{"path/.tmpl", "path", "", ""}},
			templateContent: [][]byte{content("mocked mergeSort")},
			values:          map[string][]byte{"c1": content(``)},
			version:         "99.99.99",
			takeControl:     false,
			alreadyPresent: map[string][]byte{
				"path/c1.yaml": content(fmt.Sprint(
					fmt.Sprintf(genFileHeader, "99", "98"),
					"mocked mergeSort",
				)),
			},
		},
		m: mock{mockMergeSort: true},
		o: renderOutput{
			want: map[string][]byte{
				"path/c1.yaml": content(fmt.Sprint(
					fmt.Sprintf(genFileHeader, "99", "98"),
					"mocked mergeSort",
				)),
			},
		},
	},
	{
		title: "skip incompatible versions - general",
		i: renderInput{
			templates:       []template{{"path/.tmpl", "path", "", ""}},
			templateContent: [][]byte{content(`minimal: template`)},
			values:          map[string][]byte{"c1": content(``)},
			version:         "99.99.99",
			takeControl:     false,
			alreadyPresent: map[string][]byte{
				"path/c1.yaml": content("# this does not match"),
			},
		},
		m: mock{mockMergeSort: true},
		o: renderOutput{
			want: map[string][]byte{
				"path/c1.yaml": content("# this does not match"),
			},
		},
	},
	{
		title: "skip incompatible versions - major",
		i: renderInput{
			templates:       []template{{"path/.tmpl", "path", "", ""}},
			templateContent: [][]byte{content(`minimal: template`)},
			values:          map[string][]byte{"c1": content(``)},
			version:         "99.99.99",
			takeControl:     false,
			alreadyPresent: map[string][]byte{
				"path/c1.yaml": content(fmt.Sprintf(genFileHeader, "1", "99")),
			},
		},
		m: mock{mockMergeSort: true},
		o: renderOutput{
			want: map[string][]byte{
				"path/c1.yaml": content(fmt.Sprintf(genFileHeader, "1", "99")),
			},
		},
	},
	{
		title: "skip incompatible versions - minor",
		i: renderInput{
			templates:       []template{{"path/.tmpl", "path", "", ""}},
			templateContent: [][]byte{content(`minimal: template`)},
			values:          map[string][]byte{"c1": content(``)},
			version:         "99.1.99",
			takeControl:     false,
			alreadyPresent: map[string][]byte{
				"path/c1.yaml": content(fmt.Sprintf(genFileHeader, "99", "99")),
			},
		},
		m: mock{mockMergeSort: true},
		o: renderOutput{
			want: map[string][]byte{
				"path/c1.yaml": content(fmt.Sprintf(genFileHeader, "99", "99")),
			},
		},
	},
	{
		title: "hard overwrite incompatible versions",
		i: renderInput{
			templates:       []template{{"path/.tmpl", "path", "", ""}},
			templateContent: [][]byte{content(`minimal: template`)},
			values:          map[string][]byte{"c1": content(``)},
			version:         "99.99.99",
			takeControl:     true,
			alreadyPresent: map[string][]byte{
				"path/c1.yaml": content(fmt.Sprintf(genFileHeader, "1", "1")),
			},
		},
		m: mock{mockMergeSort: true},
		o: renderOutput{
			want: map[string][]byte{
				"path/c1.yaml": content(fmt.Sprint(
					fmt.Sprintf(genFileHeader, "99", "99"),
					"mocked mergeSort",
				)),
			},
		},
	},

	{
		title: "e2e example",
		i: renderInput{
			templates: []template{{"path/X/.tmpl", "path/X", "", ""}},
			templateContent: [][]byte{content(`
constant: const-value
array: [{{.value3}}]
key: {{.value1}}
array2:
  - default
nested: {{.nested.key}}
{{- if eq .ifKey "parse" }}
parse:
  conditional: {{.value2}} # inline comment
{{- end }}
	`,
			)},
			alreadyPresent: map[string][]byte{"path/X/c1.yaml": content(`
# Code generated by CLI 'coco generate ...' (version: 99.99); DO NOT EDIT.

array2:
  - !HumanOverwrite persistent-1 
  - persistent-2 # HumanOverwrite
key1: oldValue
key2: !HumanOverwrite persistentValue
	`)},
			values: map[string][]byte{
				"c1": content(`
value1: fromValues-1
value2: fromValues-2
ifKey: parse
nested:
  key: fromValues-4
value3: fromValues-3
	`,
				),
				"c2": content(`value1: fromValues-1`),
			},
			persistenceComment: "HumanOverwrite",
			version:            "99.99.99",
			takeControl:        true,
		},
		o: renderOutput{
			want: map[string][]byte{
				"path/X/c1.yaml": content(`
# Code generated by CLI 'coco generate ...' (version: 99.99); DO NOT EDIT.

array: [fromValues-3]
array2:
  - !HumanOverwrite persistent-1
  - persistent-2 # HumanOverwrite
constant: const-value
key: fromValues-1
key2: !HumanOverwrite persistentValue
nested: fromValues-4
parse:
  conditional: fromValues-2 # inline comment
`),
				"path/X/c2.yaml": content(`
# Code generated by CLI 'coco generate ...' (version: 99.99); DO NOT EDIT.

array: [<no value>]
array2:
  - default
constant: const-value
key: fromValues-1
nested: <no value>
`),
			},
			wantReport: []logItem{
				{
					Msg:   "sequence length from (2) does not match length into (1)",
					Level: log.Warn(),
					Context: map[string]interface{}{
						"file":       `{{.TmpDir}}/path/X/c1.yaml`,
						"go-routine": "e2e example",
						"keys":       []string{"array2"},
						"template":   `{{.TmpDir}}/path/X/.tmpl`,
						"values": map[string]interface{}{
							"ifKey":  "parse",
							"nested": map[string]interface{}{"key": "fromValues-4"},
							"value1": "fromValues-1",
							"value2": "fromValues-2",
							"value3": "fromValues-3",
						},
					},
				},
			},
		},
	},
}

var (
	defaultFileContent = content(`
# Code generated by CLI 'coco generate ...' (version: 99.99); DO NOT EDIT.

mocked mergeSort
	`)
)

func TestRender(te *testing.T) {
	if err := log.Init(log.Debug(), "", true); err != nil {
		te.Logf("unable to initialize logger")
		te.FailNow()
	}

	for _, s := range scenariosRender {
		if selectTest != "" && s.title != selectTest {
			continue
		}
		te.Logf("test scenario: %s\n", s.title)

		s.Test(te)
	}
}

func (s *scenarioRender) Test(te *testing.T) {
	v, err := setVersion(s.i.version)
	if err != nil {
		te.Log(err)
		te.FailNow()
	}

	yamlProcessor = mergeSort
	if s.m.mockMergeSort {
		yamlProcessor = s.m.mergeSort
	}
	if s.m.mockParser {
		parserConfig = parserMock{
			Mock: true,
			Err:  s.m.o.err,
		}
	} else {
		parserConfig = parserMock{}
	}

	td, err := s.i.setupFiles()
	if err != nil {
		te.Logf("unable to create test dir tree: %v\n", err)
		te.FailNow()
	}
	defer td.Cleanup(te)
	tmpDir := td.Path()

	testTemplates := make([]template, len(s.i.templates))

	for i, t := range s.i.templates {
		testTemplates[i] = template{
			source:     filepath.Join(tmpDir, t.source),
			basepath:   filepath.Join(tmpDir, t.basepath),
			namePrefix: t.namePrefix,
			subpath:    t.subpath,
		}
	}

	valueFileContent := make(map[string]interface{}, len(s.i.values))
	for k, rawValues := range s.i.values {
		d := yaml.NewDecoder(bytes.NewReader(rawValues))
		var values interface{}
		err := d.Decode(&values)
		if err == io.EOF {
			values = ""
		} else if err != nil {
			te.Logf("unable to decode value file content: %v\n", err)
			te.FailNow()
		}
		valueFileContent[k] = values
	}

	report := make(chan renderReport, 1)

	render(
		s.title, testTemplates, valueFileContent, report,
		log.Debug(), s.i.persistenceComment,
		&v, s.i.takeControl,
	)
	rep := <-report

	s.o.CheckReport(te, rep, tmpDir)
	s.o.CheckRes(te, tmpDir)
	s.m.Check(te)
}

func setVersion(versionString string) (version.Version, error) {
	res := version.Version{Version: fmt.Sprintf("v%v", versionString)}
	re := regexp.MustCompile(`(\d*)\.(\d*)\.(\d*)`)
	v := re.FindStringSubmatch(versionString)
	if len(v) != 4 {
		return res, fmt.Errorf("illegal version \"%v\"", v)
	}
	var err error
	res.SemVer.Major, err = strconv.Atoi(v[1])
	if err != nil {
		return res, fmt.Errorf("illegal major version \"%v\"", v)
	}
	res.SemVer.Minor, err = strconv.Atoi(v[2])
	if err != nil {
		return res, fmt.Errorf("illegal major version \"%v\"", v)
	}
	res.SemVer.Patch, err = strconv.Atoi(v[3])
	if err != nil {
		return res, fmt.Errorf("illegal major version \"%v\"", v)
	}
	return res, nil
}

type mock struct {
	mockMergeSort bool
	mockParser    bool
	g             gotInput
	w             wantInput
	o             mockOutput
}

type gotInput struct {
	from, into  []byte
	persistence string
}
type wantInput struct {
	from, into []byte
}

func (m *mock) Check(t *testing.T) {
	if m.w.from != nil {
		if !bytes.Equal(m.w.from, m.g.from) {
			t.Errorf(
				"mergeSort input \"from\" does not match: \nwant = \"\n%+v\n\"\ngot  = \"\n%+v\n\"",
				string(m.w.from),
				string(m.g.from),
			)
		}
	}
	if m.w.into != nil {
		if !bytes.Equal(m.w.into, m.g.into) {
			t.Errorf(
				"mergeSort input \"into\" does not match: \nwant = \"\n%+v\n\"\ngot  = \"\n%+v\n\"",
				string(m.w.into),
				string(m.g.into),
			)
		}
	}
}

type mockOutput struct {
	warnings []yamlfile.Warning
	err      error
}

func (m *mock) mergeSort(from, into []byte, persistence string) (
	[]byte, []yamlfile.Warning, error,
) {
	if m.o.err != nil {
		return nil, m.o.warnings, m.o.err
	}
	m.g.from = from
	m.g.into = into
	m.g.persistence = persistence
	return []byte("mocked mergeSort\n"), m.o.warnings, nil
}

func (m *mock) parse(filename string, funcs gotemplate.FuncMap) error {
	return nil
}

func (m *mock) execute(data interface{}) ([]byte, error) {
	return nil, m.o.err
}

func (ri *renderInput) setupFiles() (testfuncs.TestDir, error) {
	genFiles := make(map[string][]byte, len(ri.templates))
	for i, t := range ri.templates {
		if len(ri.templateContent[i]) > 0 {
			genFiles[t.source] = ri.templateContent[i]
		}
	}

	for f, c := range ri.alreadyPresent {
		genFiles[f] = c
	}

	return testfuncs.PrepareTestDirTree(genFiles)
}

func (ro renderOutput) CheckRes(t *testing.T, basedir string) {
	failed := false
	for name, content := range ro.want {
		data, err := os.ReadFile(filepath.Join(basedir, name))
		if err != nil {
			t.Errorf("failed to read result file %s. Error: %s", data, err)
			failed = true
			break
		}
		if !bytes.Equal(content, data) {
			t.Errorf(
				"results do not match: \nwant = \"\n%+v\n\"\ngot  = \"\n%+v\n\"",
				string(content),
				string(data),
			)
			failed = true
			t.Fail()
		}
	}
	if failed {
		t.Fail()
		_ = filepath.WalkDir(basedir, func(path string, e fs.DirEntry, err error) error {
			fmt.Println(path)
			return nil
		})
	}
}

func (ro renderOutput) CheckReport(t *testing.T, r renderReport, tmpDir string) {
	if len(ro.wantReport) != len(r.items) {
		t.Errorf("unexpected report length: \nwant = \"%+v\"\ngot  = \"%+v\"",
			ro.wantReport, r.items,
		)
		t.FailNow()
	}
	wantReport := make([]logItem, 0, len(ro.wantReport))
	for _, r := range ro.wantReport {
		tmpl := gotemplate.Must(gotemplate.New("msg").Parse(r.Msg))
		msgBytes := new(bytes.Buffer)
		if err := tmpl.Execute(msgBytes, struct{ TmpDir string }{tmpDir}); err != nil {
			t.Errorf("failed to parse error template: %s", err)
			t.FailNow()
		}
		item := logItem{
			Msg:     msgBytes.String(),
			Level:   r.Level,
			Context: make(map[string]interface{}, len(r.Context)),
		}
		for k, v := range r.Context {
			if reflect.TypeOf(v).Kind() != reflect.String {
				item.Context[k] = v
				continue
			}
			tmpl := gotemplate.Must(gotemplate.New(k).Parse(v.(string)))
			valBytes := new(bytes.Buffer)
			if err := tmpl.Execute(valBytes, struct{ TmpDir string }{tmpDir}); err != nil {
				t.Errorf("failed to parse error template: %s", err)
				t.FailNow()
			}
			item.Context[k] = valBytes.String()
		}
		wantReport = append(wantReport, item)
	}

	compareReports(t, wantReport, r.items)
}

func compareReports(t *testing.T, wantReport, gotReports []logItem) {
	reportsEqual := true
	for i, want := range wantReport {
		got := gotReports[i]
		if want.Msg != got.Msg || want.Level != got.Level {
			reportsEqual = false
			break
		}
		for k, wantV := range want.Context {
			gotV, found := got.Context[k]
			if !found {
				t.Errorf("report context key not found: \nwant = \"%+v\"", k)
				t.Fail()
				reportsEqual = false
				break
			}
			if strings.Compare(fmt.Sprintf("%+v", gotV), fmt.Sprintf("%+v", wantV)) != 0 {
				t.Errorf("report context values do not match: \nwant = \"%+v\"\ngot  = \"%+v\"",
					wantV, gotV,
				)
				t.Fail()
				reportsEqual = false
				break
			}
		}
		for k := range got.Context {
			if _, found := want.Context[k]; !found {
				t.Errorf("report context key present but should not: \ngot = \"%+v\"", k)
				t.Fail()
				reportsEqual = false
				break
			}
		}
		if !reportsEqual {
			break
		}
	}
	if !reportsEqual {
		t.Errorf(
			"reports do not match: \nwant = \"%+v\"\ngot  = \"%+v\"",
			wantReport,
			gotReports,
		)
		t.Fail()
	}
}
func content(c string) []byte {
	return []byte(fmt.Sprintf("%s\n", strings.TrimSpace(c)))
}
