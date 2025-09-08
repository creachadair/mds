// Package mtest is a support library for writing tests.
package mtest

import (
	"bytes"
	"strings"

	"github.com/creachadair/mds/mdiff"
)

// TB is the subset of the testing.TB interface used by this package.
type TB interface {
	Cleanup(func())
	Fatalf(string, ...any)
	Helper()
}

// MustPanic executes a function f that is expected to panic.
// If it does so, MustPanic returns the value recovered from the
// panic. Otherwise, it logs a fatal error in t.
func MustPanic(t TB, f func()) any {
	t.Helper()
	return MustPanicf(t, f, "expected panic was not observed")
}

// MustPanicf executes a function f that is expected to panic.  If it does so,
// MustPanicf returns the value recovered from the panic. Otherwise it logs a
// fatal error in t.
func MustPanicf(t TB, f func(), msg string, args ...any) (val any) {
	t.Helper()
	defer func() { val = recover() }()
	f()
	t.Fatalf(msg, args...)
	return
}

// Swap replaces the target of p with v, and restores the original value when
// the governing test exits. It returns the original value.
func Swap[T any](t TB, p *T, v T) T {
	save := *p
	*p = v
	t.Cleanup(func() { *p = save })
	return save
}

// DiffLines returns a unified diff in textual format of the line-oriented
// difference between got and want, or "" if the strings are equal.
func DiffLines(got, want string) string {
	d := mdiff.New(strings.Split(got, "\n"), strings.Split(want, "\n"))
	if len(d.Chunks) == 0 {
		return ""
	}
	var buf bytes.Buffer
	d.AddContext(3).Unify().Format(&buf, mdiff.Unified, nil)
	return buf.String()
}
