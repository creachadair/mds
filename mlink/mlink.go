// Package mlink implements basic linked container data structures.
//
// The types defined here are not safe for concurrent use by multiple
// goroutines without external synchronization.
package mlink

// An entry is a singly-linked value container.
type entry[T any] struct {
	X    T
	link *entry[T]
}

// invalidate makes e and all its successor entries point to themselves, as a
// flag that they are detached from their original list and are invalid.
func (e *entry[T]) invalidate() {
	for e != nil {
		next := e.link
		e.link = e
		e = next
	}
}

// checkValid panics if e is an invalid entry, otherwise it returns e.
func (e *entry[T]) checkValid() *entry[T] {
	if e.link == e {
		panic("invalid cursor")
	}
	return e
}
