// Package queue implements an array-based FIFO queue.
package queue

import (
	"github.com/creachadair/mds/slice"
)

// Queue is an array-based first-in, first-out sequence of values.
// A zero Queue is ready for use.
//
// Add and Pop operations take amortized O(1) time and storage.
// All other operations on a Queue are constant time.
type Queue[T any] struct {
	vs   []T
	head int
	n    int
}

// New constructs a new empty queue.
func New[T any]() *Queue[T] { return new(Queue[T]) }

// NewSize constructs a new empty queue with storage pre-allocated for n items.
// The queue will automatically grow beyond the initial size as needed.
func NewSize[T any](n int) *Queue[T] { return &Queue[T]{vs: make([]T, n)} }

// Add adds v to the end of q.
func (q *Queue[T]) Add(v T) {
	if q.n < len(q.vs) {
		// We have spaces left in the buffer.
		pos := (q.head + q.n) % len(q.vs)
		q.n++
		q.vs[pos] = v
		return
	} else if q.head > 0 {
		// Shift the existing items to initial position so that the append below
		// can handle extending the buffer. This costs O(1) space, O(n) time; but
		// we amortize this against the allocation we're (probably) going to do.
		slice.Rotate(q.vs, -q.head)
		q.head = 0
	}

	// The buffer is in the initial regime, head == 0.
	w := append(q.vs, v)
	q.vs = w[:cap(w)]
	q.n++
}

// IsEmpty reports whether q is empty.
func (q *Queue[T]) IsEmpty() bool { return q.n == 0 }

// Len reports the number of entries in q.
func (q *Queue[T]) Len() int { return q.n }

// Clear discards all the values in q, leaving it empty.
func (q *Queue[T]) Clear() { q.vs, q.head, q.n = nil, 0, 0 }

// Front returns the frontmost (oldest) element of q.  If q is empty, Front
// returns a zero value.
func (q *Queue[T]) Front() T {
	if q.n == 0 {
		var zero T
		return zero
	}
	return q.vs[q.head]
}

// Peek reports whether q has a value at offset n from the front of the queue,
// and if so returns its value. Peek(0) returns the same value as Front.
func (q *Queue[T]) Peek(n int) (T, bool) {
	if n < 0 {
		panic("index out of range")
	} else if n >= q.n {
		var zero T
		return zero, false
	}
	p := (q.head + n) % len(q.vs)
	return q.vs[p], true
}

// Pop reports whether q is non-empty, and if so removes and returns its
// frontmost (oldest) value. If q is empty, Pop returns a zero value.
func (q *Queue[T]) Pop() (T, bool) {
	if q.n == 0 {
		var zero T
		return zero, false
	}
	out := q.vs[q.head]
	q.n--
	if q.n == 0 {
		q.head = 0 // reset to initial conditions
	} else {
		q.head = (q.head + 1) % len(q.vs)
	}
	return out, true
}

// Each calls f with each value in q, in order from oldest to newest.
// If f returns false, Each stops and returns false.
// Otherwise, Each returns true after visiting all elements of q.
func (q *Queue[T]) Each(f func(T) bool) bool {
	cur := q.head
	for i := 0; i < q.n; i++ {
		if !f(q.vs[cur]) {
			return false
		}
		cur = (cur + 1) % len(q.vs)
	}
	return true
}

// Slice returns a slice of the values of q in order from oldest to newest.
// If q is empty, Slice returns nil.
func (q *Queue[T]) Slice() []T {
	if q.n == 0 {
		return nil
	}
	buf := make([]T, q.n)
	cur := q.head
	for i := 0; i < q.n; i++ {
		buf[i] = q.vs[cur]
		cur = (cur + 1) % len(q.vs)
	}
	return buf
}

/*
  A queue is an expanding ring buffer with amortized O(1) access.

  The queue tracks a buffer (buf) and two values, the head (H) is the offset of
  the oldest item in the queue (if any), and the length (n) is the number of
  queue entries.

  Initially the queue is empty, n = 0 and H = 0.

  As long as there is unused space, n < len(buf), we can add to the queue by
  simply bumping the length and storing the item in the next unused slot.

  When items are removed from the queue, H moves forward, leaving spaces at the
  beginning of the ring:

  * * * d e f g h i
  - - - - - - - - -
        H

  In this regime, a new item (j) wraps around and consumes an empty slot:

  j * * d e f g h i
  - - - - - - - - -
  ^     H

  If the queue is empty after removing an item (n = 0, we can reset to the
  initial condition by setting H = 0, since it no longer matters where H is
  when there are no values.

  Once the buffer fills (n = len(buf)), there are two cases to consider: In the
  simple case, when H == 0, new items are appended to the end, extending the
  buffer (and the append function handles amortized allocation for us):

  a b c d e f g
  - - - - - - -
  H             ^ next

  On the other hand, if H > 0, we cannot append directly to the end of buf,
  because that will put it out of order with respect to the offsets < H.  To
  fix this, we rotate the contents of buf forward so that H = 0 again, at which
  point we can now safely append again:

  1. Before insert, the buffer is full with H > 0:

    j k l d e f g h i
    - - - - - - - - -
          H

  2. Rotate the elements down to offset 0. This can be done in O(n) time
     in-place by chasing the cycles of the rotation:

    < < < rotate

    d e f g h i j k l
    - - - - - - - - -
    H

  At this point we are back in the initial regime. We can append m to grow buf:

    d e f g h i j k l m
    - - - - - - - - - -
    H
*/
