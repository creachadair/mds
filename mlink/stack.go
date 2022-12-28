package mlink

// A Stack is a last-in, first-out sequence of values.
// A zero value is ready for use.
type Stack[T any] struct {
	list []T
}

// NewStack constructs a new empty stack.
func NewStack[T any]() *Stack[T] { return new(Stack[T]) }

// Push adds an entry for v to the top of s.
func (s *Stack[T]) Push(v T) { s.list = append(s.list, v) }

// Add is a synonym for Push.
func (s *Stack[T]) Add(v T) { s.list = append(s.list, v) }

// IsEmpty reports whether s is empty.
func (s *Stack[T]) IsEmpty() bool { return len(s.list) == 0 }

// Clear discards all the values in s, leaving it empty.
func (s *Stack[T]) Clear() { s.list = s.list[:0] }

// Top returns the top element of the stack. If the stack is empty, it returns
// a zero value.
func (s *Stack[T]) Top() T {
	if len(s.list) == 0 {
		var zero T
		return zero
	}
	return s.list[len(s.list)-1]
}

// Peek reports whether s has value at offset n from the top of the stack, and
// if so returns its value. Peek(0) returns the same value as Top.
//
// Peek will panic if n < 0.
func (s *Stack[T]) Peek(n int) (T, bool) {
	if n >= len(s.list) {
		var zero T
		return zero, false
	}
	return s.list[len(s.list)-1-n], true
}

// Pop reports whether s is non-empty, and if so it removes and returns its top
// value.
func (s *Stack[T]) Pop() (T, bool) {
	out, ok := s.Peek(0)
	if ok {
		s.list = s.list[:len(s.list)-1]
	}
	return out, ok
}

// Each calls f with each value in s, in order from newest to oldest.
// If f returns false, Each stops and returns false.
// Otherwise, Each returns true after visiting all elements of s.
func (s *Stack[T]) Each(f func(T) bool) bool {
	for i := len(s.list) - 1; i >= 0; i-- {
		if !f(s.list[i]) {
			return false
		}
	}
	return true
}

// Len reports the number of elements in s. This is a constant-time operation.
func (s *Stack[T]) Len() int { return len(s.list) }
