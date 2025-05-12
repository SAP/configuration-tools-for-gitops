package maputils

import (
	"github.com/SAP/configuration-tools-for-gitops/v2/pkg/sliceutils"
	"golang.org/x/exp/constraints"
)

func Keys[K comparable, V any](m map[K]V) []K {
	res := make([]K, 0, len(m))
	for k := range m {
		res = append(res, k)
	}
	return res
}

func KeysSorted[K constraints.Ordered, V any](m map[K]V) []K {
	res := Keys(m)
	sliceutils.Sort(res)
	return res
}
