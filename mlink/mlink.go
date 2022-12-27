// Package mlink implements basic linked container data structures.
//
// Most types in this package share certain common behaviors:
//
//   - A Clear method that discards all the contents of the container.
//   - A Peek method that returns an order statistic of the container.
//   - An Each method that iterates the container in its natural order.
//   - An IsEmpty method that reports whether the container is empty.
//   - A Len method that reports the number of elements in the container.
//
// The types defined here are not safe for concurrent use by multiple
// goroutines without external synchronization.
package mlink

// An entry is a singly-linked value container.
type entry[T any] struct {
	X    T
	link *entry[T]
}
