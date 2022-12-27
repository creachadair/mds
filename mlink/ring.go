package mlink

// A Ring is a doubly-linked circular chain of one or more data items.
// There is no designated beginning of a ring; each element in the chain is a
// valid entry point for the entire ring.
type Ring[T any] struct {
	Value T

	prev, next *Ring[T]
}

// NewRing constructs a new ring containing only the item given.
func NewRing[T any](item T) *Ring[T] {
	r := &Ring[T]{Value: item}
	r.prev = r
	r.next = r
	return r
}

// Adjoin splices ring s into r.  If r = [r1 ... rn] and s = [s1 ... sm], the
// resulting ring is
//
//	rs = [r1 s1 ... sm r2 ... rn]
//
// where r remains at r1 and s remains at s1.
func (r *Ring[T]) Adjoin(s *Ring[T]) {
	rnext, sprev := r.next, s.prev

	r.next = s         // successor of r is now s
	s.prev = r         // predecessor of s is now r
	sprev.next = rnext // successor of s end is now rnext
	rnext.prev = sprev // predecessor of rnext is now s end
}

// Add splices vs into r. If r = [r1 ... 4n] and vs = [v1 ... vm], the
// resulting ring is
//
//	r = [r1 v1 ... vm r2 ... rn]
//
// where r rmains at r1.
func (r *Ring[T]) Add(vs ...T) *Ring[T] {
	for i := len(vs) - 1; i >= 0; i-- {
		r.Adjoin(NewRing(vs[i]))
	}
	return r
}

// Pop detaches r from its ring, leaving it linked only to itself.
// It returns r to permit method chaining.
func (r *Ring[T]) Pop() *Ring[T] {
	if r.prev != r {
		rprev, rnext := r.prev, r.next
		rprev.next = r.next
		rnext.prev = r.prev
		r.prev = r
		r.next = r
	}
	return r
}

// Next returns the successor of r (which may be r itself).
func (r *Ring[T]) Next() *Ring[T] { return r.next }

// Prev returns the predecessor of r (which may be r itself).
func (r *Ring[T]) Prev() *Ring[T] { return r.prev }

// Each calls f with each value in r, in circular order. If f returns false,
// Each stops and returns false.  Otherwise, Each returns true after visiting
// all elements of r.
func (r *Ring[T]) Each(f func(v T) bool) bool {
	cur := r
	for f(cur.Value) {
		if cur.next == r {
			return true
		}
		cur = cur.next
	}
	return false
}
