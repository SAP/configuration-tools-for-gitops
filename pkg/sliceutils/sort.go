package sliceutils

import (
	"golang.org/x/exp/constraints"
	"sort"
)

func Sort[T constraints.Ordered](s []T) {
	sort.Slice(s, func(i, j int) bool {
		return s[i] < s[j]
	})
}
