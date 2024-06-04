package value

import "fmt"

// Maybe is a container that can hold a value of type T, which may either be
// present or absent. A zero value is ready for use, and is absent.
//
// It is safe to copy and assign a Maybe value, but note that if a value is
// present, only a shallow copy of the underlying value is made. Maybe values
// are comparable if and only if T is comparable.
type Maybe[T any] struct {
	value   T
	present bool
}

// Just returns a Maybe whose present value is v.
func Just[T any](v T) Maybe[T] { return Maybe[T]{value: v, present: true} }

// Check returns a present Maybe with value v if err == nil; otherwise it
// returns absent.
func Check[T any](v T, err error) Maybe[T] {
	if err == nil {
		return Just(v)
	}
	return Maybe[T]{}
}

// Absent returns an absent Maybe of the specified type.  It is a legibility
// notation equivalent to the zero value of type Maybe[T].
func Absent[T any]() Maybe[T] { return Maybe[T]{} }

// Present reports whether a value is present in m.
func (m Maybe[T]) Present() bool { return m.present }

// GetOK reports whether a value is present in m, and if so returns that value.
// If a value is not present, GetOK returns the zero of T.
func (m Maybe[T]) GetOK() (T, bool) { return m.value, m.present }

// Get returns the value present in m, if any; or else the zero value.
func (m Maybe[T]) Get() T { return m.value }

// Or returns the value of m if a value is present, otherwise it returns o.
func (m Maybe[T]) Or(o T) T {
	if m.present {
		return m.value
	}
	return o
}

// String returns the string representation of m.  If m is present, its string
// representation is that of the enclosed value.
func (m Maybe[T]) String() string {
	if m.present {
		return fmt.Sprint(m.value)
	}
	return fmt.Sprintf("Absent[%T]", m.value)
}

// MapMaybe returns a function from Maybe[T] to Maybe[U].
// If the argument is present and has value v, the result is present and has
// value f(v).  Otherwise, the result is absent and f is not called.
func MapMaybe[T, U any](f func(T) U) func(Maybe[T]) Maybe[U] {
	return func(a Maybe[T]) Maybe[U] {
		if v, ok := a.GetOK(); ok {
			return Just(f(v))
		}
		return Absent[U]()
	}
}

// First returns the first present value in vs, if any exists; otherwise it
// returns absent.
func First[T any](vs ...Maybe[T]) Maybe[T] {
	for _, v := range vs {
		if v.Present() {
			return v
		}
	}
	return Maybe[T]{}
}
