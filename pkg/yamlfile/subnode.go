package yamlfile

import (
	"errors"
	"fmt"
	"strings"

	"gopkg.in/yaml.v3"
)

var (
	errUnknownNodeKind         = subnodeError{errors.New("unknown node kind")}
	errKeyNotPresent           = subnodeError{errors.New("key not present in yaml.Node")}
	errDocumentNodeWrongLength = subnodeError{errors.New("document Content must have length 1")}
)

func (y Yaml) SelectSubElement(keys []string) (Yaml, error) {
	subNode, err := subnode{nil}.node(y.Node, keys)
	if err != nil {
		return Yaml{}, err
	}
	s := y.settings.Copy()
	return Yaml{subNode, &s}, nil
}

func (y *Yaml) Insert(parentKeys []string, subelement interface{}) ([]Warning, error) {
	_, err := subnode{subelement}.node(y.Node, parentKeys)
	if errUnknownNodeKind.Is(err) || errKeyNotPresent.Is(err) {
		newContent, e := constructYaml(parentKeys, subelement)
		if e != nil {
			return []Warning{}, fmt.Errorf("failed to construct Yaml: %v", e)
		}
		return y.Merge(newContent)
	}
	return []Warning{}, err
}

func constructYaml(keys []string, content interface{}) (Yaml, error) {
	raw, err := yaml.Marshal(constructInterface(keys, content))
	if err != nil {
		return Yaml{}, fmt.Errorf("failed to marshal interface: %v", err)
	}
	return New(raw)
}

func constructInterface(keys []string, content interface{}) interface{} {
	if len(keys) == 0 {
		return content
	}
	return map[string]interface{}{keys[0]: constructInterface(keys[1:], content)}
}

type subnode struct {
	insert interface{}
}

func (s subnode) node(n *yaml.Node, keys []string) (*yaml.Node, error) {
	if len(keys) == 0 {
		if s.insert != nil {
			subYaml, err := constructYaml([]string{}, s.insert)
			if err != nil {
				return nil, err
			}
			*n = *subYaml.Node.Content[0]
		}
		return n, nil
	}
	switch n.Kind {
	case yaml.MappingNode:
		return s.mapNode(n, keys)
	case yaml.DocumentNode:
		return s.documentNode(n, keys)
	default:
		return nil, fmt.Errorf("%v %v", errUnknownNodeKind, n.Kind)
	}
}

func (s subnode) mapNode(n *yaml.Node, keys []string) (*yaml.Node, error) {
	i := 0
	key0, keyRemain := keys[0], keys[1:]

	selectedByKey := false
	for i < len(n.Content) {
		if i%2 == 0 {
			// map keys appear at even positions in the Content slice
			if n.Content[i].Value == key0 {
				selectedByKey = true
			}
			i++
			continue
		}
		if selectedByKey {
			return s.node(n.Content[i], keyRemain)
		}
		i++
	}
	return nil, fmt.Errorf("%v: %q", errKeyNotPresent, key0)
}

func (s subnode) documentNode(n *yaml.Node, keys []string) (*yaml.Node, error) {
	if len(n.Content) != 1 {
		return nil, fmt.Errorf("%v: found length %v", errDocumentNodeWrongLength, len(n.Content))
	}
	return s.node(n.Content[0], keys)
}

type subnodeError struct {
	error
}

func (e subnodeError) Is(err error) bool {
	if err == nil {
		return false
	}
	return strings.Contains(err.Error(), e.Error())
}
