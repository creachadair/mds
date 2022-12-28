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

	// Initial conditions.
	check()
	rcheck(mlink.NewRing[int](0))
	rcheck(mlink.RingOf[int]())
	rcheck(mlink.NewRing[int](4), 0, 0, 0, 0)

	r = mlink.RingOf(0)
	check(0)

	// Popping a single-element chain should be a noop.
	r.Pop()
	check(0)

	r.Join(mlink.RingOf(1, 2, 3))
	check(0, 1, 2, 3)

	// Joining r to itself should do nothing.
	r.Join(r)
	check(0, 1, 2, 3)

	// Test removing items from the ring.
	r.Next().Pop()
	check(0, 2, 3)
	r.Prev().Pop()
	check(0, 2)

	// Test adding items to various places in the ring.
	r.Next().Join(mlink.RingOf(4, 5, 6))
	check(0, 2, 4, 5, 6)
	r.Prev().Join(mlink.RingOf(7))
	check(0, 2, 4, 5, 6, 7)

	// Check circular equivalence.
	r = r.Next().Next()
	check(4, 5, 6, 7, 0, 2)
	r = r.Prev()
	check(2, 4, 5, 6, 7, 0)
	r.Prev().Pop()
	check(2, 4, 5, 6, 7)

	s := mlink.RingOf(10, 20, 30)
	rcheck(s, 10, 20, 30)

	// Joining two separate rings splices in.
	{
		x := r.Next().Join(s)
		check(2, 4, 10, 20, 30, 5, 6, 7)
		//    ^- r  ^- s        ^- x
		rcheck(s, 10, 20, 30, 5, 6, 7, 2, 4)
		//        ^- s        ^- x     ^- r
		rcheck(x, 5, 6, 7, 2, 4, 10, 20, 30)
		//        ^- x     ^- r  ^- s
	}

	// Joining members of the same ring splices out.
	rcheck(r.Join(s), 4) // just the excised part
	check(2, 10, 20, 30, 5, 6, 7)

	rcheck(r.Join(r.Next())) // nothing was removed
	check(2, 10, 20, 30, 5, 6, 7)
}
