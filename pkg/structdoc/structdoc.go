package structdoc

import (
	"fmt"
	"reflect"
	"strings"
)

const (
	valuesTag = "doc"
	keyTag    = "yaml"
)

// Generate constructs a documentation interface from the provided struct by
// using the struct field types as well as the information given in the "doc" struct
// tag. In the doc tag the following information can be given in a comma-separated list:
//
//   - msg: describing message
//   - default: default value
//   - req: required field
//   - option,o: 1 possible value (multiple values are specified by multiple occurrences)
//
// e.g.: `doc:"msg=this is my message,default=0,o=0, o=1, req"`. If no "doc"
// tag is present the field will be ignored. The output format is map[string]interface{},
// where the interface is either a doc-string or a substructure of the same format.
func Generate(s interface{}) interface{} {
	return parseNode(reflect.TypeOf(s))
}

func parseNode(t reflect.Type) interface{} {
	switch t.Kind() {
	case reflect.Ptr:
		return parseNode(t.Elem())
	case reflect.Struct:
		return parseStruct(t)
	case reflect.Array, reflect.Slice:
		return parseSlice(t)
	case reflect.Map:
		return parseMap(t)
	default:
		return parsePrimitive(t)
	}
}

func parseMap(t reflect.Type) interface{} {
	tV := parseNode(t.Elem())

	return map[string]interface{}{str(t.Key()): tV}
}

func parseSlice(t reflect.Type) interface{} {
	tV := parseNode(t.Elem())
	return []interface{}{tV}
}

func parsePrimitive(t reflect.Type) interface{} {
	return fmt.Sprintf("(%v)", findType(t))
}

func parseStruct(t reflect.Type) interface{} {
	res := map[string]interface{}{}
	for i := 0; i < t.NumField(); i++ {
		field := t.Field(i)
		tKind := findType(field.Type)

		vTag, ok := field.Tag.Lookup(valuesTag)
		if vTag == "-" {
			continue
		}
		key := structKeyName(&field)
		if ok {
			res[key] = docFromTag(vTag, str(field.Type))
			continue
		}
		switch tKind {
		case reflect.Struct:
			subStruct := parseNode(field.Type)

			if !reflect.DeepEqual(subStruct, map[string]interface{}{}) {
				res[key] = subStruct
			}
		case reflect.Array, reflect.Slice:
			subStruct := parseNode(field.Type)
			if !reflect.DeepEqual(subStruct, []interface{}{}) {
				res[key] = subStruct
			}
		case reflect.Map:
			subStruct := parseNode(field.Type)
			if !reflect.DeepEqual(subStruct, []interface{}{}) {
				res[key] = subStruct
			}
		default:
			res[key] = docFromTag(vTag, str(field.Type))
		}
	}
	return res
}

func str(t reflect.Type) string {
	switch t.Kind() {
	case reflect.Ptr:
		return str(t.Elem())
	case reflect.Map:
		return fmt.Sprintf("map[%s]%s", str(t.Key()), str(t.Elem()))
	case reflect.Array, reflect.Slice:
		return fmt.Sprintf("[]%s", str(t.Elem()))
	default:
		return t.Kind().String()
	}
}

func findType(raw reflect.Type) reflect.Kind {
	var t reflect.Kind
	switch raw.Kind() {
	case reflect.Ptr:
		t = raw.Elem().Kind()
	default:
		t = raw.Kind()
	}
	return t
}

func structKeyName(f *reflect.StructField) string {
	name := f.Tag.Get(keyTag)
	if name == "" || name == "-" {
		name = f.Name
	}
	return strings.TrimSuffix(name, ",omitempty")
}
func docFromTag(tag, typeVal string) string {
	tags := strings.Split(tag, ",")
	var msg, reqVal, defaultVal string
	optionsSlice := []string{}
	required := false

	for _, raw := range tags {
		t := strings.TrimSpace(raw)
		kv := strings.Split(t, "=")
		if len(kv) == 0 {
			continue
		}
		if len(kv) == 1 && kv[0] == "req" {
			required = true
			continue
		}
		switch kv[0] {
		case "msg":
			msg = kv[1]
		case "default":
			defaultVal = fmt.Sprintf("default:%q", kv[1])
		case "req":
			reqVal = kv[1]
		case "option", "o":
			optionsSlice = append(optionsSlice, kv[1])
		}
	}
	options := ""
	if len(optionsSlice) > 0 {
		options = fmt.Sprintf("options:[%v]", strings.Join(optionsSlice, ","))
	}
	info := fmt.Sprintf("(%v)", strings.Join(appendNonEmpty(
		[]string{}, typeVal, defaultVal, options,
	), ", "))
	var req string
	if reqVal != "" {
		req = fmt.Sprintf("REQUIRED:%q", reqVal)
	}
	if required {
		req = "REQUIRED"
	}

	return strings.Join(appendNonEmpty([]string{}, msg, info, req), " ")
}

func appendNonEmpty(s []string, els ...string) []string {
	res := make([]string, 0, len(s))
	_ = copy(res, s)
	for _, e := range els {
		if e == "" {
			continue
		}
		res = append(res, e)
	}
	return res
}
