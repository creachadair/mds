// Package mtest is a support library for writing tests.
package mtest

// TB is the subset of the testing.TB interface used by this package.
type TB interface {
	Helper()
	Fatal(...any)
	Fatalf(string, ...any)
}

// MustPanic executes a function f that is expected to panic.
// If it does so, MustPanic returns the value recovered from the
// panic. Otherwise, it logs a fatal error in t.
func MustPanic(t TB, f func()) (val any) {
	t.Helper()
	defer func() { val = recover() }()
	f()
	t.Fatal("expected panic was not observed")
	return
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
