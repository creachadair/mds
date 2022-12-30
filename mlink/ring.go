package mlink

import (
	"fmt"
)

// A Ring is a doubly-linked circular chain of data items.  There is no
// designated beginning or end of a ring; each element is a valid entry point
// for the entire ring. A ring with no elements is represented as nil.
type Ring[T any] struct {
	Value T

	prev, next *Ring[T]
}

func (r *Ring[T]) ptr(p *Ring[T]) string {
	if p == nil {
		return "-"
	} else if p == r {
		return "@"
	} else {
		return "*"
	}
}

func (r *Ring[T]) String() string {
	if r == nil {
		return "Ring(empty)"
	}
	return fmt.Sprintf("Ring(%v, %v%v)", r.Value, r.ptr(r.prev), r.ptr(r.next))
}

// NewRing constructs a new ring with n zero-valued elements.
// If n ≤ 0, NewRing returns nil.
func NewRing[T any](n int) *Ring[T] {
	if n <= 0 {
		return nil
	}
	r := newRing[T]()
	for n > 1 {
		elt := newRing[T]()
		elt.next = r.next
		r.next.prev = elt
		elt.prev = r
		r.next = elt
		n--
	}
	return r
}

// RingOf constructs a new ring containing the given elements.
func RingOf[T any](vs ...T) *Ring[T] {
	r := NewRing[T](len(vs))
	cur := r
	for _, v := range vs {
		cur.Value = v
		cur = cur.Next()
	}
	return r
}

// Join splices ring s into a non-empty ring r. There are two cases:
//
// If r and s belong to different rings, [r1 ... rn] and [s1 ... sm], the
// elements of s are spliced in after r and the resulting ring is:
//
//	[r1 s1 ... sm r2 ... rn]
//
// In this case Join returns the ring [r2 ... rn r1 ... sm].
//
// If r and s belong to the same ring, [r1 r2 ... ri s1 ... sm ... rn], then
// the loop of the ring from r2 ... ri is spliced out of r and the resulting
// ring is:
//
//	[r1 s1 ... sm ... rn]
//
// In this case Join returns the ring [r2 ... ri] that was spliced out.  This
// may be empty (nil) if there were no elements between r1 and s1.
func (r *Ring[T]) Join(s *Ring[T]) *Ring[T] {
	if r == s || r.next == s {
		return nil // same ring, nothing to do
	}
	rnext, sprev := r.next, s.prev

	r.next = s         // successor of r is now s
	s.prev = r         // predecessor of s is now r
	sprev.next = rnext // successor of s end is now rnext
	rnext.prev = sprev // predecessor of rnext is now s end
	return rnext
}

// Pop detaches r from its ring, leaving it linked only to itself.
// It returns r to permit method chaining.
func (r *Ring[T]) Pop() *Ring[T] {
	if r != nil && r.prev != r {
		rprev, rnext := r.prev, r.next
		rprev.next = r.next
		rnext.prev = r.prev
		r.prev = r
		r.next = r
	}
	return r
}

// Next returns the successor of r (which may be r itself).
// This will panic if r == nil.
func (r *Ring[T]) Next() *Ring[T] { return r.next }

// Prev returns the predecessor of r (which may be r itself).
// This will panic if r == nil.
func (r *Ring[T]) Prev() *Ring[T] { return r.prev }

// At returns the entry at offset n from r.  Negative values of n are
// permitted, and r.At(0) == r. If r == nil or the absolute value of n is
// greater than the length of the ring, At returns nil.
func (r *Ring[T]) At(n int) *Ring[T] {
	if r == nil {
		return nil
	}

	next := (*Ring[T]).Next
	if n < 0 {
		n = -n
		next = (*Ring[T]).Prev
	}

	cur := r
	for n > 0 {
		cur = next(cur)
		if cur == r {
			return nil
		}
		n--
	}
	return cur
}

// Peek reports whether the ring has a value at offset n from r, and if so
// returns its value. Negative values of n are permitted. If the absolute value
// of n is greater than the length of the ring, Peek reports a zero value.
func (r *Ring[T]) Peek(n int) (T, bool) {
	cur := r.At(n)
	if cur == nil {
		var zero T
		return zero, false
	}
	return cur.Value, true
}

// Each calls f with each value in r, in circular order. If f returns false,
// Each stops and returns false.  Otherwise, Each returns true after visiting
// all elements of r.
func (r *Ring[T]) Each(f func(v T) bool) bool {
	return scan(r, func(cur *Ring[T]) bool { return f(cur.Value) })
}

// Len reports the number of elements in r. If r == nil, Len is 0.
// This operation takes time proportional to the size of the ring.
func (r *Ring[T]) Len() int {
	if r == nil {
		return 0
	}
	var n int
	scan(r, func(*Ring[T]) bool { n++; return true })
	return n
}

// IsEmpty reports whether r is the empty ring.
func (r *Ring[T]) IsEmpty() bool { return r == nil }

func scan[T any](r *Ring[T], f func(*Ring[T]) bool) bool {
	if r == nil {
		return true
	}

	cur := r
	for f(cur) {
		if cur.next == r {
			return true
		}
		cur = cur.next
	}
	return false
}

func newRing[T any]() *Ring[T] { r := new(Ring[T]); r.next = r; r.prev = r; return r }
