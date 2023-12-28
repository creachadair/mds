// Package value defines comparisons and adapters for value types.
package value

// LessCompare converts a less function, which reports whether its first
// argument precedes its second in an ordering relation, into a comparison that
// returns -1 if a < b, 0 if a == b, and 1 if a > b in the same relation.
func LessCompare[T any](less func(a, b T) bool) func(a, b T) int {
	return func(a, b T) int {
		if less(a, b) {
			return -1
		} else if less(b, a) {
			return 1
		}
		return 0
	}
}

// Ptr returns a pointer to its argument type containing v.
func Ptr[T any](v T) *T { return &v }
