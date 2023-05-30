package yamlfile

import (
	"fmt"
	"reflect"

	"gopkg.in/yaml.v3"
)

// Merge merges the input Yaml (from) into the Yaml object (y). Merge rules are:
//   - maps are merged on matching keys
//   - any submap in from is added under the last matching key in into
//   - slices are merged following the 2 rules:
//     1) merges happen element by element
//     2) if the keys in sub-elements match, elements are deep-merged
//   - scalars from overwrite scalars in into
//   - all other combinations the object in into is overwritten with the object in from
//
// The resulting Yaml object is NOT sorted.
func (y *Yaml) Merge(from Yaml) ([]Warning, error) {
	return y.mergeSelective(from, "", true)
}

// MergeSelective merges the input Yaml (from) into the Yaml object (y). It
// follows the same merge rules as the Merge function.
// Before merging, the MergeSelective function filters the from Yaml by the selectFlag
// and will only keep the elements that have the relevant flag (as line comment or yaml flag).
//
// The resulting Yaml object is NOT sorted.
func (y *Yaml) MergeSelective(from Yaml, selectFlag string) ([]Warning, error) {
	return y.mergeSelective(from, selectFlag, false)
}

func (y *Yaml) mergeSelective(from Yaml, selectFlag string, parentSelected bool,
) ([]Warning, error) {
	// from Yaml is empty
	if from.Node.Kind == 0 || len(from.Node.Content) == 0 {
		return []Warning{}, nil
	}
	m := newMerger(selectFlag)
	// into yaml is empty
	if y.Node.Kind == 0 || len(y.Node.Content) == 0 {
		y.Node.Kind = 1
		add, err := m.newContent(from.Node, parentSelected)
		if err != nil {
			return []Warning{}, nil
		}
		if !reflect.DeepEqual(*add, yaml.Node{}) {
			y.Node.Content = append(y.Node.Content, add.Content...)
		}
		return m.warnings, nil
	}
	err := m.merge(from.Node, y.Node, parentSelected, []string{})
	return m.warnings, err
}

func newMerger(selectFlag string) merger {
	return merger{selectFlag, []Warning{}}
}

// merger holds general information for the yaml merging procedure. It holds the
// selectFlag which will be used for filtering the from Yaml and a slice to capture
// all occurring warnings.
type merger struct {
	selectFlag string
	warnings   []Warning
}

// Warning holds the ordered list of nested keys for which a warning occurred and
// the associated warning itself
type Warning struct {
	Keys    []string
	Warning string
}

// newContent will return the selected content of the given yaml Node. Content is
// selected if
// - the parent node was selected
// - the node itself is selected (via the selectFlag, see selectNode)
func (m merger) newContent(from *yaml.Node, parentSelected bool) (*yaml.Node, error) {
	if parentSelected || m.selectNode(from) {
		return from, nil
	}
	selectedNodes := Yaml{from}
	if err := selectedNodes.FilterBy(m.selectFlag); err != nil {
		return nil, err
	}
	return selectedNodes.Node, nil
}

func (m *merger) selectNode(n *yaml.Node) bool {
	if n.ShortTag() == fmt.Sprintf("!%s", m.selectFlag) {
		return true
	}
	if n.LineComment == fmt.Sprintf("# %s", m.selectFlag) {
		return true
	}
	return false
}

// merge checks the types of the top level from and into yaml Nodes and deligates
// the merging to dedicated functions
func (m *merger) merge(from, into *yaml.Node, parentSelected bool, parentKeys []string) error {
	t, err := mergeCombination(from.Kind, into.Kind)
	if err != nil {
		return err
	}
	switch t {
	case scalar2scalar:
		m.mergeScalar(from, into, parentSelected)
	case map2map:
		err = m.mergeMaps(from, into, parentSelected, parentKeys)
	case sequence2sequence:
		into.LineComment = from.LineComment
		into.Tag = from.Tag
		err = m.mergeSequences(from, into, parentSelected, parentKeys)
	case document2document:
		err = m.merge(from.Content[0], into.Content[0], parentSelected, []string{})
	default:
		err = m.mergeDefault(from, into, parentSelected, parentKeys)
	}
	return err
}

// mergeScalar overwrites the content of the into node with the content of the
// from node if the from node is selected
func (m merger) mergeScalar(from, into *yaml.Node, parentSelected bool) {
	if parentSelected || m.selectNode(from) {
		into.LineComment = from.LineComment
		into.Tag = from.Tag
		into.Value = from.Value
	}
}

// mergeDefault is called when 2 non equal types of yaml Nodes are merged.
// In principle, the selected content of the from node will overwrite the content
// of the into node (if the selected content is non-empty).
func (m *merger) mergeDefault(from, into *yaml.Node, parentSelected bool, parentKeys []string,
) error {
	// creates a warning, if length of the content for from and into nodes differ
	// If the length of either from or into changes, then the merge logic can give
	// unwanted results - hence the warning.
	if from.Kind == yaml.SequenceNode && len(from.Content) != len(into.Content) {
		m.warnings = append(m.warnings, Warning{
			Keys: parentKeys,
			Warning: fmt.Sprintf(
				"sequence length from (%v) does not match length into (%v)",
				len(from.Content), len(into.Content)),
		})
	}
	overwrite, err := m.newContent(from, parentSelected)
	if err != nil {
		return err
	}
	if !reflect.DeepEqual(*overwrite, yaml.Node{}) {
		*into = *overwrite
	}
	return nil
}

