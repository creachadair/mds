// Package oset implements a set-like collection on ordered values.
//
// # Basic Operations
//
// Create an empty set with New or NewFunc. A zero-valued Set is ready for use
// as a read-only empty set, but it will panic if modified.
//
//	s := oset.New[string]()
//
// Alternatively, you can provide initial values at construction time, which in
// the current implementation is slightly more efficient for memory use:
//
//	s := oset.New("grape", "lemon", "banana", "kumquat")
//
// Add items using Add and remove items using Remove:
//
//	s.Add("apple")
//	s.Remove("pear")
//
// Look up items using Has. Report the number of elements in the set using Len.
//
// # Iterating in Order
//
// The elements of a Set can be traversed in order using an iterator.
// Construct an iterator for m by calling First or Last. The IsValid
// method reports whether the iterator has an element available, and
// the Next and Prev methods advance or retract the iterator:
//
//	for it := s.First(); it.IsValid(); it.Next() {
//	   doThingsWith(it.Value())
//	}
//
// Use the Seek method to seek to a particular point in the order.  Seek
// returns an iterator at the first element greater than or equal to the
// specified value:
//
//	for it := s.Seek("cherry"); it.IsValid(); it.Next() {
//	   doThingsWith(it.Value())
//	}
//
// Note that it is not safe to modify the set while iterating it.  If you
// modify a set while iterating it, you will need to re-synchronize any
// iterators after the edits, e.g.,
//
//	for it := s.First(); it.IsValid(); {
//	   if val := it.Value(); shouldDelete(val) {
//	      s.Remove(val)
//	      it.Seek(val) // update the iterator
//	   } else {
//	      it.Next()
//	   }
//	}
package oset

import (
	"cmp"
	"fmt"
	"strings"

	"github.com/creachadair/mds/stree"
)

// A Set represents a set of arbitrary values with an ordering.  It supports
// efficient insertion, deletion and lookup, and also allows values to be
// traversed in order.
//
// A zero Set behaves as an empty read-only set, whose Clear, Remove, Has,
// Slice, Len, First, Last, and Seek will work without error.  However, calling
// Add or AddAll on a zero Set will panic.
type Set[T any] struct {
	s *stree.Tree[T]
}

// New constructs a new empty Set using the natural comparison order for an
// ordered value type.  Copies of the set share storage.
func New[T cmp.Ordered](items ...T) Set[T] { return NewFunc[T](cmp.Compare, items...) }

// NewFunc constructs a new empty Set using cf to compare values.
// If cf == nil, NewFunc will panic.  Copies of the set share storage.
func NewFunc[T any](cf func(a, b T) int, items ...T) Set[T] {
	return Set[T]{s: stree.New(250, cf, items...)}
}

// String returns a string representation of the contents of s.
func (s Set[T]) String() string {
	if s.s == nil {
		return `oset[]`
	}
	var sb strings.Builder
	sb.WriteString("oset")

	tag := "["
	for it := s.First(); it.IsValid(); it.Next() {
		fmt.Fprint(&sb, tag, it.Value())
		tag = " "
	}
	sb.WriteString("]")
	return sb.String()
}

// IsEmpty reports whether s is empty.
func (s Set[T]) IsEmpty() bool { return s.s == nil || s.s.Len() == 0 }

// Len reports the number of elements in s.  This operation is constant-time.
func (s Set[T]) Len() int {
	if s.s == nil {
		return 0
	}
	return s.s.Len()
}

// Clear deletes all the elements from m, leaving it empty.
//
// This operation is constant-time.
func (s Set[T]) Clear() {
	if s.s != nil {
		s.s.Clear()
	}
}

// Clone returns a new set with the same contents as s.
func (s Set[T]) Clone() Set[T] {
	cp := s
	if s.s != nil {
		cp.s = s.s.Clone()
	}
	return cp
}

// Has reports whether value is present in the set.
func (s Set[T]) Has(value T) bool {
	if s.s == nil {
		return false
	}
	_, ok := s.s.Get(value)
	return ok
}

// Add adds the specified value to s, and returns s.
//
// This operation takes amortized O(lg n) time for each element, given a set
// with n elements.
func (s Set[T]) Add(values ...T) Set[T] {
	for _, v := range values {
		s.s.Add(v)
	}
	return s
}

// AddAll adds all the elements of set t to s and returns s.
func (s Set[T]) AddAll(t Set[T]) Set[T] {
	if t.s == nil {
		return s
	}
	t.s.Inorder(func(v T) bool {
		s.s.Add(v)
		return true
	})
	return s
}

// Remove removes the specified values from s and returns s.
func (s Set[T]) Remove(values ...T) Set[T] {
	if s.s == nil {
		return s
	}
	for _, v := range values {
		s.s.Remove(v)
	}
	return s
}

// RemoveAll removes all the elements of set t from s and returns s.
func (s Set[T]) RemoveAll(t Set[T]) Set[T] {
	if s.s != nil && t.s != nil {
		t.s.Inorder(func(v T) bool {
			s.s.Remove(v)
			return true
		})
	}
	return s
}

// Slice returns a slice of all the values in s, in order.
func (s Set[T]) Slice() []T {
	if s.s == nil || s.s.Len() == 0 {
		return nil
	}
	out := make([]T, 0, s.Len())
	s.s.Inorder(func(val T) bool {
		out = append(out, val)
		return true
	})
	return out
}

// First returns an iterator to the first element of the set, if any.
func (s Set[T]) First() *Iter[T] {
	it := &Iter[T]{s: s.s}
	if s.s != nil {
		it.c = s.s.Root().Min()
	}
	return it
}

// Last returns an iterator to the last element of the set, if any.
func (s Set[T]) Last() *Iter[T] {
	it := &Iter[T]{s: s.s}
	if s.s != nil {
		it.c = s.s.Root().Max()
	}
	return it
}

// Seek returns an iterator to the first element of s greater than or equal to
// value, if any; if not, the resulting iterator is invalid.
func (s Set[T]) Seek(value T) *Iter[T] { return s.First().Seek(value) }

// An Iter is an iterator for a Set.
type Iter[T any] struct {
	s *stree.Tree[T]
	c *stree.Cursor[T]
}

// IsValid reports whether it is pointing at an element of its set.
func (it *Iter[T]) IsValid() bool { return it.c.Valid() }

// Next advances it to the next element in the set, if any, and returns it.  If
// no such element exists, it becomes invalid.
func (it *Iter[T]) Next() *Iter[T] { it.c.Next(); return it }

// Prev advances it to the previous element in the set, if any, and returns
// it. If no such element exists, it becomes invalid.
func (it *Iter[T]) Prev() *Iter[T] { it.c.Prev(); return it }

// Value returns the current value, or a zero value if it is invalid.
func (it *Iter[T]) Value() T { return it.c.Key() }

// Seek advances it to the first element greater than or equal to value, and
// returns it.  If no such element exists, it becomes invalid.
func (it *Iter[T]) Seek(value T) *Iter[T] {
	it.c = nil
	if it.s != nil {
		it.s.InorderAfter(value, func(key T) bool {
			it.c = it.s.Cursor(key)
			return false
		})
	}
	return it
}
