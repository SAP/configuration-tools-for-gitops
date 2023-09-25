package generate

import (
	"bytes"
	"fmt"
	"os"
	gotemplate "text/template"

	"github.com/SAP/configuration-tools-for-gitops/v2/pkg/files"
	"github.com/SAP/configuration-tools-for-gitops/v2/pkg/yamlfile"
)

var (
	filesWrite = files.Write
)

func ParseTemplate(filename string, valueFiles []string, target string) error {
	p := parser{}
	if err := p.parse(filename); err != nil {
		return fmt.Errorf("failed to parse file %q: %w", filename, err)
	}

	combinedValues, err := mergeValues(valueFiles)
	if err != nil {
		return err
	}
	var templateInputs interface{}
	if e := combinedValues.Decode(&templateInputs); e != nil {
		return fmt.Errorf("failed to decode values: %w", e)
	}
	output, err := p.execute(templateInputs)
	if err != nil {
		return fmt.Errorf("failed to render template %q: %w", filename, err)
	}
	if err := filesWrite(target, files.AllReadWrite, output); err != nil {
		return fmt.Errorf("failed to write to file %q: %w", target, err)
	}
	return nil
}

type parserInt interface {
	parse(filename string) error
	execute(data interface{}) ([]byte, error)
}

type parserMock struct {
	Mock bool
	Err  error
}

func (m parserMock) parse(filename string) error {
	return nil
}

func (m parserMock) execute(data interface{}) ([]byte, error) {
	return nil, m.Err
}

type parser struct {
	tmpl *gotemplate.Template
}

func (p *parser) parse(filename string) error {
	b, err := os.ReadFile(filename)
	if err != nil {
		return err
	}
	parsed, err := gotemplate.New(filename).Funcs(tmplFuncs()).Parse(string(b))
	if err != nil {
		return err
	}
	p.tmpl = parsed
	return nil
}

func (p parser) execute(data interface{}) ([]byte, error) {
	generated := new(bytes.Buffer)
	err := p.tmpl.Execute(generated, data)
	if err != nil {
		return nil, err
	}
	return generated.Bytes(), nil
}

func mergeValues(valueFiles []string) (res yamlfile.Yaml, err error) {
	res, err = yamlfile.New([]byte{}, yamlfile.SetArrayMergePolicy(yamlfile.Strict))
	if err != nil {
		err = fmt.Errorf("failed to create combined values file: %w", err)
		return
	}
	for _, v := range valueFiles {
		content, e := files.Read(v)
		if e != nil {
			err = fmt.Errorf("failed to read file %q: %w", v, e)
			return
		}
		if _, e := res.MergeBytes(content); e != nil {
			err = fmt.Errorf("failed to combine values file %q: %w", v, err)
			return
		}
	}
	return
}
