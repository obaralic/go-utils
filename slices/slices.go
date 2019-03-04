// -----------------------------------------------------------------------------
// This package contains utility functions when working with the slices.
// -----------------------------------------------------------------------------
package slices

const INVALID = -1

type Any interface {
}

func IndexOf(limit int, predicate func(i int) bool) int {
	for index := 0; index < limit; index++ {
		if predicate(index) {
			return index
		}
	}
	return INVALID
}
