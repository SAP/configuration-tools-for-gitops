package generate

import (
	"fmt"
	"strings"
	gotemplate "text/template"

	"github.com/google/uuid"
)

func tmplFuncs() gotemplate.FuncMap {
	// https://www.calhoun.io/intro-to-templates-p3-functions/
	// https://golang.org/pkg/text/template/#FuncMap
	return gotemplate.FuncMap{
		"trimSuffix": strings.TrimSuffix,
		"trimPrefix": strings.TrimPrefix,
		"join": func(sep string, elems ...string) string {
			res := make([]string, 0, len(elems))
			for _, e := range elems {
				if e != "" {
					res = append(res, e)
				}
			}
			return strings.Join(res, sep)
		},
		"split": strings.Split,
		"quote": func(s string) string {
			return fmt.Sprintf("%q", s)
		},
		"select": func(el int, sl []string) string {
			if el >= len(sl) {
				return ""
			}
			return sl[el]
		},
		"uuid4": uuid.NewString,
	}
}
