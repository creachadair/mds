// Package mtest is a support library for writing tests.
package mtest

import "testing"

// MustPanic executes a function f that is expected to panic.
// If it does so, MustPanic returns the value recovered from the
// panic. Otherwise, it logs a fatal error in t.
func MustPanic(t *testing.T, f func()) (val any) {
	t.Helper()
	defer func() { val = recover() }()
	f()
	t.Fatal("expected panic was not observed")
	return
}
