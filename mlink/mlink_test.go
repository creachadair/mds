package mlink_test

import (
	"testing"

	"github.com/creachadair/mds/mlink"
	"github.com/google/go-cmp/cmp"
)

type shared[T any] interface {
	Clear()
	Peek(int) (T, bool)
	Each(func(T) bool) bool
	IsEmpty() bool
	Len() int
}

var (
	_ shared[any] = (*mlink.Stack[any])(nil)
	_ shared[any] = (*mlink.Queue[any])(nil)
	_ shared[any] = (*mlink.List[any])(nil)
)

func checker(t *testing.T, obj shared[int]) func(want ...int) {
	return func(want ...int) {
		t.Helper()
		var got []int
		obj.Each(func(v int) bool {
			got = append(got, v)
			return true
		})
		if diff := cmp.Diff(want, got); diff != "" {
			t.Errorf("Wrong contents (-got, +want):\n%s", diff)
		}
		if n := obj.Len(); n != len(got) || n != len(want) {
			t.Errorf("Wrong length: got %d, want %d == %d", n, len(got), len(want))
		}
	}
}

func TestStack(t *testing.T) {
	s := mlink.NewStack[int]()
	check := checker(t, s)

	// Top and Pop of an empty stack report no value.
	if v, ok := s.Top(); ok {
		t.Errorf("Top: got (%v, %v), want (0, false)", v, ok)
	}
	if v, ok := s.Pop(); ok {
		t.Errorf("Pop: got (%v, %v), want (0, valse)", v, ok)
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

	top, ok := s.Top()
	if !ok || top != 3 {
		t.Errorf("Top: got (%v, %v), want (3, true)", top, ok)
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

func TestQueue(t *testing.T) {
	q := mlink.NewQueue[int]()
	check := checker(t, q)

	// Front and Pop of an empty queue report no value.
	if v, ok := q.Front(); ok {
		t.Errorf("Front: got (%v, %v), want (0, false)", v, ok)
	}
	if v, ok := q.Pop(); ok {
		t.Errorf("Pop: got (%v, %v), want (0, false)", v, ok)
	}

	check()
	if !q.IsEmpty() {
		t.Error("IsEmpty is incorrectly false")
	}
	if n := q.Len(); n != 0 {
		t.Errorf("Len: got %d, want 0", n)
	}

	q.Add(1)
	if q.IsEmpty() {
		t.Error("IsEmpty is incorrectly true")
	}
	check(1)

	q.Add(2)
	check(1, 2)

	q.Add(3)
	check(1, 2, 3)
	if n := q.Len(); n != 3 {
		t.Errorf("Len: got %d, want 3", n)
	}

	front, ok := q.Front()
	if !ok || front != 1 {
		t.Errorf("Top: got (%v, %v), want (1, true)", front, ok)
	}
	if v, ok := q.Peek(0); !ok || v != front {
		t.Errorf("Peek(0): got (%v, %v), want (%v, true)", v, ok, front)
	}
	if v, ok := q.Peek(1); !ok || v != 2 {
		t.Errorf("Peek(1): got (%v, %v), want (2, true)", v, ok)
	}
	if v, ok := q.Peek(10); ok {
		t.Errorf("Peek(10): got (%v, %v), want (0, false)", v, ok)
	}

	if v, ok := q.Pop(); !ok || v != front {
		t.Errorf("Pop: got (%v, %v), want (%v, true)", v, ok, front)
	}
	check(2, 3)

	q.Clear()
	check()
}
