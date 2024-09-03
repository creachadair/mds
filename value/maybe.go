package value

import "fmt"

// Maybe is a container that can hold a value of type T.
// Just(v) returns a Maybe holding the value v.
// Absent() returns a Maybe that holds no value.
// A zero Maybe is ready for use and is equivalent to Absent().
//
// It is safe to copy and assign a Maybe value, but note that if a value is
// present, only a shallow copy of the underlying value is made. Maybe values
// are comparable if and only if T is comparable.
type Maybe[T any] struct {
	value   T
	present bool
}

// Just returns a Maybe holding the value v.
func Just[T any](v T) Maybe[T] { return Maybe[T]{value: v, present: true} }

// Absent returns a Maybe holding no value.
// A zero Maybe is equivalent to Absent().
func Absent[T any]() Maybe[T] { return Maybe[T]{} }

// Present reports whether m holds a value.
func (m Maybe[T]) Present() bool { return m.present }

// GetOK reports whether m holds a value, and if so returns that value.
// If m is empty, GetOK returns the zero of T.
func (m Maybe[T]) GetOK() (T, bool) { return m.value, m.present }

// Get returns value held in m, if present; otherwise it returns the zero of T.
func (m Maybe[T]) Get() T { return m.value }

// Or returns m if m holds a value; otherwise it returns Just(o).
func (m Maybe[T]) Or(o T) Maybe[T] {
	if m.present {
		return m
	}
	return Just(o)
}

// Ptr converts m to a pointer. It returns nil if m is empty, otherwise it
// returns a pointer to a location containing the value held in m.
func (m Maybe[T]) Ptr() *T {
	if m.present {
		return &m.value
	}
	return nil
}

// String returns the string representation of m. If m holds a value v, the
// string representation of m is that of v.
func (m Maybe[T]) String() string {
	if m.present {
		return fmt.Sprint(m.value)
	}
	return fmt.Sprintf("Absent[%T]", m.value)
}

// Check returns Just(v) if err == nil; otherwise it returns Absent().
func Check[T any](v T, err error) Maybe[T] {
	if err == nil {
		return Just(v)
	}
	return Maybe[T]{}
}
