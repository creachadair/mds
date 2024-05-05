package oset_test

import (
	"testing"

	"github.com/creachadair/mds/mtest"
	"github.com/creachadair/mds/oset"
	gocmp "github.com/google/go-cmp/cmp"
)

func TestSet(t *testing.T) {
	s := oset.New[string]("two", "three", "five", "seven")
	checkHas := func(key string, want bool) {
		t.Helper()
		if got := s.Has(key); got != want {
			t.Errorf("Has %q: got %v, want %v", key, got, want)
		}
	}
	checkLen := func(want int) {
		t.Helper()
		if n := s.Len(); n != want {
			t.Errorf("Len: got %d, want %d", n, want)
		}
	}

	checkLen(4)
	s.Clear()
	checkLen(0)

	s.Add("apple", "pear", "plum", "cherry")
	checkLen(4)

	checkHas("apple", true)
	checkHas("pear", true)
	checkHas("plum", true)
	checkHas("cherry", true)
	checkHas("dog", false)

	s.Add("plum")
	checkLen(4)

	// Note we want the string to properly reflect the set ordering.
	if got, want := s.String(), `oset[apple cherry pear plum]`; got != want {
		t.Errorf("String:\n got: %q\nwant: %q", got, want)
	}

	var got []string
	for it := s.First(); it.IsValid(); it.Next() {
		got = append(got, it.Value())
	}
	if diff := gocmp.Diff(got, []string{"apple", "cherry", "pear", "plum"}); diff != "" {
		t.Errorf("Iter (-got, +want):\n%s", diff)
	}
	if diff := gocmp.Diff(s.Slice(), []string{"apple", "cherry", "pear", "plum"}); diff != "" {
		t.Errorf("Slice (-got, +want):\n%s", diff)
	}

	got = got[:0]
	for it := s.Seek("dog"); it.IsValid(); it.Next() {
		got = append(got, it.Value())
	}
	if diff := gocmp.Diff(got, []string{"pear", "plum"}); diff != "" {
		t.Errorf("Seek dog (-got, +want):\n%s", diff)
	}

	s.Remove("dog")
	checkLen(4)

	s.Remove("pear")
	checkHas("pear", false)
	checkLen(3)

	s.Clear()
	checkLen(0)
}

func TestZero(t *testing.T) {
	var zero oset.Set[string]

	if zero.Len() != 0 {
		t.Errorf("Len is %d, want 0", zero.Len())
	}
	if zero.Has("whatever") {
		t.Error("Unexpectedly has whatever")
	}
	zero.Remove("whatever")
	if it := zero.First(); it.IsValid() {
		t.Errorf("Iter zero: unexected value %q", it.Value())
	}
	if it := zero.First().Seek("whatever"); it.IsValid() {
		t.Errorf("Seek(whatever): unexected value %q", it.Value())
	}
	zero.Clear() // don't panic

	mtest.MustPanicf(t, func() { zero.Add("bad") },
		"Set on a zero set should panic")
}

func TestIterEdit(t *testing.T) {
	s := oset.New[string]()

	s.Add("a", "b", "c", "d", "e")

	var got []string
	for it := s.First(); it.IsValid(); {
		key := it.Value()
		if key == "b" || key == "d" {
			s.Remove(key)
			it.Seek(key)
		} else {
			got = append(got, key)
			it.Next()
		}
	}
	if diff := gocmp.Diff(got, []string{"a", "c", "e"}); diff != "" {
		t.Errorf("Result (-got, +want):\n%s", diff)
	}
}
