package yamlfile

import (
	"fmt"
	"sort"

	"gopkg.in/yaml.v3"
)

// Sort deeply sorts the Yaml object. Sorting rules:
//   - maps are sorted alphabetically by key
//   - arrays are not sorted
func (y *Yaml) Sort() {
	sorter{}.sort(y.Node)
}

type sorter struct{}

// sort sends a sorting request for n to the specific node-type implementation.
func (s sorter) sort(n *yaml.Node) {
	switch n.Kind {
	case yaml.ScalarNode:
		return
	case yaml.MappingNode:
		s.sortMaps(n)
	case yaml.SequenceNode:
		for _, c := range n.Content {
			s.sort(c)
		}
	case yaml.DocumentNode:
		for _, c := range n.Content {
			s.sort(c)
		}
	default:
		return
	}
}

// sortMaps first sorts the n.Content slice of the node (which is known to be of
// map-type). Then the sort method is invoced for every value of the map.
func (s sorter) sortMaps(n *yaml.Node) {
	if len(n.Content)%2 != 0 {
		panic(fmt.Errorf(
			"illegal content length %v found: content slice must be even for maps",
			len(n.Content),
		))
	}
	sortContent(n)
	for i, c := range n.Content {
		if i%2 == 0 {
			continue
		}
		s.sort(c)
	}
}

func sortContent(n *yaml.Node) {
	mapKeys := make([]string, 0, len(n.Content)/2)
	keyIndex := make(map[string]int, len(n.Content)/2)
	for i := 0; i < len(n.Content); i += 2 {
		k := n.Content[i].Value
		mapKeys = append(mapKeys, k)
		keyIndex[k] = i
	}
	sort.Strings(mapKeys)
	sortedContent := make([]*yaml.Node, len(n.Content))
	for i, k := range mapKeys {
		sortedContent[2*i] = n.Content[keyIndex[k]]
		sortedContent[2*i+1] = n.Content[keyIndex[k]+1]
	}
	n.Content = sortedContent
}
