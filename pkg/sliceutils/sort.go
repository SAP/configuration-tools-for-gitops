package sliceutils

import (
	"cmp"
	"sort"
)

func Sort[T cmp.Ordered](s []T) {
	sort.Slice(s, func(i, j int) bool {
		return s[i] < s[j]
	})
}
