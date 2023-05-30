package yamlfile

import (
	"fmt"
	"io"

	"gopkg.in/yaml.v3"
)

// Yaml objects hold a yaml file in form of a yaml.Node object. This is the central
// object for all implementations in this package.
// The yaml.Node representation is recursive and every level of a yaml file is
// encoded into a dedicated yaml.Node. The different levels are nested via the
// yaml.Node.Content field.
// All Implementations in the yamlfile package follow this recursive approach and
// solve their requirements for each level in the recursive yaml.Node structure.
type Yaml struct {
	Node *yaml.Node
}

// New unmarshalls a yaml input into a yaml.Node representation and returns a Yaml type.
func New(input []byte) (Yaml, error) {
	var node yaml.Node
	if err := yaml.Unmarshal(input, &node); err != nil {
		return Yaml{}, fmt.Errorf("unmarshalling failed %s", err)
	}
	return Yaml{&node}, nil
}

func NewFromInterface(i interface{}) (Yaml, error) {
	var node yaml.Node
	input, err := yaml.Marshal(i)
	if err != nil {
		return Yaml{}, fmt.Errorf("marshaling failed %s", err)
	}
	if err := yaml.Unmarshal(input, &node); err != nil {
		return Yaml{}, fmt.Errorf("unmarshalling failed %s", err)
	}
	return Yaml{&node}, nil
}

// PartialCopy creates a copy of the input n but with only the subslice of the
// content n.Content[start:end]
func PartialCopy(n Yaml, start, end int) Yaml {
	newContent := make([]*yaml.Node, 0, len(n.Node.Content))
	for i, el := range n.Node.Content {
		if i >= start && i < end {
			newContent = append(newContent, deepCopy(el))
		}
	}
	var newAlias *yaml.Node
	if n.Node.Alias != nil {
		newAlias = &yaml.Node{}
		*newAlias = *n.Node.Alias
	}
	newNode := yaml.Node{
		Kind:        n.Node.Kind,
		Style:       n.Node.Style,
		Tag:         n.Node.Tag,
		Value:       n.Node.Value,
		Anchor:      n.Node.Anchor,
		Alias:       newAlias,
		Content:     newContent,
		HeadComment: n.Node.HeadComment,
		LineComment: n.Node.LineComment,
		FootComment: n.Node.FootComment,
		Line:        n.Node.Line,
		Column:      n.Node.Column,
	}
	return Yaml{&newNode}
}

// Copy creates a deep copy of the Yaml object.
func (y Yaml) Copy() Yaml {
	return Yaml{deepCopy(y.Node)}
}

func deepCopy(n *yaml.Node) *yaml.Node {
	newContent := make([]*yaml.Node, 0, len(n.Content))
	for _, el := range n.Content {
		newContent = append(newContent, deepCopy(el))
	}
	return &yaml.Node{
		Kind:        n.Kind,
		Style:       n.Style,
		Tag:         n.Tag,
		Value:       n.Value,
		Anchor:      n.Anchor,
		Alias:       nil,
		Content:     newContent,
		HeadComment: n.HeadComment,
		LineComment: n.LineComment,
		FootComment: n.FootComment,
		Line:        n.Line,
		Column:      n.Column,
	}
}

// Decode unmarshals the Yaml into a provided interface v.
func (y *Yaml) Decode(v interface{}) error {
	return y.Node.Decode(v)
}

// Encode marshalls a Yaml into the provided Writer w. The number of space
// indentations in the output can be controlled via the indent parameter.
func (y *Yaml) Encode(w io.Writer, indent int) error {
	if y.Node.Kind == 0 {
		return nil
	}
	if y.Node.Kind == yaml.DocumentNode && len(y.Node.Content) == 0 {
		return nil
	}
	e := yaml.NewEncoder(w)
	defer e.Close()
	e.SetIndent(indent)
	return e.Encode(y.Node)
}

// FilterBy puts a positive filter on the Node in Yaml. Only elements
// remain that have the provided filter as a LineComment or a yaml Tag.
// E.g. for the filter "keepThis", the following yaml
//
//	key: "will be removed"
//	persistentComment: "will stay" # keepThis
//	persistentTag: !keepThis "will stay"
//		root:
//			nestedComment: "will stay" # keepThis
//			nestedTag: !keepThis "will stay" # keepThis
//		key: "will be removed"
//
// turns into
//
//	persistentComment: "will stay" # keepThis
//	persistentTag: !keepThis "will stay"
//	root:
//		nestedComment: "will stay" # keepThis
//		nestedTag: !keepThis "will stay" # keepThis
func (y *Yaml) FilterBy(filter string) error {
	if y.Node.Kind == 0 {
		return nil
	}
	s := sieve{filter}
	remove, err := s.node(y.Node, false, []string{})
	if remove {
		y.Node = &yaml.Node{}
	}
	return err
}

// FilterByKeys traverses the Yaml object via the ordered slice of keys and
// removes all parts that are not children of the full keys slice.
//
// For the key slice: ["willStay", "subkey"]
//
// the Yaml object:
//
//	key1: "will be removed"
//	key2: "will be removed"
//	willStay:
//		subkey:
//			s1: "will stay"
//			s2: "will stay"
//			s3: ["will stay", "will stay"]
//
// will be reduced to
//
//	willStay:
//		subkey:
//			s1: "will stay"
//			s2: "will stay"
//			s3: ["will stay", "will stay"]
func (y *Yaml) FilterByKeys(keys []string) error {
	if y.Node.Kind == 0 {
		return nil
	}
	s := sieve{}
	remove, err := s.node(y.Node, false, keys)
	if remove {
		y.Node = &yaml.Node{}
	}
	return err
}

