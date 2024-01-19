package mlink

// A List is a singly-linked ordered list. A zero value is ready for use.
//
// The methods of a List value do not allow direct modification of the list.
// To insert and update entries in the list, use the At, Find, Last, or End
// methods to obtain a Cursor to a location in the list. A cursor can be used
// to insert, update, and delete elements of the list.
type List[T any] struct {
	first entry[T] // sentinel; first.link points to the real first element
}

// NewList returns a new empty list.
func NewList[T any]() *List[T] { return new(List[T]) }

// IsEmpty reports whether lst is empty.
func (lst *List[T]) IsEmpty() bool { return lst.first.link == nil }

// Clear discards all the values in lst, leaving it empty.  Calling Clear
// invalidates all cursors to the list.
func (lst *List[T]) Clear() { lst.first.link.invalidate(); lst.first.link = nil }

// Peek reports whether lst has a value at offset n from the front of the list,
// and if so returns its value.
//
// This method takes time proportional to n. Peek will panic if n < 0.
func (lst *List[T]) Peek(n int) (T, bool) {
	cur := lst.At(n)
	return cur.Get(), !cur.AtEnd()
}

// Each calls f with each value in lst, in order from first to last.
// If f returns false, Each stops and returns false.
// Otherwise, Each returns true after visiting all elements of lst.
func (lst *List[T]) Each(f func(T) bool) bool {
	for cur := lst.cfirst(); !cur.AtEnd(); cur.Next() {
		if !f(cur.Get()) {
			return false
		}
	}
	return true
}

// Len reports the number of elements in lst. This method takes time proportional
// to the length of the list.
func (lst *List[T]) Len() int {
	var n int
	lst.Each(func(T) bool { n++; return true })
	return n
}

// At returns a cursor to the element at index n ≥ 0 in the list.  If n is
// greater than or equal to n.Len(), At returns a cursor to the end of the list
// (equivalent to End).
//
// At will panic if n < 0.
func (lst *List[T]) At(n int) *Cursor[T] {
	if n < 0 {
		panic("index out of range")
	}

	cur := lst.cfirst()
	for ; !cur.AtEnd(); cur.Next() {
		if n == 0 {
			break
		}
		n--
	}
	return &cur
}

// Last returns a cursor to the last element of the list. If lst is empty, it
// returns a cursor to the end of the list (equivalent to End).
// This method takes time proportional to the length of the list.
func (lst *List[T]) Last() *Cursor[T] {
	cur := lst.cfirst()
	if cur.AtEnd() {
		return &cur
	}
	for !cur.AtEnd() {
		cur.Next()
	}
	cur.pred = cur.pred[:len(cur.pred)-1]
	return &cur
}

// End returns a cursor to the position just past the end of the list.
// This method takes time proportional to the length of the list.
func (lst *List[T]) End() *Cursor[T] { c := lst.Last(); c.Next(); return c }

// Find returns a cursor to the first element of the list for which f returns
// true. If no such element is found, the resulting cursor is at the end of the
// list.
func (lst *List[T]) Find(f func(T) bool) *Cursor[T] {
	cur := lst.cfirst()
	for !cur.AtEnd() {
		if f(cur.Get()) {
			break
		}
		cur.Next()
	}
	return &cur
}

func (lst *List[T]) cfirst() Cursor[T] { return Cursor[T]{pred: []*entry[T]{&lst.first}} }

// A Cursor represents a location in a list.  A nil *Cursor is not valid, and
// operations on it will panic. Through a valid cursor, the caller can navigate
// forward and backward, and add, modify, and remove elements.
//
// Multiple cursors into the same list are fine, but note that modifying the
// list through one cursor may invalidate others.
type Cursor[T any] struct {
	// pred is the sequence of entries from the front of the list to the target.
	// This permits the cursor to navigate in both directions in the list.
	//
	// Invariant: len(pred) != 0.
	pred []*entry[T]
}

func (c *Cursor[T]) last() *entry[T] { return c.pred[len(c.pred)-1].checkValid() }
func (c *Cursor[T]) popLast()        { c.pred = c.pred[:len(c.pred)-1]; c.last() }

