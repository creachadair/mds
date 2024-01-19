package stack_test

import (
	"testing"

	"github.com/creachadair/mds/stack"
	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
)

func checkStack[T any](t *testing.T, s *stack.Stack[T], want []T) {
	t.Helper()
	got := s.Slice()
	if diff := cmp.Diff(want, got, cmpopts.EquateEmpty()); diff != "" {
		t.Errorf("Wrong contents (-got, +want):\n%s", diff)
	}
	if n := s.Len(); n != len(got) || n != len(want) {
		t.Errorf("Wrong length: got %d, want %d == %d", n, len(got), len(want))
	}
}

func TestStack(t *testing.T) {
	s := stack.New[int]()
	check := func(want ...int) { checkStack(t, s, want) }

	// Top and Pop of an empty stack report no value.
	if v := s.Top(); v != 0 {
		t.Errorf("Top: got %v, want 0", v)
	}
	if v, ok := s.Peek(0); ok || v != 0 {
		t.Errorf("Peek(0): got (%v, %v), want (0, false)", v, ok)
	}
	if v, ok := s.Pop(); ok {
		t.Errorf("Pop: got (%v, %v), want (0, false)", v, ok)
	}

	check()
	if !s.IsEmpty() {
		t.Error("IsEmpty is incorrectly false")
	}
	if n := s.Len(); n != 0 {
		t.Errorf("Len: got %d, want 0", n)
	}

	s.Push(1)
	if s.IsEmpty() {
		t.Error("IsEmpty is incorrectly true")
	}
	check(1)

	s.Push(2)
	check(2, 1)

	s.Push(3)
	check(3, 2, 1)
	if n := s.Len(); n != 3 {
		t.Errorf("Len: got %d, want 3", n)
	}

	top := s.Top()
	if top != 3 {
		t.Errorf("Top: got %v, want 3", top)
	}
	if v, ok := s.Peek(0); !ok || v != top {
		t.Errorf("Peek(0): got (%v, %v), want (%v, true)", v, ok, top)
	}
	if v, ok := s.Peek(1); !ok || v != 2 {
		t.Errorf("Peek(1): got (%v, %v), want (2, true)", v, ok)
	}
	if v, ok := s.Peek(10); ok {
		t.Errorf("Peek(10): got (%v, %v), want (0, false)", v, ok)
	}

	if v, ok := s.Pop(); !ok || v != top {
		t.Errorf("Pop: got (%v, %v), want (%v, true)", v, ok, top)
	}
	check(2, 1)

	s.Clear()
	check()
}
