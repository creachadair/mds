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
	cmp  func(a, b T) int
	move func(T, int)
}

// nmove is a no-op move function used by default in a queue on which no update
// function has been set.
func nmove[T any](T, int) {}

// New constructs an empty [Queue] with the given comparison function, where
// cmp(a, b) must be <0 if a < b, =0 if a == b, and >0 if a > b.
func New[T any](cmp func(a, b T) int) *Queue[T] { return &Queue[T]{cmp: cmp, move: nmove[T]} }

// NewWithData constructs an empty [Queue] with the given comparison function
// that uses the given slice as storage.  This allows the caller to initialize
// a heap with existing data without copying, or to preallocate storage.  To
// preallocate storage without any initial values, pass a slice with length 0
// and the desired capacity.
//
// For example, to initialize a queue with fixed elements:
//
//	q := heapq.NewWithData(cfunc, []string{"u", "v", "w", "x", "y"})
//
// To initialize an empty queue with a pre-allocated buffer of n elements:
//
//	q := heapq.NewWithData(cfunc, make([]string, 0, n))
//
// The resulting queue takes ownership of the slice, and the caller must not
// access the contents data after the call until the queue is cleared.  Calling
// [Queue.Clear] will dissociate the queue from data.
func NewWithData[T any](cmp func(a, b T) int, data []T) *Queue[T] {
	q := &Queue[T]{data: data, cmp: cmp, move: nmove[T]}
	for i := len(q.data) / 2; i >= 0; i-- {
		q.pushDown(i)
	}
	return q
}

// SetUpdate sets u as the update function on q. This function is called
// whenever an element of the queue is moved to a new position, giving the
// value and its new position. If u == nil, an existing update function is
// removed.  SetUpdate returns q to allow chaining.
//
// Setting an update function makes q intrusive, allowing values in the queue
// to keep track of their current offset in the queue as items are added and
// removed. By default location information is not reported.
func (q *Queue[T]) SetUpdate(u func(T, int)) *Queue[T] {
	if u == nil {
		q.move = nmove[T]
	} else {
		q.move = u
	}
	return q
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

// Peek reports whether q has a value at offset n from the front of the queue,
// and if so returns its value. Peek(0) returns the same value as Front.  The
// order of elements at offsets n > 0 is unspecified.
//
// Peek will panic if n < 0.
func (q *Queue[T]) Peek(n int) (T, bool) {
	if n < 0 {
		panic("index out of range")
	} else if n >= len(q.data) {
		var zero T
		return zero, false
	}
	return q.data[n], true
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
	q.move(q.data[n], n)
	return q.pushUp(n)
}

// Remove reports whether q has a value at offset n from the front of the
// queue, and if so removes and returns it. Remove(0) is equivalent to Pop().
//
// Remove will panic if n < 0.
func (q *Queue[T]) Remove(n int) (T, bool) {
	if n < 0 {
		panic("index out of range")
	} else if n >= len(q.data) {
		var zero T
		return zero, false
	}
	return q.pop(n), true
}

// Set replaces the contents of q with the specified values. Any previous
// values in the queue are discarded. This operation takes time proportional to
// len(vs) to restore heap order. Set returns q to allow chaining.
func (q *Queue[T]) Set(vs []T) *Queue[T] {
	// Copy the values so we do not alias the original slice.
	// If the existing buffer already has enough space, reslice it; otherwise
	// allocate a fresh one.
	if cap(q.data) < len(vs) {
		q.data = make([]T, len(vs))
	} else {
		q.data = q.data[:len(vs)]
	}
	copy(q.data, vs)
	for i := len(q.data) - 1; i >= 0; i-- {
		q.move(q.data[i], i)
		q.pushDown(i)
	}
	return q
}

// Reorder replaces the ordering function for q with a new function. This
// operation takes time proportional to the length of the queue to restore the
// (new) heap order. The queue retains the same elements.
func (q *Queue[T]) Reorder(cmp func(a, b T) int) {
	q.cmp = cmp
	for i := len(q.data) / 2; i >= 0; i-- {
		q.pushDown(i)
	}
}

// Each is a range function that calls f with each value in q in heap order.
// If f returns false, Each returns immediately.
func (q *Queue[T]) Each(f func(T) bool) {
	for _, v := range q.data {
		if !f(v) {
			return
		}
	}
}

// Clear discards all the entries in q, leaving it empty. Clear releases all
// heap storage held by q, and further use of the queue will allocate fresh
// storage.
//
// If q was created by [NewWithData], calling Clear dissociates q from the
// provided slice, and subsequent
func (q *Queue[T]) Clear() {
	// Drop the slice entirely rather than reslicing it so that we do not pin
	// the array or any pointers it refers to from the GC. We could zero the
	// array and reslice, but then the caller has no way to recover control of a
	// delegated slice without overwriting its contents.
	q.data = nil
}

// pop removes and returns the value at index i of the heap, after restoring
// heap order. Precondition: i < len(q.data).
func (q *Queue[T]) pop(i int) T {
	out := q.data[i]
	n := len(q.data) - 1
	if n == 0 {
		q.data = q.data[:0]
	} else {
		q.data[i], q.data[n] = q.data[n], out
		q.move(q.data[i], i) // N.B. we do not report a move of out.
		q.data = q.data[:n]
		q.pushDown(i)
	}
	return out
}

// pushUp pushes the value at index i of the heap up until it is correctly
// ordered relative to its parent, and returns the resulting heap index.
func (q *Queue[T]) pushUp(i int) int {
	old := i
	for i > 0 {
		par := i / 2
		if q.cmp(q.data[i], q.data[par]) >= 0 {
			break
		}
		q.swap(i, par)
		i = par
	}
	// If the input moved, update its final position.
	if old != i {
		q.move(q.data[i], i)
	}
	return i
}

// pushDown pushes the value at index i of the heap down until it is correctly
// ordered relative to its children, and returns the resulting heap index.
func (q *Queue[T]) pushDown(i int) int {
	old := i
	lc := 2*i + 1
	for lc < len(q.data) {
		min := i
		if q.cmp(q.data[lc], q.data[min]) < 0 {
			min = lc
		}
		if rc := lc + 1; rc < len(q.data) && q.cmp(q.data[rc], q.data[min]) < 0 {
			min = rc
		}
		if min == i {
			break // no more work to do
		}
		q.swap(i, min)
		i, lc = min, 2*min+1
	}
	// If the input moved, update its final position.
	if i != old {
		q.move(q.data[i], i)
	}
	return i
}

// swap exchanges the elements at positions i and j of the heap, invoking the
// update function as needed.
func (q *Queue[T]) swap(i, j int) {
	q.data[i], q.data[j] = q.data[j], q.data[i]

	// Update the position of the item that was exchanged with the LHS
	// (originally at j), but not the new position of LHS itself (originally at
	// i). This avoids repeatedly updating LHS in the middle of a push up or
	// down until it lands in its final location.
	q.move(q.data[i], i)
}

// Sort reorders the contents of vs in-place using the heap-sort algorithm, in
// non-decreasing order by the comparison function provided.
func Sort[T any](cmp func(a, b T) int, vs []T) {
	if len(vs) < 2 {
		return
	}
	rcmp := func(a, b T) int { return -cmp(a, b) }
	q := NewWithData(rcmp, vs)
	for !q.IsEmpty() {
		q.Pop()
	}
}
