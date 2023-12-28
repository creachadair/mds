package value_test

import (
	"math/rand"
	"testing"

	"github.com/creachadair/mds/value"
)

func TestLessCompare(t *testing.T) {
	for _, less := range [](func(a, b int) bool){
		func(a, b int) bool { return a < b },
		func(a, b int) bool { return a > b },
	} {
		cmp := value.LessCompare(less)

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
			case nm:
				if diff <= 0 {
					t.Errorf("Compare %d %d: got %v, want ≤ 0", m, n, diff)
				}
			default:
				if diff != 0 {
					t.Errorf("Compare %d %d: got %v, want 0", m, n, diff)
				}
			}
		}
	}
}

func TestPtr(t *testing.T) {
	p1 := value.Ptr("foo")
	p2 := value.Ptr("foo")
	if p1 == p2 {
		t.Errorf("Values should have distinct pointers (%p == %p)", p1, p1)
	}
	if *p1 != "foo" || *p2 != "foo" {
		t.Errorf("Got p1=%q, p2=%q; wanted both foo", *p1, *p2)
	}
}
