package generate

import (
	"strings"
	gotemplate "text/template"

	"github.com/Masterminds/sprig"
)

func tmplFuncs() gotemplate.FuncMap {
	// https://www.calhoun.io/intro-to-templates-p3-functions/
	// https://golang.org/pkg/text/template/#FuncMap
	funcMaps := sprig.FuncMap()
	funcMaps["select"] = func(el int, sl []string) string {
		if el >= len(sl) {
			return ""
		}
		return sl[el]
	}
	funcMaps["joinElems"] = func(sep string, elems ...string) string {
		res := make([]string, 0, len(elems))
		for _, e := range elems {
			if e != "" {
				res = append(res, e)
			}
		}
		return strings.Join(res, sep)
	}
	return funcMaps
}
