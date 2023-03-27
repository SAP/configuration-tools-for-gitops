package yamlfile

import (
	"fmt"
	"reflect"
)

// Merge merges x2 into x1. Merge rules are:
//   - maps are merged on matching keys
//   - any submap in from is added under the last matching key in into
//   - 2 slices are merged by appending the from slice to the into slice
//   - scalars from overwrite scalars in into
//   - for all other combinations Merge with return an error
func Merge(x1, x2 interface{}) (interface{}, error) {
	if !iterable(x2) {
		return x2, nil
	}
	switch x1 := x1.(type) {
	case map[string]interface{}:
		x2c, ok := x2.(map[string]interface{})
		if !ok {
			return typeErr(reflect.TypeOf(x1).String(), reflect.TypeOf(x2).String())
		}
		for k, v2 := range x2c {
			if v1, ok := x1[k]; ok {
				// keys of x1 and x2 match
				var err error
				// call merge for the subvalues (e.g. submaps) of x1 and x2
				x1[k], err = Merge(v1, v2)
				if err != nil {
					return nil, err
				}
			} else {
				// keys of x1 and x2 do not match
				x1[k] = v2
			}
		}
		return x1, nil
	case nil:
		return x2, nil
	case []interface{}:
		x2c, ok := x2.([]interface{})
		if !ok {
			return typeErr(reflect.TypeOf(x1).String(), reflect.TypeOf(x2).String())
		}
		// append the slice in x2 to the slice in x1
		x1 = append(x1, x2c...)
		return x1, nil
	case []string:
		x2c, ok := x2.([]string)
		if !ok {
			return typeErr(reflect.TypeOf(x1).String(), reflect.TypeOf(x2).String())
		}
		// append the slice in x2 to the slice in x1
		x1 = append(x1, x2c...)
		return x1, nil

	default:
		return nil, fmt.Errorf("type %s not implemented for merging", reflect.TypeOf(x1).String())
	}
}

func iterable(x interface{}) bool {
	switch x.(type) {
	case map[string]interface{}:
		return true
	case []interface{}, []string, []bool, []int, []float32:
		return true
	default:
		return false
	}
}
func typeErr(type1, type2 string) (interface{}, error) {
	return nil, fmt.Errorf("cannot merge types %s and %s", type1, type2)
}
