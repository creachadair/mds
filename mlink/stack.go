package mlink

// A Stack is a linked last-in, first-out sequence of values.
// A zero value is ready for use.
type Stack[T any] struct {
	list List[T]
	size int
}

// NewStack constructs a new empty stack.
func NewStack[T any]() *Stack[T] { return new(Stack[T]) }

// Push adds an entry for v to the top of s.
func (s *Stack[T]) Push(v T) { s.list.At(0).Push(v); s.size++ }

// IsEmpty reports whether s is empty.
func (s *Stack[T]) IsEmpty() bool { return s.list.IsEmpty() }

// Clear discards all the values in s, leaving it empty.
func (s *Stack[T]) Clear() { s.list.Clear(); s.size = 0 }

// Top reports whether s is non-empty, and if so returns its top value.
func (s *Stack[T]) Top() (T, bool) { return s.list.Peek(0) }

// Peek reports whether s has value at offset n from the top of the stack, and
// if so returns its value. Peek(0) is equivalent to Top.
//
// Peek will panic if n < 0.
func (s *Stack[T]) Peek(n int) (T, bool) { return s.list.Peek(n) }

// Pop reports whether s is non-empty, and if so it removes and returns its top
// value.
func (s *Stack[T]) Pop() (T, bool) {
	out, ok := s.list.Peek(0)
	if ok {
		s.list.At(0).Remove()
		s.size--
	}
	return out, ok
}

// Each calls f with each value in s, in order from newest to oldest.
// If f returns false, Each stops and returns false.
// Otherwise, Each returns true after visiting all elements of s.
func (s *Stack[T]) Each(f func(T) bool) bool { return s.list.Each(f) }

// Len reports the number of elements in s. This is a constant-time operation.
func (s *Stack[T]) Len() int { return s.size }
