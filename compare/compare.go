// Package compare contains support functions for comparison of values.
//
// # Comparison Functions
//
// For the purposes of this package, a comparison function takes two values A
// and B of a type and reports their relative order, returning:
//
//	-1 if A precedes B,
//	 0 if A and B are equivalent,
//	+1 if A follows B
//
// Comparison functions are expected to implement a strict weak ordering.
// Unless otherwise noted, any negative value is accepted in place of -1, and
// any positive value in place of 1.
//
// # Less Functions
//
// For the purposes of this package, a less function takes two values A and B
// of a type and reports whether A precedes B in relative order.
package compare

import "time"

// FromLessFunc converts a less function, which reports whether its first
// argument precedes its second in an ordering relation, into a comparison
// function on that same relation.
func FromLessFunc[T any](less func(a, b T) bool) func(a, b T) int {
	return func(a, b T) int {
		if less(a, b) {
			return -1
		} else if less(b, a) {
			return 1
		}
		return 0
	}
}

// ToLessFunc converts a comparison function into a less function on the same
// relation.
func ToLessFunc[T any](cmp func(a, b T) int) func(a, b T) bool {
	return func(a, b T) bool { return cmp(a, b) < 0 }
}

// Time is a comparison function for time.Time values that orders earlier times
// before later ones.
func Time(a, b time.Time) int {
	if a.Before(b) {
		return -1
	} else if a.Equal(b) {
		return 0
	}
	return 1
}
