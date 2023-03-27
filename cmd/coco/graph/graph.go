package graph

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"sort"
	"strings"

	"github.com/yourbasic/graph"
)

// ComponentDependencies holds a complete representation of all dependencies from upstream
// to downstream dependencies for all components. Dependencies are weighted by
// their distance to the reference component (the top-level key)
//
//	weight 0: a direct dependency
//	weight 1: indirect dependency (1 level in between)
//
// E.g.:
//
//	A ← + ← C ← E
//	    ↑   ↑
//	    + ← D
//
// encoded as
//
//	map[
//	  A: map[],
//	  C: map[0:map[A:true]],
//	  D: map[0:map[A:true, C:true]],
//	  E: map[0:map[C:true], 1:map[A:true]]
//	]
type ComponentDependencies map[string]WeightedDeps
type WeightedDeps map[int]map[string]bool

type OutputFormat string

const (
	JSON  OutputFormat = "json"
	YAML  OutputFormat = "yaml"
	FLAT  OutputFormat = "flat"
	UNSET OutputFormat = "unset"
)

func CastOutputFormat(s string) (OutputFormat, bool) {
	switch OutputFormat(s) {
	case YAML:
		return YAML, true
	case JSON:
		return JSON, true
	case FLAT:
		return FLAT, true
	default:
		return UNSET, false
	}
}

func (wd WeightedDeps) ToArray() [][]string {
	res := make([][]string, 0, len(wd))
	for _, w := range keys(wd) {
		d := keys(wd[w])
		res = append(res, d)
	}
	return res
}

// For each components Print returns the upstream dependencies weighted by the
// number of links separating them from the component in question
// In yaml format the output will have the form
//
//	{componentName: {weight: [dependency-1, dependency-2, ...], ...}, ...}
//
// In json format the output will have the form
//
//	{componentName: [[dependency-1, dependency-2, ...], ...], ...}
//
// where the outer array is ordered by increasing weights.
func (cd ComponentDependencies) Print(w io.Writer, format OutputFormat) error {
	switch format {
	case YAML:
		res, err := cd.yamlFormat()
		if err != nil {
			return err
		}
		return write(w, res)
	case JSON:
		res := make(map[string][][]string, len(cd))
		for dep, wd := range cd {
			res[dep] = wd.ToArray()
		}
		out, err := json.Marshal(res)
		if err != nil {
			return err
		}
		return write(w, string(out))
	case FLAT:
		allComponents := map[string]bool{}
		for c, wd := range cd {
			allComponents[c] = true
			for _, deps := range wd {
				for d := range deps {
					allComponents[d] = true
				}
			}
		}
		res := make([]string, 0, len(allComponents))
		for c := range allComponents {
			res = append(res, c)
		}
		sort.Strings(res)
		return write(w, strings.Join(res, " "))
	default:
		return fmt.Errorf("illegal format \"%s\" received", format)
	}
}

func (cd ComponentDependencies) yamlFormat() (string, error) {
	raw := bytes.Buffer{}
	for _, c := range keys(cd) {
		weightedDeps := cd[c]
		if err := write(&raw, fmt.Sprintf("%s:", c)); err != nil {
			return "", err
		}

		weights := keys(weightedDeps)
		if len(weights) == 0 {
			if err := write(&raw, " {}\n"); err != nil {
				return "", err
			}
			continue
		}

		if err := write(&raw, "\n"); err != nil {
			return "", err
		}
		for _, weight := range keys(weightedDeps) {
			deps := keys(weightedDeps[weight])
			if len(deps) == 0 {
				if err := write(&raw, fmt.Sprintf("  %v: []\n", weight)); err != nil {
					return "", err
				}
				continue
			}
			if err := write(&raw, fmt.Sprintf("  %v: %+v\n", weight, deps)); err != nil {
				return "", err
			}
		}
	}
	return raw.String(), nil
}

func (cd ComponentDependencies) Keys() []string {
	return keys(cd)
}

