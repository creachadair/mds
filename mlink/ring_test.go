package mlink_test

import (
	"testing"

	"github.com/creachadair/mds/mlink"
	"github.com/google/go-cmp/cmp"
)

func TestRing(t *testing.T) {
	var r *mlink.Ring[int]

	rcheck := func(r *mlink.Ring[int], want ...int) {
		t.Helper()
		var got []int
		r.Each(func(v int) bool {
			got = append(got, v)
			return true
		})
		if diff := cmp.Diff(want, got); diff != "" {
			t.Errorf("Wrong contents (-got, +want):\n%s", diff)
		}
		if got := r.Len(); got != len(want) {
			t.Errorf("Wrong length: got %v, want %v", got, len(want))
		}
	}
	check := func(want ...int) { t.Helper(); rcheck(r, want...) }

	t.Run("Initialize", func(t *testing.T) {
		check()
		rcheck(mlink.NewRing[int](0))
		rcheck(mlink.RingOf[int]())
		rcheck(mlink.NewRing[int](4), 0, 0, 0, 0)

		r = mlink.RingOf(0)
		check(0)
	})

	t.Run("Joining", func(t *testing.T) {
		r.Join(mlink.RingOf(1, 2, 3))
		check(0, 1, 2, 3)

		// Joining r to itself should do nothing.
		r.Join(r)
		check(0, 1, 2, 3)

		// Test adding items to various places in the ring.
		r.Next().Join(mlink.RingOf(4, 5, 6))
		check(0, 1, 4, 5, 6, 2, 3)
		r.Prev().Join(mlink.RingOf(7))
		check(0, 1, 4, 5, 6, 2, 3, 7)
	})

	t.Run("Popping", func(t *testing.T) {
		rcheck(mlink.RingOf(1).Pop(), 1)
		q := mlink.RingOf(2, 3, 5, 7, 11)
		rcheck(q, 2, 3, 5, 7, 11)
		q.Next().Pop()
		rcheck(q, 2, 5, 7, 11)
		q.Prev().Pop()
		rcheck(q, 2, 5, 7)
	})

	t.Run("Circularity", func(t *testing.T) {
		r = r.Next().Next()
		check(4, 5, 6, 2, 3, 7, 0, 1)
		r = r.Prev()
		check(1, 4, 5, 6, 2, 3, 7, 0)
		r.Next().Join(r.At(4))
		check(1, 4, 2, 3, 7, 0)
	})

	s := mlink.RingOf(10, 20, 30)
	rcheck(s, 10, 20, 30)

	t.Run("SplicingIn", func(t *testing.T) {
		x := r.Next().Join(s)
		check(1, 4, 10, 20, 30, 2, 3, 7, 0)
		//    ^- r  ^- s        ^- x
		rcheck(s, 10, 20, 30, 2, 3, 7, 0, 1, 4)
		//        ^- s        ^- x        ^- r
		rcheck(x, 2, 3, 7, 0, 1, 4, 10, 20, 30)
		//        ^- x        ^- r  ^- s
	})

	t.Run("SplicingOut", func(t *testing.T) {
		rcheck(r.Join(s), 4) // just the excised part
		check(1, 10, 20, 30, 2, 3, 7, 0)

		rcheck(r.Join(r.Next())) // nothing was removed
		check(1, 10, 20, 30, 2, 3, 7, 0)
	})

	t.Run("Peek", func(t *testing.T) {
		checkPeek := func(n, want int, wantok bool) {
			t.Helper()
			got, ok := r.Peek(n)
			if got != want || ok != wantok {
				t.Errorf("Peek(%d): got (%v, %v), want (%v, %v)", n, got, ok, want, wantok)
			}
		}

		checkPeek(0, 1, true)
		checkPeek(-2, 7, true)
		checkPeek(3, 30, true)
		checkPeek(6, 7, true)
		checkPeek(7, 0, true)
		checkPeek(8, 0, false)
		checkPeek(-10, 0, false)
	})
}
