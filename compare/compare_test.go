package compare_test

import (
	"cmp"
	"math/rand"
	"slices"
	"testing"
	"time"

	"github.com/creachadair/mds/compare"
)

func TestConversion(t *testing.T) {
	for _, less := range [](func(a, b int) bool){
		func(a, b int) bool { return a < b },
		func(a, b int) bool { return a > b },
	} {
		cmp := compare.FromLessFunc(less)
		cless := compare.ToLessFunc(cmp)

		for i := 0; i < 1000; i++ {
			m := rand.Intn(1000) - 500
			n := rand.Intn(1000) - 500

			mn, nm := less(m, n), less(n, m)
			if mn && nm {
				t.Fatalf("Invalid less function: %d < %d and %d < %d", m, n, n, m)
			}
			diff := cmp(m, n)
			switch {
			case mn:
				if diff >= 0 {
					t.Errorf("Compare %d %d: got %v, want ≥ 0", m, n, diff)
				}
				if !cless(m, n) {
					t.Errorf("Less %d %d: got false, want true", m, n)
				}
			case nm:
				if diff <= 0 {
					t.Errorf("Compare %d %d: got %v, want ≤ 0", m, n, diff)
				}
				if cless(m, n) {
					t.Errorf("Less %d %d: got true, want false", m, n)
				}
			default:
				if diff != 0 {
					t.Errorf("Compare %d %d: got %v, want 0", m, n, diff)
				}
				if cless(m, n) {
					t.Errorf("Less %d %d: got true, want false", m, n)
				}
				if cless(n, m) {
					t.Errorf("Less %d %d: got true, want false", n, m)
				}
			}
		}
	}
}

func TestTime(t *testing.T) {
	ptime := func(s string) time.Time {
		ts, err := time.Parse(time.RFC3339Nano, s)
		if err != nil {
			t.Fatalf("Parse time %q: %v", s, err)
		}
		return ts
	}
	tests := []struct {
		a, b string // RFC3339
		want int
	}{
		{"1989-11-09T17:53:00Z", "1989-11-09T17:53:00Z", 0},
		{"2009-11-10T19:00:00-04:00", "2009-11-10T23:00:00Z", 0},
		{"1983-11-20T18:30:45-08:00", "1983-11-21T06:00:00+01:00", -1},
		{"2022-01-31T12:00:00Z", "2021-01-31T12:00:00Z", 1},
	}
	for _, tc := range tests {
		got := compare.Time(ptime(tc.a), ptime(tc.b))
		if got != tc.want {
			t.Errorf("Compare %s ? %s: got %d, want %d", tc.a, tc.b, got, tc.want)
		}
		rev := compare.Time(ptime(tc.b), ptime(tc.a))
		switch {
		case got < 0 && rev <= 0,
			got > 0 && rev >= 0,
			got == 0 && rev != 0:
			t.Errorf("Compare %s ? %s: strict weak order violation: %d / %d", tc.b, tc.a, got, rev)
		}
	}
}

func TestReversed(t *testing.T) {
	buf := make([]int, 37)
	for i := range buf {
		buf[i] = i
	}

	cz := cmp.Compare[int]
	rev := compare.Reversed(cz)

	slices.SortFunc(buf, rev)
	for i := 0; i+1 < len(buf); i++ {
		if buf[i] <= buf[i+1] {
			t.Errorf("Output disordered at %d: %d <= %d", i, buf[i], buf[i+1])
		}
	}

	slices.SortFunc(buf, compare.Reversed(rev))
	if !slices.IsSorted(buf) {
		t.Errorf("Reversed output is not sorted: %v", buf)
	}
}