// mergeSequences applies the following rules for merging sequences
//
//	from = [a,b]
//	into = [d,e,f]
//
// will result in
//
//	res = [a,b,f]
//
//	from = [{k3: NN3, k4: NN4}, {k3: NN5, k7: NN7}, d]
//	into = [{k1: o1, k2:o2}, {k3: o3, k4: o4}, {k5: o5, k6: o6}]
//
// will result in
//
//	res = [{k1: o1, k2:o2, k3: NN3, k4: NN4}, {k3: NN5, k4: NN4, k7: NN7}, d]
func (m *merger) mergeSequences(from, into *yaml.Node,
	parentSelected bool, parentKeys []string,
) error {
	lenFrom := len(from.Content)
	lenInto := len(into.Content)

	if lenFrom != lenInto {
		// creates a warning, if length of the content for from and into nodes differ
		// If the length of either from or into changes, then the merge logic can give
		// unwanted results - hence the warning.
		m.warnings = append(m.warnings, Warning{
			Keys: parentKeys,
			Warning: fmt.Sprintf(
				"sequence length from (%v) does not match length into (%v)",
				lenFrom, lenInto),
		})
	}

	for i, fromEl := range from.Content {
		// the sequence in the merge source (from) is longer than the merge target (into)
		if i >= lenInto {
			add, err := m.newContent(fromEl, parentSelected)
			if err != nil {
				return err
			}
			if !reflect.DeepEqual(*add, yaml.Node{}) {
				into.Content = append(into.Content, add)
			}
			continue
		}

		err := m.merge(
			fromEl,
			into.Content[i],
			parentSelected,
			append(parentKeys, fmt.Sprintf("%v", i)),
		)
		if err != nil {
			return err
		}
	}
	return nil
}

// mergeMaps finds matching keys in the from and into yaml.Node, merges comments
// and tags and subsequently calls the merge method for its values.
func (m *merger) mergeMaps(from, into *yaml.Node,
	parentSelected bool, parentKeys []string,
) error {
	if len(from.Content)%2 != 0 || len(into.Content)%2 != 0 {
		return fmt.Errorf(
			"illegal content length %v found: content slice must be even for maps",
			len(from.Content),
		)
	}

	// iterating over keys (which are located at even positions in the slice)
	for i := 0; i < len(from.Content); i += 2 {
		fromKey, err := readKey(from.Content[i])
		if err != nil {
			return err
		}
		keyExistsInTarget := false
		for j := 0; j < len(into.Content); j += 2 {
			intoKey, err := readKey(into.Content[j])
			if err != nil {
				return err
			}
			if fromKey == intoKey {
				into.LineComment = from.LineComment
				into.Tag = from.Tag
				// keys match: call merge on the sub-objects of from and into
				selected := parentSelected
				if m.selectNode(from.Content[i+1]) {
					selected = true
				}
				if err := m.merge(
					from.Content[i+1],
					into.Content[j+1],
					selected,
					append(parentKeys, intoKey),
				); err != nil {
					return err
				}
				keyExistsInTarget = true
				break
			}
		}
		if !keyExistsInTarget {
			if parentSelected || m.selectNode(from.Content[i+1]) {
				into.Content = append(into.Content, from.Content[i:i+2]...)
				continue
			}
			selectedNodes := PartialCopy(Yaml{from}, i, i+2)
			if err := selectedNodes.FilterBy(m.selectFlag); err != nil {
				return err
			}
			if len(selectedNodes.Node.Content) > 0 {
				into.Content = append(into.Content, selectedNodes.Node.Content...)
			}
		}
	}
	return nil
}

func readKey(n *yaml.Node) (string, error) {
	if n.Kind != yaml.ScalarNode {
		return "", fmt.Errorf("merge for non-scalar map keys is not implemented")
	}
	return n.Value, nil
}

type mergeType uint32

const (
	scalar2scalar mergeType = 1 << iota
	scalar2x

	map2map
	map2scalar
	map2sequence

	sequence2sequence
	sequence2scalar
	sequence2map

	document2document
)

func mergeCombination(from, into yaml.Kind) (mergeType, error) {
	switch from {
	case yaml.ScalarNode:
		switch into {
		case yaml.ScalarNode:
			return scalar2scalar, nil
		default:
			return scalar2x, nil
		}

	case yaml.MappingNode:
		switch into {
		case yaml.ScalarNode:
			return map2scalar, nil
		case yaml.MappingNode:
			return map2map, nil
		case yaml.SequenceNode:
			return map2sequence, nil
		default:
			return scalar2scalar, fmt.Errorf(
				"merge combination from %v (yaml.MappingNode) into %v not supported",
				from, into,
			)
		}

	case yaml.SequenceNode:
		switch into {
		case yaml.ScalarNode:
			return sequence2scalar, nil
		case yaml.MappingNode:
			return sequence2map, nil
		case yaml.SequenceNode:
			return sequence2sequence, nil
		default:
			return scalar2scalar, fmt.Errorf(
				"merge combination from %v (yaml.SequenceNode) into %v not supported",
				from, into,
			)
		}
	case yaml.DocumentNode:
		switch into {
		case yaml.DocumentNode:
			return document2document, nil
		default:
			return scalar2scalar, fmt.Errorf(
				"merge combination from %v (yaml.DocumentNode) into %v not supported",
				from, into,
			)
		}
	case yaml.AliasNode:
		return scalar2scalar, fmt.Errorf(
			"merge combination from %v (yaml.AliasNode) into %v not supported",
			from, into,
		)
	default:
		return scalar2scalar, fmt.Errorf(
			"merge combination from %v into %v not supported",
			from, into,
		)
	}
}