func (cd ComponentDependencies) MaxDepth(m int) ComponentDependencies {
	if m < 0 {
		return cd
	}
	res := ComponentDependencies{}
	for c, wd := range cd {
		resWeights := WeightedDeps{}
		for w, deps := range wd {
			if w < m {
				resWeights[w] = deps
			}
		}
		res[c] = resWeights
	}
	return res
}

func (wd WeightedDeps) Keys() []int {
	return keys(wd)
}

func write(w io.Writer, s string) error {
	_, err := w.Write([]byte(s))
	return err
}

// DownToUp is a map from downstream components to all their direct upstream
// dependencies. E.g. the graph
//
//	A -+ → C
//	   |
//	   + → D
//
// would be encoded as map[A] = {C,D}
type DownToUp map[string]map[string]bool

// GenerateUpToDown constructs a dependency graph that flows from
// upstream to downstream components (where component A is downstream of C if it
// depends on C). The graph is constructed from a map of all direct dependencies
// (given in form downstream to upstream). E.g. the input
//
//	A → + → C → E
//	    ↓   ↓
//	    + → D
//
// will result in the upstream to downstream graph
//
//	A ← + ← C ← E
//	    ↑   ↑
//	    + ← D
//
// encoded as
//
//	map[
//	  A: map[],
//	  C: map[0:map[A:true]],
//	  D: map[0:map[A:true, C:true]],
//	  E: map[0:map[C:true], 1:map[A:true]]
//	]
func GenerateUpToDown(components DownToUp) ComponentDependencies {
	sanitized := validateInputs(components)
	compIndices, names := namesToIndices(sanitized)

	graphDownToUp := graph.New(len(sanitized))
	for c, deps := range compIndices {
		for d := range deps {
			graphDownToUp.AddCost(c, d, 1)
		}
	}
	graphUpToDown := graph.Transpose(graphDownToUp)
	result := make(ComponentDependencies, len(sanitized))
	for v := 0; v < graphUpToDown.Order(); v++ {
		paths, distances := graph.ShortestPaths(graphUpToDown, v)
		// weight info map for current vertex
		weightedDeps := make(map[int]map[string]bool)
		for i, s := range distances {
			if paths[i] >= 0 {
				vertexToAdd := names.idToName[i]
				// check if current weightgroup is already present
				if val, ok := weightedDeps[int(s-1)]; ok {
					// if weight is already precent, append new vertex to it
					// {1:{A:true}} becomes {1:{A:true, B:true}}
					val[vertexToAdd] = true
				} else {
					// else add new weight and vertex to weightInfo map
					// {1:{A:true}} becomes {1:{A:true}, 2:{B:true}}
					weightedDeps[int(s-1)] = map[string]bool{vertexToAdd: true}
				}
			}
		}
		currentVertex := names.idToName[v]
		result[currentVertex] = weightedDeps
	}

	return result
}

type nameDict struct {
	idToName []string
	nameToID map[string]int
}

func namesToIndices(allDeps DownToUp) (map[int]map[int]bool, nameDict) {
	stringNames := keys(allDeps)
	names := nameDict{
		idToName: stringNames,
		nameToID: make(map[string]int, len(stringNames)),
	}
	for i, n := range stringNames {
		names.nameToID[n] = i
	}

	indices := make(map[int]map[int]bool, len(stringNames))
	for comp, deps := range allDeps {
		depsIndices := make(map[int]bool, len(deps))
		for d := range deps {
			depsIndices[names.nameToID[d]] = true
		}
		indices[names.nameToID[comp]] = depsIndices
	}
	return indices, names
}

// Checks if there are dependencies that do not appear as component keys. If
// any such dependencies are found they are added as components without further
// dependencies.
func validateInputs(comp DownToUp) DownToUp {
	names := keys(comp)
	for _, n := range names {
		deps := comp[n]
		for depName := range deps {
			_, ok := comp[depName]
			if !ok {
				comp[depName] = map[string]bool{}
			}
		}
	}
	return comp
}

// returns the sorted list of keys of an input map
func keys[O string | int, A any](m map[O]A) []O {
	keys := make([]O, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	sort.Slice(
		keys,
		func(i, j int) bool {
			return keys[i] < keys[j]
		})
	return keys
}