// sieve holds the information needed for filtering a yamlfile down to to all
// elements that are marked by the sieve (either as line-comment or as yaml-tag)
type sieve struct {
	filter string
}

// node sends a filtering request for n to the specific node-type filtering
// implementation. The parentSelected parameter hands down information if the
// parent node was already selected by the filter. In this case any child will
// be selected as well.
// The restrictToKeys parameter holds a slice of keys that are selected. If this
// parameter is non-empty all keys that don't have the restrictToKeys slice as parents
// will be sifted out (if the slice is empty no key is sifted out by this filter).
// The method reports back whether all the nodes content has been removed in the
// filtering process.
func (s sieve) node(
	n *yaml.Node, parentSelected bool, restrictToKeys []string,
) (removed bool, err error) {
	switch n.Kind {
	case yaml.ScalarNode:
		return s.scalarNode(n, parentSelected)
	case yaml.SequenceNode:
		return s.sequenceNode(n, parentSelected, restrictToKeys)
	case yaml.MappingNode:
		return s.mapNode(n, parentSelected, restrictToKeys)
	case yaml.DocumentNode:
		return s.documentNode(n, parentSelected, restrictToKeys)
	case yaml.AliasNode:
		return s.aliasNode(n, parentSelected)
	default:
		return false, fmt.Errorf("unknown node kind %v", n.Kind)
	}
}

func (s sieve) scalarNode(n *yaml.Node, parentSelected bool) (remove bool, err error) {
	if parentSelected {
		return false, nil
	}
	if s.selected(n) {
		return false, nil
	}
	return true, nil
}

// selected identifies whether a yaml.Node has a specific linecomment or yaml-tag.
func (s sieve) selected(n *yaml.Node) bool {
	if n.ShortTag() == fmt.Sprintf("!%s", s.filter) {
		return true
	}
	if n.LineComment == fmt.Sprintf("# %s", s.filter) {
		return true
	}
	return false
}

func (s sieve) aliasNode(n *yaml.Node, parentSelected bool) (removed bool, err error) {
	return false, fmt.Errorf("unmarshal yaml.AliasNode not implemented")
}

// mapNode implements the filtering logic for maps. The node method is called
// for every value of the map and if the value must be removed, the key and value
// are taken out of the n.Content. If the Content slice is empty at the end of
// the procedure, the method reports back, that the Node is ready for removal.
func (s sieve) mapNode(
	n *yaml.Node, parentSelected bool, restrictToKeys []string,
) (removeAll bool, err error) {
	removeAll = false
	err = nil
	if parentSelected {
		return
	}
	if s.selected(n) {
		parentSelected = true
	}

	lastKey := 0
	i := 0

	selectedByKey := false
	var restrictKey *string
	var childRestrictKeys []string
	if len(restrictToKeys) == 0 {
		restrictKey, childRestrictKeys = nil, []string{}
	} else {
		restrictKey, childRestrictKeys = &(restrictToKeys[0]), restrictToKeys[1:]
	}
	if len(restrictToKeys) == 1 {
		parentSelected = true
	}
	for i < len(n.Content) {
		if i%2 == 0 {
			// map keys appear at even positions in the Content slice
			if restrictKey == nil {
				selectedByKey = true
			} else if n.Content[i].Value == *restrictKey {
				selectedByKey = true
			}
			lastKey = i
			i++
			continue
		}

		remove := true
		if selectedByKey {
			remove, err = s.node(n.Content[i], parentSelected, childRestrictKeys)
			if err != nil {
				return
			}
		}
		if remove {
			// remove value
			n.Content = append(n.Content[:i], n.Content[i+1:]...)
			// remove key
			n.Content = append(n.Content[:lastKey], n.Content[lastKey+1:]...)
			i--
			continue
		}
		i++
	}
	if len(n.Content) == 0 {
		removeAll = true
	}
	return removeAll, nil
}

// sequenceNode implements the filtering logic for sequences. The node method is called
// for every value of the sequence and if the value must be removed, it is taken
// out of the n.Content.
func (s sieve) sequenceNode(
	n *yaml.Node, parentSelected bool, restrictToKeys []string) (removeAll bool, err error) {
	removeAll = false
	err = nil
	if parentSelected {
		return
	}
	if s.selected(n) {
		parentSelected = true
	}

	i := 0
	for i < len(n.Content) {
		remove := true
		if len(restrictToKeys) == 0 {
			remove, err = s.node(n.Content[i], parentSelected, restrictToKeys)
			if err != nil {
				return
			}
		}
		if remove {
			// remove value from slice
			n.Content = append(n.Content[:i], n.Content[i+1:]...)
			continue
		}
		i++
	}
	if len(n.Content) == 0 {
		removeAll = true
	}
	return removeAll, err
}

// documentNode implements the filtering logic for documents. The node method is called
// for every value of the sequence and if the value must be removed, it is taken
// out of the n.Content.
func (s sieve) documentNode(
	n *yaml.Node, parentSelected bool, restrictToKeys []string) (removeAll bool, err error) {
	removeAll = false
	err = nil
	if parentSelected {
		return
	}
	if s.selected(n) {
		parentSelected = true
	}

	i := 0
	for i < len(n.Content) {
		var remove bool
		remove, err = s.node(n.Content[i], parentSelected, restrictToKeys)
		if err != nil {
			return
		}
		if remove {
			// remove value from slice
			n.Content = append(n.Content[:i], n.Content[i+1:]...)
			continue
		}
		i++
	}
	if len(n.Content) == 0 {
		removeAll = true
	}
	return
}
