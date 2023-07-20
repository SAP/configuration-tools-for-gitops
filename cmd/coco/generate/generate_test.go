package generate

import (
	"errors"
	"fmt"
	"reflect"
	"sync"
	"testing"

	// The Generate function makes use of the general CLI logger. Hence its test
	// needs to set it up correctly to test logging output as well.
	"github.com/SAP/configuration-tools-for-gitops/pkg/log"
	"github.com/SAP/configuration-tools-for-gitops/pkg/testfuncs"
	"github.com/SAP/configuration-tools-for-gitops/pkg/version"
	"go.uber.org/zap"
)

type scenarioGenerate struct {
	title          string
	tmplIdentifier string
	configFileName string
	valueFilters   []string
	envFilters     []string
	folderFilters  []string
	exclFilters    []string
	logs           []logItem
	want           map[string]want
	wantErr        error
	templates      map[string][]byte
	values         map[string][]byte
}

type want struct {
	tmpls []template
	vals  map[string]interface{}
}

var scenariosGenerate = []scenarioGenerate{
	{
		title:          "simple test",
		tmplIdentifier: ".tmpl",
		configFileName: "coco.yaml",
		valueFilters:   []string{"values"},
		envFilters:     []string{},
		folderFilters:  []string{},
		templates: map[string][]byte{
			"services/a/.tmpl": []byte(`
key: {{.value1}}
constant: const-value
{{- if eq .ifKey "parse" }}
parse:
  conditional: {{.value2}} # inline comment
{{- end }}
`),
		},
		values: map[string][]byte{
			"values/c1/coco.yaml": []byte(`
type: environment
name: c1
values:
  - v1
`),
			"values/c2/coco.yaml": []byte(`
type: environment
name: c2
values:
  - v1
`),
			"values/c1/v1.yaml": []byte(`
value1: v1
value2: v2
ifKey: parse
`),
			"values/c2/v1.yaml": []byte(`value1: v22`),
		},
		logs: []logItem{},
		want: map[string]want{
			"services/a": {
				tmpls: []template{
					{
						source:     "services/a/.tmpl",
						basepath:   "services/a",
						namePrefix: "",
						subpath:    "",
					},
				},
				vals: map[string]interface{}{
					"c1": map[string]interface{}{
						"ifKey": "parse", "value1": "v1", "value2": "v2",
					},
					"c2": map[string]interface{}{
						"value1": "v22",
					},
				},
			},
		},
		wantErr: nil,
	},
	{
		title:          "test error",
		tmplIdentifier: ".tmpl",
		configFileName: "coco.yaml",
		valueFilters:   []string{"values"},
		envFilters:     []string{},
		folderFilters:  []string{},
		templates: map[string][]byte{
			"services/a/.tmpl": []byte(`key: {`),
		},
		values: map[string][]byte{
			"values/coco.yaml": []byte(`
type: environment
name: values
values: 
  - c1
`),
			"values/c1.yaml": []byte(`value1: fromValues-1`),
		},
		logs: []logItem{
			{
				Msg:     "report message",
				Level:   log.New("Error"),
				Context: map[string]interface{}{},
			},
		},
		want: map[string]want{
			"services/a": {
				tmpls: []template{
					{
						source:     "services/a/.tmpl",
						basepath:   "services/a",
						namePrefix: "",
						subpath:    "",
					},
				},
				vals: map[string]interface{}{
					"values": map[string]interface{}{"value1": "fromValues-1"},
				},
			},
		},
		wantErr: errors.New("1 rendering errors encountered"),
	},
	{
		title:          "test template does not exist error",
		tmplIdentifier: ".tmpl",
		configFileName: "coco.yaml",
		valueFilters:   []string{"values"},
		envFilters:     []string{},
		folderFilters:  []string{},
		templates: map[string][]byte{
			"services/a/.tmpl": []byte(`key: {`),
		},
		values: map[string][]byte{
			"values/coco.yaml": []byte(`
type: environment
values:
  - c1`),
			"values/c1.yaml": []byte(`value1: fromValues-1`),
		},
		logs:    []logItem{},
		want:    map[string]want{},
		wantErr: errors.New("lstat : no such file or directory"),
	},
	{
		title:          "template folder",
		tmplIdentifier: ".tmpl",
		configFileName: "coco.yaml",
		valueFilters:   []string{"values"},
		envFilters:     nil,
		folderFilters:  nil,
		exclFilters:    nil,
		logs:           nil,
		want: map[string]want{
			"services/a": {
				tmpls: []template{
					{
						source:     "services/a/.tmpl/template1.yaml",
						basepath:   "services/a",
						namePrefix: "",
						subpath:    "/template1.yaml",
					},
					{
						source:     "services/a/.tmpl/template2.yaml",
						basepath:   "services/a",
						namePrefix: "",
						subpath:    "/template2.yaml",
					},
				},
				vals: map[string]interface{}{
					"env1": map[string]interface{}{
						"test":  "v1",
						"test2": "v2",
					},
				},
			},
		},
		wantErr: nil,
		templates: map[string][]byte{
			"services/a/.tmpl/template1.yaml": []byte(`
test: {{.test}}`),
			"services/a/.tmpl/template2.yaml": []byte(`
test2: {{.test2}}`),
		},
		values: map[string][]byte{
			"services/a/values/coco.yaml": []byte(`
type: environment
name: env1
values: 
  - value1`),
			"services/a/values/value1.yaml": []byte(`
test: v1
test2: v2`),
		},
	},
}