// Get returns the value at c's location. If c is at the end of the list, Get
// returns a zero value.
func (c *Cursor[T]) Get() T {
	if c.AtEnd() {
		var zero T
		return zero
	}
	return c.last().link.X
}

// Set replaces the value at c's location. If c is at the end of the list,
// calling Set is equivalent to calling Push.
//
// Before:
//
//	[1, 2, 3]
//	    ^--- c
//
// After c.Set(9)
//
//	[1, 9, 3]
//	    ^--- c
func (c *Cursor[T]) Set(v T) {
	if c.AtEnd() {
		c.last().link = &entry[T]{X: v}
		// N.B.: c is now no longer AtEnd
	} else {
		c.last().link.X = v
	}
}

// AtEnd reports whether c is at the end of its list.
func (c *Cursor[T]) AtEnd() bool { return c.last().link == nil }

// Next advances c to the next position in the list if it is not at the end. If
// c was already at the end its position is unchanged. Next returns false if
// the resulting position is at the end of the list, otherwise true.
func (c *Cursor[T]) Next() bool {
	if c.AtEnd() {
		return false
	}
	c.pred = append(c.pred, c.last().link)
	return !c.AtEnd()
}

// Prev moves c to the previous position in the list if it is not at the
// front. If c was already at the front its position is unchanged. Prev returns
// false if the resulting position is at the front of the list, otherwise true.
func (c *Cursor[T]) Prev() bool {
	if len(c.pred) > 1 {
		c.popLast()
		return true
	}
	return false
}

// Push inserts a new value into the list at c's location. After insertion, c
// points to the newly-added item and the previous value is now at c.Next().
//
// Before:
//
//	[1, 2, 3]
//	 ^--- c
//
// After c.Push(4):
//
//	[4, 1, 2, 3]
//	 ^--- c
func (c *Cursor[T]) Push(v T) {
	last := c.last()
	added := &entry[T]{X: v, link: last.link}
	last.link = added
}

// Add inserts one or more new values into the list at c's location. After
// insertion, c points to the original item, now in the location after the
// newly-added values.  This is a shorthand for Push followed by Next.
//
// Before:
//
//	[1, 2, 3]
//	 ^--- c
//
// After c.Add(4):
//
//	[4, 1, 2, 3]
//	    ^--- c
func (c *Cursor[T]) Add(vs ...T) {
	for _, v := range vs {
		c.Push(v)
		c.Next()
	}
}

// Remove removes and returns the element at c's location from the list.  If c
// is at the end of the list, Remove does nothing and returns a zero value.
//
// After removal, c is still valid and points the element after the one that
// was removed, or the end of the list.
//
// Successfully removing an element invalidates any cursors to the location
// after the element that was removed.
//
// Before:
//
//	[1, 2, 3, 4]
//	    ^--- c
//
// After c.Remove()
//
//	[1, 3, 4]
//	    ^--- c
func (c *Cursor[T]) Remove() T {
	if c.AtEnd() {
		var zero T
		return zero
	}

	// Detach the discarded entry from its neighbor so that any cursors pointing
	// to that entry will be AtEnd, and changes made through them will not
	// affect the remaining list.
	last := c.last()
	out := last.link
	last.link = out.link // successor
	out.link = out       // invalidate the outgoing element
	return out.X
}

// Truncate removes all the elements of the list at and after c's location.
// After calling Truncate, c is at the end of the remaining list. If c is at
// the end of the list, Truncate does nothing. After truncation, c remains
// valid.
//
// Truncate invalidates any cursors to locations after c in the list.
//
// Before:
//
//	[1, 2, 3, 4]
//	       ^--- c
//
// After c.Truncate():
//
//	[1, 2] *
//	       ^--- c (c.AtEnd() == true)
func (c *Cursor[T]) Truncate() { last := c.last(); last.link.invalidate(); last.link = nil }

// Copy returns a copy of c pointing to the same location. Changes to c do not
// affect the copy and vice versa.
func (c *Cursor[T]) Copy() *Cursor[T] {
	cp := make([]*entry[T], len(c.pred))
	copy(cp, c.pred)
	return &Cursor[T]{pred: cp}
}
