// Package heapq implements a generic heap-structured priority queue.
package heapq

// A Queue is a heap-structured priority queue. The contents of a Queue are
// partially ordered, and the minimum element is accessible in constant time.
// Adding or removing an element has worst-case time complexity O(lg n).
//
// The order of elements in the Queue is determined by a comparison function
// provided when the queue is constructed.
type Queue[T any] struct {
	data []T
	less func(a, b T) bool
}

// New constructs an empty Queue with the given comparison function.
func New[T any](lessThan func(a, b T) bool) *Queue[T] { return &Queue[T]{less: lessThan} }

// NewSize constructs an empty Queue with the given comparison function, and a
// queue buffer pre-allocated to size slots.
func NewSize[T any](lessThan func(a, b T) bool, size int) *Queue[T] {
	return &Queue[T]{
		data: make([]T, 0, size),
		less: lessThan,
	}
}

// Len reports the number of elements in the queue. This is a constant-time operation.
func (q *Queue[T]) Len() int { return len(q.data) }

// IsEmpty reports whether the queue is empty.
func (q *Queue[T]) IsEmpty() bool { return len(q.data) == 0 }

// Front returns the frontmost element of the queue. If the queue is empty, it
// returns a zero value.
func (q *Queue[T]) Front() T {
	if len(q.data) == 0 {
		var zero T
		return zero
	}
	return q.data[0]
}

// Pop reports whether the queue contains any elements, and if so removes and
// returns the frontmost element.  It returns a zero value if q is empty.
func (q *Queue[T]) Pop() (T, bool) {
	if len(q.data) == 0 {
		var zero T
		return zero, false
	}
	return q.pop(0), true
}

// Add adds v to the queue. It returns the index in q where v is stored.
func (q *Queue[T]) Add(v T) int {
	n := len(q.data)
	q.data = append(q.data, v)
	return q.pushUp(n)
}

// Update replaces the item at index i of q with v, and restores heap order.
// It returns the new location of v. If i < 0 or i ≥ q.Len(), Update panics.
func (q *Queue[T]) Update(i int, v T) int {
	q.data[i] = v

	// At most one of these will move any elements of q.
	return q.pushDown(q.pushUp(i))
}

// Set replaces the contents of q with the specified values. Any previous
// values in the queue are discarded. This operation takes time proportional to
// len(vs) to restore heap order.
func (q *Queue[T]) Set(vs []T) {
	// Copy the values so we do not alias the original slice.
	// If the existing buffer already has enough space, reslice it; otherwise
	// allocate a fresh one.
	if cap(q.data) < len(vs) {
		q.data = make([]T, len(vs))
	} else {
		q.data = q.data[:len(vs)]
	}
	copy(q.data, vs)
	for i := len(q.data) / 2; i >= 0; i-- {
		q.pushDown(i)
	}
}

// Reorder replaces the ordering function for q with a new function. This
// operation takes time proportional to the length of the queue to restore the
// (new) heap order. The queue retains the same elements.
func (q *Queue[T]) Reorder(lessThan func(a, b T) bool) {
	q.less = lessThan
	for i := len(q.data) / 2; i >= 0; i-- {
		q.pushDown(i)
	}
}

// Each calls f for each value in q in heap order. If f returns false, Each
// stops and returns false. Otherwise, Each returns true after visiting all
// elements of q.
func (q *Queue[T]) Each(f func(T) bool) bool {
	for _, v := range q.data {
		if !f(v) {
			return false
		}
	}
	return true
}

// Clear discards all the entries in q, leaving it empty.
func (q *Queue[T]) Clear() { q.data = q.data[:0] }

// pop removes and returns the value at index i of the heap, after restoring
// heap order. Precondition: i < len(q.data).
func (q *Queue[T]) pop(i int) T {
	out := q.data[i]
	n := len(q.data) - 1
	if n == 0 {
		q.data = q.data[:0]
	} else {
		q.data[i] = q.data[n]
		q.data = q.data[:n]
		q.pushDown(i)
	}
	return out
}

// pushUp pushes the value at index i of the heap up until it is correctly
// ordered relative to its parent, and returns the resulting heap index.
func (q *Queue[T]) pushUp(i int) int {
	for i > 0 {
		par := i / 2
		if !q.less(q.data[i], q.data[par]) {
			break
		}
		q.data[i], q.data[par] = q.data[par], q.data[i]
		i = par
	}
	return i
}

// pushDown pushes the value at index i of the heap down until it is correctly
// ordered relative to its children, and returns the resulting heap index.
func (q *Queue[T]) pushDown(i int) int {
	lc := 2*i + 1
	for lc < len(q.data) {
		min := i
		if q.less(q.data[lc], q.data[min]) {
			min = lc
		}
		if rc := lc + 1; rc < len(q.data) && q.less(q.data[rc], q.data[min]) {
			min = rc
		}
		if min == i {
			break // no more work to do
		}
		q.data[i], q.data[min] = q.data[min], q.data[i]
		i, lc = min, 2*min+1
	}
	return i
}
