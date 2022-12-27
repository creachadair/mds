package mlink_test

import (
	"testing"

	"github.com/creachadair/mds/mlink"
	"github.com/google/go-cmp/cmp"
)

func TestRing(t *testing.T) {
	r := mlink.NewRing[int](0)

	rcheck := func(r *mlink.Ring[int], want ...int) {
		var got []int
		r.Each(func(v int) bool {
			got = append(got, v)
			return true
		})
		if diff := cmp.Diff(want, got); diff != "" {
			t.Errorf("Wrong contents (-got, +want):\n%s", diff)
		}
	}
	check := func(want ...int) { rcheck(r, want...) }

	check(0)
	r.Pop()
	check(0)
	r.Add(1, 2, 3)
	check(0, 1, 2, 3)
	r.Next().Pop()
	check(0, 2, 3)
	r.Prev().Pop()
	check(0, 2)
	r.Next().Add(4, 5, 6)
	check(0, 2, 4, 5, 6)
	r.Prev().Add(7)
	check(0, 2, 4, 5, 6, 7)
	r = r.Next()
	check(2, 4, 5, 6, 7, 0)
	r.Prev().Pop()
	check(2, 4, 5, 6, 7)

	s := mlink.NewRing(10).Add(20, 30)
	rcheck(s, 10, 20, 30)
	r.Next().Adjoin(s)
	check(2, 4, 10, 20, 30, 5, 6, 7)
	rcheck(s, 10, 20, 30, 5, 6, 7, 2, 4)
}
