package mlink

// A Queue is a linked first-in, first out sequence of values.  A zero value is
// ready for use.
type Queue[T any] struct {
	list List[T]
	back Cursor[T]
	size int
}

// NewQueue returns a new empty FIFO queue.
func NewQueue[T any]() *Queue[T] {
	q := new(Queue[T])
	q.back = q.list.cfirst()
	return q
}

// Add adds v to the end of q.
func (q *Queue[T]) Add(v T) {
	if q.back.pred == nil {
		q.back = q.list.cfirst()
	}
	q.back.Add(v)
	q.size++
}

// IsEmpty reports whether q is empty.
func (q *Queue[T]) IsEmpty() bool { return q.list.IsEmpty() }

// Clear discards all the values in q, leaving it empty.
func (q *Queue[T]) Clear() { q.list.Clear(); q.back = q.list.cfirst(); q.size = 0 }

// Front returns the frontmost (oldest) element of the queue. If the queue is
// empty, it returns a zero value.
func (q *Queue[T]) Front() T { v, _ := q.list.Peek(0); return v }

// Peek reports whether q has a value at offset n from the front of the queue,
// and if so returns its value. Peek(0) returns the same value as Front.
//
// Peek will panic if n < 0.
func (q *Queue[T]) Peek(n int) (T, bool) { return q.list.Peek(n) }

// Pop reports whether q is non-empty, and if so removes and returns its
// frontmost (oldest) value.
func (q *Queue[T]) Pop() (T, bool) {
	cur := q.list.cfirst()
	out := cur.Get()
	if cur.AtEnd() {
		return out, false
	}
	cur.Remove()
	q.size--
	if q.list.IsEmpty() {
		q.back = q.list.cfirst()
	}
	return out, true
}

// Each calls f with each value in q, in order from oldest to newest.
// If f returns false, Each stops and returns false.
// Otherwise, Each returns true after visiting all elements of q.
func (q *Queue[T]) Each(f func(T) bool) bool { return q.list.Each(f) }

// Len reports the number of elements in q. This is a constant-time operation.
func (q *Queue[T]) Len() int { return q.size }
