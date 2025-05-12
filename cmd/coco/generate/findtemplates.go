package generate

import (
	"fmt"
	"os"
	"sort"
	"strings"

	"github.com/SAP/configuration-tools-for-gitops/v2/pkg/files"
)

func findTemplates(
	basepath, tmplIdentifier string, includeFilters, excludeFilters []string,
) (map[string][]template, error) {
	include := []string{tmplIdentifier}
	include = append(include, includeFilters...)

	list, err := files.New(basepath).
		Include(files.AND, include).
		Exclude(files.OR, excludeFilters).
		Execute()
	if err != nil {
		return nil, err
	}
	filteredTemplates := list.Content()
	res := make(map[string][]template, len(filteredTemplates))
	for path, file := range filteredTemplates {
		if file.IsDir {
			continue
		}
		addTemplate(res, path, tmplIdentifier)
	}
	for k, v := range res {
		sort.Slice(v, func(i, j int) bool {
			return v[i].source < v[j].source
		})
		res[k] = v
	}
	return res, nil
}

func addTemplate(res map[string][]template, path, tmplIdentifier string) {
	pathSlice := strings.Split(path, string(os.PathSeparator))
	var i int
	for i = len(pathSlice) - 1; i >= 0; i-- {
		if strings.Contains(pathSlice[i], tmplIdentifier) {
			break
		}
	}

	b := strings.Join(pathSlice[:i], string(os.PathSeparator))
	s := ""
	if i < len(pathSlice)-1 {
		s = fmt.Sprintf("%s%s",
			string(os.PathSeparator),
			strings.Join(pathSlice[i+1:], string(os.PathSeparator)),
		)
	}
	pre := strings.Replace(pathSlice[i], tmplIdentifier, "", 1)

	t := template{source: path, basepath: b, namePrefix: pre, subpath: s}

	templates, ok := res[b]
	if ok {
		res[b] = append(templates, t)
	} else {
		res[b] = []template{t}
	}
}
