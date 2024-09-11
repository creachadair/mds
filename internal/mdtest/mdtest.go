// Package mdtest includes some internal utilities for testing.
package mdtest

import (
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
)

// Shared is the shared interface implemented by various types in this module.
// It is defined here for use in interface satisfaction checks in tests.
type Shared[T any] interface {
	Clear()
	Peek(int) (T, bool)
	Each(func(T) bool)
	IsEmpty() bool
	Len() int
}

// Eacher is the subset of Shared provided by iterable elements.
type Eacher[T any] interface {
	Each(func(T) bool)
	Len() int
}

// CheckContents verifies that s contains the specified elements in order, or
// reports an error to t.
func CheckContents[T any](t *testing.T, s Eacher[T], want []T) {
	t.Helper()
	var got []T
	for v := range s.Each {
		got = append(got, v)
	}
	if diff := cmp.Diff(want, got, cmpopts.EquateEmpty()); diff != "" {
		t.Errorf("Wrong contents (-got, +want):\n%s", diff)
	}
	if n := s.Len(); n != len(got) || n != len(want) {
		t.Errorf("Wrong length: got %d, want %d == %d", n, len(got), len(want))
	}
}
