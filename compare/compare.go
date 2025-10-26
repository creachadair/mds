// Package compare contains support functions for comparison of values.
package compare

import "time"

// Func is a comparison function, taking two values A and B of a type and
// reporting their relative order. A valid Func must return:
//
//	-1 if A precedes B,
//	 0 if A and B are equivalent,
//	+1 if A follows B
//
// Comparison functions are expected to implement a strict weak ordering.
// Unless otherwise noted, any negative value is accepted in place of -1, and
// any positive value in place of 1.
type Func[T any] func(a, b T) int

// Reversed is a comparison function that orders its elements in the reverse of
// the ordering expressed by f.
func (f Func[T]) Reversed(a, b T) int { return -f(a, b) }

// Less is a [LessFunc] on the same relation as f.
func (f Func[T]) Less(a, b T) bool { return f(a, b) < 0 }

// A LessFunc is a one-sided comparison function, taking two values A and B of
// a type and reporting whether A precedes B in relative order.
//
// Comparison functions are expected to implement a strict weak ordering.
type LessFunc[T any] func(a, b T) bool

// Reversed is a one-sided comparison function that orders its elements in the
// reverse of the ordering expressed by f.
func (f LessFunc[T]) Reversed(a, b T) bool { return f(b, a) }

// Compare is a [Func] on the same relation as f.
func (f LessFunc[T]) Compare(a, b T) int {
	if f(a, b) {
		return -1
	} else if f(b, a) {
		return 1
	}
	return 0
}

// FromLessFunc converts a less function, which reports whether its first
// argument precedes its second in an ordering relation, into a comparison
// function on that same relation.
func FromLessFunc[T any](less LessFunc[T]) Func[T] { return less.Compare }

// ToLessFunc converts a comparison function into a less function on the same
// relation.
func ToLessFunc[T any](cmp Func[T]) LessFunc[T] { return cmp.Less }

// Time is a comparison function for time.Time values that orders earlier times
// before later ones. This is a shim for [time.Time.Compare].
func Time(a, b time.Time) int { return a.Compare(b) }

// Reversed returns a comparison function that orders its elements in the
// reverse of the ordering expressed by c.
func Reversed[T any](c Func[T]) Func[T] { return c.Reversed }

// Bool is a comparison function for bool values that orders false before true.
func Bool(a, b bool) int {
	if a == b {
		return 0
	} else if a {
		return 1
	}
	return -1
}
