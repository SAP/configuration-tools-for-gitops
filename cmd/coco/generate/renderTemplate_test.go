package generate

import (
	"bytes"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/SAP/configuration-tools-for-gitops/v2/pkg/log"
	"github.com/SAP/configuration-tools-for-gitops/v2/pkg/testfuncs"
)

var selectParserTest = ""

type scenarioParserTemplate struct {
	title string
	i     parseInput
	o     expectedOutput
}

type parseInput struct {
	templateFileName string
	templateContent  []byte
	valueFiles       map[string][]byte
	target           string
}

type expectedOutput struct {
	err     error
	content []byte
}

var scenariosParserTemplate = []scenarioParserTemplate{
	{
		title: "simple example",
		i: parseInput{
			templateFileName: "example",
			templateContent: []byte(strings.TrimSpace(`
{{.key1}}
{{.key2}}
`)),
			valueFiles: map[string][]byte{
				"values.yaml": []byte(strings.TrimSpace(`
key1: value1
key2: value2
`)),
			},
			target: "output",
		},
		o: expectedOutput{
			content: []byte(strings.TrimSpace(`
value1
value2
`)),
		},
	},
	{
		title: "multiple value files",
		i: parseInput{
			templateFileName: "example",
			templateContent: []byte(strings.TrimSpace(`
{{.key1}}
{{.key2}}
`)),
			valueFiles: map[string][]byte{
				"values.yaml":  []byte(strings.TrimSpace(`key1: value1`)),
				"values2.yaml": []byte(strings.TrimSpace(`key2: value2`)),
			},
			target: "output",
		},
		o: expectedOutput{
			content: []byte(strings.TrimSpace(`
value1
value2
	`)),
		},
	},
	{
		title: "failed to read value files",
		i: parseInput{
			templateFileName: "example",
			templateContent: []byte(strings.TrimSpace(`
{{.key1}}
{{.key2}}
`)),
			valueFiles: map[string][]byte{
				"values.yaml": []byte(strings.TrimSpace(`{`)),
			},
			target: "output",
		},
		o: expectedOutput{
			err: fmt.Errorf("failed to create combined values file: "),
		},
	},
}

func TestParseTemplate(te *testing.T) {
	if err := log.Init(log.Debug(), "", true); err != nil {
		te.Logf("unable to initialize logger")
		te.FailNow()
	}

	for _, s := range scenariosParserTemplate {
		if selectParserTest != "" && s.title != selectTest {
			continue
		}
		te.Logf("test scenario: %s\n", s.title)

		s.Test(te)
	}
}
func (s *scenarioParserTemplate) Test(te *testing.T) {
	td, err := s.i.setupFiles()
	if err != nil {
		te.Logf("unable to create test dir tree: %v\n", err)
		te.FailNow()
	}
	defer td.Cleanup(te)
	tmpDir := td.Path()

	valueFiles := make([]string, 0, len(s.i.valueFiles))
	for v := range s.i.valueFiles {
		valueFiles = append(valueFiles, filepath.Join(tmpDir, v))
	}
	err = ParseTemplate(
		filepath.Join(tmpDir, s.i.templateFileName),
		valueFiles,
		filepath.Join(tmpDir, s.i.target),
	)
	testfuncs.CheckSimilarErrs(te, s.o.err, err)
	if s.o.err == nil {
		s.o.CheckRes(te, filepath.Join(tmpDir, s.i.target))
	}
}

func (p parseInput) setupFiles() (testfuncs.TestDir, error) {
	genFiles := make(map[string][]byte, 1+len(p.valueFiles))
	genFiles[p.templateFileName] = p.templateContent
	for f, c := range p.valueFiles {
		genFiles[f] = c
	}
	return testfuncs.PrepareTestDirTree(genFiles)
}

func (o expectedOutput) CheckRes(t *testing.T, output string) {
	data, err := os.ReadFile(output)
	if err != nil {
		t.Errorf("failed to read result file %s. Error: %s", data, err)
		t.FailNow()
	}
	if !bytes.Equal(o.content, data) {
		t.Errorf(
			"results do not match: \nwant = \"\n%+v\n\"\ngot  = \"\n%+v\n\"",
			string(o.content),
			string(data),
		)
	}
}