func TestGenerate(t *testing.T) {
	if err := log.Init(log.Debug(), "", true); err != nil {
		zap.S().Fatal(err)
	}
	for _, s := range scenariosGenerate {
		t.Logf("test scenario: %s\n", s.title)
		s.Test(t)
	}
}

func (s *scenarioGenerate) Test(t *testing.T) {
	td, err := s.setupFiles()
	if err != nil {
		t.Logf("unable to create test dir tree: %v\n", err)
		t.FailNow()
	}
	defer td.Cleanup(t)
	tmpDir := td.Path()

	w := make(map[string]want, len(s.want))
	for k, v := range s.want {
		for i, t := range v.tmpls {
			v.tmpls[i].source = fmt.Sprintf("%s/%s", tmpDir, t.source)
			v.tmpls[i].basepath = fmt.Sprintf("%s/%s", tmpDir, t.basepath)
		}
		w[fmt.Sprintf("%s/%s", tmpDir, k)] = v
	}
	r := renderMock{
		report:     s.logs,
		t:          t,
		want:       w,
		foundNames: map[string]bool{},
	}
	if s.title == "test template does not exist error" {
		tmpDir = ""
	}

	renderer = r.render
	err = Generate(
		tmpDir,
		s.tmplIdentifier,
		"irrelevant",
		s.configFileName,
		&version.Version{},
		s.valueFilters,
		s.envFilters,
		s.folderFilters,
		s.exclFilters,
		log.New("Debug"),
		false,
	)
	testfuncs.CheckErrs(t, s.wantErr, err)

	for k := range s.want {
		if !r.found(fmt.Sprintf("%s/%s", tmpDir, k)) {
			t.Errorf("template name not found: \nwant = \"%+v\"", k)
			t.Fail()
		}
	}
}

func (s *scenarioGenerate) setupFiles() (testfuncs.TestDir, error) {
	genFiles := make(map[string][]byte, len(s.templates))
	for k, v := range s.templates {
		genFiles[k] = v
	}
	for k, v := range s.values {
		genFiles[k] = v
	}
	return testfuncs.PrepareTestDirTree(genFiles)
}

type renderMock struct {
	lock       sync.RWMutex
	report     []logItem
	t          *testing.T
	want       map[string]want
	foundNames map[string]bool
}

func (rm *renderMock) found(name string) bool {
	rm.lock.RLock()
	defer rm.lock.RUnlock()
	_, found := rm.foundNames[name]
	return found
}

func (rm *renderMock) render(
	name string, tmpls []template, vals map[string]interface{},
	reportChan chan<- renderReport,
	logLvl log.Level,
	persistenceComment string, v *version.Version,
	takeControl bool,
) {
	rm.lock.Lock()
	defer rm.lock.Unlock()
	reportChan <- renderReport{rm.report}
	want, ok := rm.want[name]
	if !ok {
		rm.t.Errorf("unknown template name found: \ngot = \"%+v\"", name)
		rm.t.Fail()
	}
	rm.foundNames[name] = true
	if !reflect.DeepEqual(want.vals, vals) {
		rm.t.Errorf(
			"values do not match: \nwant = \"%+v\"\ngot  = \"%+v\"",
			want.vals,
			vals,
		)
		rm.t.Fail()
	}
	if !reflect.DeepEqual(want.tmpls, tmpls) {
		rm.t.Errorf(
			"values do not match: \nwant = \"%+v\"\ngot  = \"%+v\"",
			want.tmpls,
			tmpls,
		)
		rm.t.Fail()
	}
}
