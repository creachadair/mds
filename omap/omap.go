// Package omap implements a map-like collection on ordered keys.
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
// Len, Range, and RangeAfter will work without error; however, calling Set on
// a zero Map will panic.
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
	m.Range(func(key T, val U) bool {
		fmt.Fprintf(&sb, sp, key, val)
		sp = " %v:%v"
		return true
	})
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

// Range iterates all the key-value pairs in m in key order.  If f returns
// false, Range stops and returns false. Otherwise, Range returns true after
// visiting all the entries in m.
func (m Map[T, U]) Range(f func(key T, value U) bool) bool {
	if m.m == nil {
		return true
	}
	return m.m.Inorder(func(kv stree.KV[T, U]) bool {
		return f(kv.Key, kv.Value)
	})
}

// RangeAfter iterates all the key-value pairs in m in key order, whose key is
// greater than or equal to start. If f returns false, RangeAfter stops and
// returns false. Otherwise, RangeAfter returns true after visiting all the
// eligible entries in m.
func (m Map[T, U]) RangeAfter(start T, f func(key T, value U) bool) bool {
	if m.m == nil {
		return true
	}
	return m.m.InorderAfter(stree.KV[T, U]{Key: start}, func(kv stree.KV[T, U]) bool {
		return f(kv.Key, kv.Value)
	})
}

// Keys returns a slice of all the keys in m, in order.
func (m Map[T, U]) Keys() []T {
	if m.m == nil || m.m.Len() == 0 {
		return nil
	}
	out := make([]T, 0, m.Len())
	m.Range(func(key T, _ U) bool {
		out = append(out, key)
		return true
	})
	return out
}
