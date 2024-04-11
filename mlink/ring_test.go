package mlink_test

import (
	"testing"

	"github.com/creachadair/mds/internal/mdtest"
	"github.com/creachadair/mds/mlink"
)

func rc[T any](t *testing.T, r *mlink.Ring[T], want ...T) {
	mdtest.CheckContents(t, r, want)
}

func TestRing(t *testing.T) {
	t.Run("Initialize", func(t *testing.T) {
		rc[int](t, nil)
		rc(t, mlink.NewRing[int](0))
		rc(t, mlink.RingOf[int]())
		rc(t, mlink.NewRing[int](4), 0, 0, 0, 0)
		rc(t, mlink.RingOf(0), 0)
	})

	t.Run("Joining", func(t *testing.T) {
		r := mlink.NewRing[int](1)
		r.Join(mlink.RingOf(1, 2, 3))
		rc(t, r, 0, 1, 2, 3)

		// Joining r to itself should do nothing.
		r.Join(r)
		rc(t, r, 0, 1, 2, 3)

		// Test adding items to various places in the ring.
		r.Next().Join(mlink.RingOf(4, 5, 6))
		rc(t, r, 0, 1, 4, 5, 6, 2, 3)
		r.Prev().Join(mlink.RingOf(7))
		rc(t, r, 0, 1, 4, 5, 6, 2, 3, 7)
	})

	t.Run("Popping", func(t *testing.T) {
		rc(t, mlink.RingOf(1).Pop(), 1)
		q := mlink.RingOf(2, 3, 5, 7, 11)
		rc(t, q, 2, 3, 5, 7, 11)
		q.Next().Pop()
		rc(t, q, 2, 5, 7, 11)
		q.Prev().Pop()
		rc(t, q, 2, 5, 7)
	})

	t.Run("Circularity", func(t *testing.T) {
		r := mlink.RingOf(1, 3, 5, 7, 9)
		rc(t, r, 1, 3, 5, 7, 9)
		rc(t, r.Next().Next(), 5, 7, 9, 1, 3)
		rc(t, r.Prev(), 9, 1, 3, 5, 7)
		r.Next().Join(r.Prev())
		rc(t, r, 1, 3, 9)
	})

	t.Run("SplicingIn", func(t *testing.T) {
		s := mlink.RingOf(10, 20, 30)
		rc(t, s, 10, 20, 30)
		r := mlink.RingOf(1, 2, 3, 4, 5, 6)

		x := r.Next().Join(s)
		rc(t, r, 1, 2, 10, 20, 30, 3, 4, 5, 6)
		//       ^- r  ^- s        ^- x
		rc(t, s, 10, 20, 30, 3, 4, 5, 6, 1, 2)
		//       ^- s        ^- x        ^- r
		rc(t, x, 3, 4, 5, 6, 1, 2, 10, 20, 30)
		//       ^- x        ^- r  ^- s
	})

	t.Run("SplicingOut", func(t *testing.T) {
		r := mlink.RingOf(1, 20, 30, 40, 5, 6)
		tail := r.At(4)
		rc(t, r.Join(tail), 20, 30, 40) // just the excised part
		rc(t, r, 1, 5, 6)

		rc(t, r.Join(r.Next())) // nothing was removed
		rc(t, r, 1, 5, 6)

		q := mlink.RingOf("fat", "cats", "get", "dizzy", "after", "eating", "beans")
		s := q.At(2).Join(q.Prev())

		rc(t, q, "fat", "cats", "get", "beans")
		rc(t, s, "dizzy", "after", "eating")
	})

	t.Run("Peek", func(t *testing.T) {
		r := mlink.RingOf("kingdom", "phylum", "class", "order", "family", "genus", "species")
		checkPeek := func(n int, want string, wantok bool) {
			t.Helper()
			got, ok := r.Peek(n)
			if got != want || ok != wantok {
				t.Errorf("Peek(%d): got (%v, %v), want (%v, %v)", n, got, ok, want, wantok)
			}
		}

		checkPeek(0, "kingdom", true)
		checkPeek(-2, "genus", true)
		checkPeek(3, "order", true)
		checkPeek(6, "species", true)
		checkPeek(7, "", false)
		checkPeek(-10, "", false)
	})
}
