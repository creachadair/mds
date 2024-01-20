// Package compare contains support functions for comparison of values.
package compare

// FromLessFunc converts a less function, which reports whether its first
// argument precedes its second in an ordering relation, into a comparison that
// returns -1 if a < b, 0 if a == b, and 1 if a > b in the same relation.
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
