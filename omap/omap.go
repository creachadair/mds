// Package omap implements a map-like collection on ordered keys.
//
// # Basic Operations
//
// Create an empty map with New or NewFunc. A zero-valued Map is ready for use
// as a read-only empty map, but it will panic if modified.
//
//	m := omap.New[string, int]()
//
// Add items using Set and remove items using Delete:
//
//	m.Set("apple", 1)
//	m.Delete("pear")
//
// Look up items using Get and GetOK:
//
//	v := m.Get(key)        // returns a zero value if key not found
//	v, ok := m.GetOK(key)  // ok indicates whether key was found
//
// Report the number of elements in the map using Len.
//
// # Iterating in Order
//
// The elements of a map can be traversed in order using an iterator.
// Construct an iterator for m by calling First or Last. The IsValid
// method reports whether the iterator has an element available, and
// the Next and Prev methods advance or retract the iterator:
//
//	for it := m.First(); it.IsValid(); it.Next() {
//	   doThingsWith(it.Key(), it.Value())
//	}
//
// Use the Seek method to seek to a particular point in the order.  Seek
// returns an iterator at the first item greater than or equal to the specified
// key:
//
//	for it := m.Seek("cherry"); it.IsValid(); it.Next() {
//	   doThingsWith(it.Key(), it.Value())
//	}
//
// Note that it is not safe to modify the map while iterating it.  If you
// modify a map while iterating it, you will need to re-synchronize any
// iterators after the edits, e.g.,
//
//	for it := m.First(); it.IsValid(); {
//	   if key := it.Key(); shouldDelete(key) {
//	      m.Delete(key)
//	      it.Seek(key) // update the iterator
//	   } else {
//	      it.Next()
//	   }
//	}
package omap

import (
	"cmp"
	"fmt"
	"strings"

	"github.com/creachadair/mds/stree"
)

// A Map represents a mapping over arbitrary key and value types. It supports
// efficient insertion, deletion and lookup, and also allows keys to be
// traversed in order.
//
// A zero Map behaves as an empty read-only map, and Clear, Delete, Get, Keys,
// Len, First, and Last will work without error; however, calling Set on a zero
// Map will panic.
type Map[T, U any] struct {
	m *stree.Tree[stree.KV[T, U]]
}

// New constructs a new empty Map using the natural comparison order for an
// ordered key type.  Copies of the map share storage.
func New[T cmp.Ordered, U any]() Map[T, U] { return NewFunc[T, U](cmp.Compare) }

// NewFunc constructs a new empty Map using cf to compare keys.  If cf == nil,
// NewFunc will panic.  Copies of the map share storage.
func NewFunc[T, U any](cf func(a, b T) int) Map[T, U] {
	type kv = stree.KV[T, U]
	return Map[T, U]{m: stree.New(250, kv{}.Compare(cf))}
}

// String returns a string representation of the contents of m.
func (m Map[T, U]) String() string {
	if m.m == nil {
		return `omap[]`
	}
	var sb strings.Builder
	sb.WriteString("omap[")

	sp := "%v:%v"
	for it := m.First(); it.IsValid(); it.Next() {
		fmt.Fprintf(&sb, sp, it.Key(), it.Value())
		sp = " %v:%v"
	}
	sb.WriteString("]")
	return sb.String()
}

// Len reports the number of key-value pairs in m.
// This operation is constant-time.
func (m Map[T, U]) Len() int {
	if m.m == nil {
		return 0
	}
	return m.m.Len()
}

// Get returns the value associated with key in m if it is present, or returns
// a zero value. To check for presence, use GetOK.
func (m Map[T, U]) Get(key T) U { u, _ := m.GetOK(key); return u }

// GetOK reports whether key is present in m, and if so returns the value
// associated with it, or otherwise a zero value.
//
// This operation takes O(lg n) time for a map with n elements.
func (m Map[T, U]) GetOK(key T) (U, bool) {
	if m.m != nil {
		kv, ok := m.m.Get(stree.KV[T, U]{Key: key})
		if ok {
			return kv.Value, true
		}
	}
	var zero U
	return zero, false
}

// Set adds or replaces the value associated with key in m, and reports whether
// the key was new (true) or updated (false).
//
// This operation takes amortized O(lg n) time for a map with n elements.
func (m Map[T, U]) Set(key T, value U) bool {
	return m.m.Replace(stree.KV[T, U]{Key: key, Value: value})
}

// Delete deletes the specified key from m, and reports whether it was present.
//
// This operation takes amortized O(lg n) time for a map with n elements.
func (m Map[T, U]) Delete(key T) bool {
	if m.m == nil {
		return false
	}
	return m.m.Remove(stree.KV[T, U]{Key: key})
}

// Clear deletes all the elements from m, leaving it empty.
//
// This operation is constant-time.
func (m Map[T, U]) Clear() {
	if m.m != nil {
		m.m.Clear()
	}
}

// Keys returns a slice of all the keys in m, in order.
func (m Map[T, U]) Keys() []T {
	if m.m == nil || m.m.Len() == 0 {
		return nil
	}
	out := make([]T, 0, m.Len())
	m.m.Inorder(func(kv stree.KV[T, U]) bool {
		out = append(out, kv.Key)
		return true
	})
	return out
}

// First returns an iterator to the first entry of the map, if any.
func (m Map[T, U]) First() *Iter[T, U] {
	it := &Iter[T, U]{m: m.m}
	if m.m != nil {
		it.c = m.m.Root().Min()
	}
	return it
}

// Last returns an iterator to the last entry of the map, if any.
func (m Map[T, U]) Last() *Iter[T, U] {
	it := &Iter[T, U]{m: m.m}
	if m.m != nil {
		it.c = m.m.Root().Max()
	}
	return it
}

// Seek returns an iterator to the first entry of the map whose key is greater
// than or equal to key, if any.
func (m Map[T, U]) Seek(key T) *Iter[T, U] { return m.First().Seek(key) }

// An Iter is an iterator for a Map.
type Iter[T, U any] struct {
	m *stree.Tree[stree.KV[T, U]]
	c *stree.Cursor[stree.KV[T, U]]
}

// IsValid reports whether it is pointing at an element of its map.
func (it *Iter[T, U]) IsValid() bool { return it.c.Valid() }

// Next advances it to the next element in the map, if any.
func (it *Iter[T, U]) Next() *Iter[T, U] { it.c.Next(); return it }

// Prev advances it to the previous element in the map, if any.
func (it *Iter[T, U]) Prev() *Iter[T, U] { it.c.Prev(); return it }

// Key returns the current key, or a zero key if it is invalid.
func (it *Iter[T, U]) Key() T { return it.c.Key().Key }

// Value returns the current value, or a zero value if it is invalid.
func (it *Iter[T, U]) Value() U { return it.c.Key().Value }

// Seek advances it to the first key greater than or equal to key.
// If no such key exists, it becomes invalid.
func (it *Iter[T, U]) Seek(key T) *Iter[T, U] {
	it.c = nil
	if it.m != nil {
		it.m.InorderAfter(stree.KV[T, U]{Key: key}, func(kv stree.KV[T, U]) bool {
			it.c = it.m.Cursor(kv)
			return false
		})
	}
	return it
}
