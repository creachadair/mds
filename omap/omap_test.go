package omap_test

import (
	"testing"

	"github.com/creachadair/mds/mtest"
	"github.com/creachadair/mds/omap"
	"github.com/google/go-cmp/cmp"
)

func TestMap(t *testing.T) {
	m := omap.New[string, int]()
	checkGet := func(key string, want int) {
		t.Helper()
		v := m.Get(key)
		if v != want {
			t.Errorf("Get %q: got %d, want %d", key, v, want)
		}
	}
	checkLen := func(want int) {
		t.Helper()
		if n := m.Len(); n != want {
			t.Errorf("Len: got %d, want %d", n, want)
		}
	}

	checkLen(0)

	m.Set("apple", 1)
	m.Set("pear", 2)
	m.Set("plum", 3)
	m.Set("cherry", 4)

	checkLen(4)

	checkGet("apple", 1)
	checkGet("pear", 2)
	checkGet("plum", 3)
	checkGet("cherry", 4)
	checkGet("dog", 0) // i.e., not found

	m.Set("plum", 100)
	checkGet("plum", 100)

	// Note we want the string to properly reflect the map ordering.
	if got, want := m.String(), `omap[apple:1 cherry:4 pear:2 plum:100]`; got != want {
		t.Errorf("String:\n got: %q\nwant: %q", got, want)
	}

	var got []string
	for it := m.First(); it.IsValid(); it.Next() {
		got = append(got, it.Key())
	}
	if diff := cmp.Diff(got, []string{"apple", "cherry", "pear", "plum"}); diff != "" {
		t.Errorf("Iter (-got, +want):\n%s", diff)
	}
	if diff := cmp.Diff(m.Keys(), []string{"apple", "cherry", "pear", "plum"}); diff != "" {
		t.Errorf("Keys (-got, +want):\n%s", diff)
	}

	got = got[:0]
	for it := m.Seek("dog"); it.IsValid(); it.Next() {
		got = append(got, it.Key())
	}
	if diff := cmp.Diff(got, []string{"pear", "plum"}); diff != "" {
		t.Errorf("Seek dog (-got, +want):\n%s", diff)
	}

	if m.Delete("dog") {
		t.Error("Delete(dog) incorrectly reported true")
	}
	checkLen(4)

	if !m.Delete("pear") {
		t.Error("Delete(pear) incorrectly reported false")
	}
	checkGet("pear", 0)
	checkLen(3)

	m.Clear()
	checkLen(0)
}

func TestZero(t *testing.T) {
	var zero omap.Map[string, string]

	if zero.Len() != 0 {
		t.Errorf("Len is %d, want 0", zero.Len())
	}
	if v, ok := zero.GetOK("whatever"); ok || v != "" {
		t.Errorf(`Get whatever: got (%q, %v), want ("", false)`, v, ok)
	}
	if zero.Delete("whatever") {
		t.Error("Delete(whatever) incorrectly reported true")
	}
	if it := zero.First(); it.IsValid() {
		t.Errorf("Iter zero: unexected key %q=%q", it.Key(), it.Value())
	}
	if it := zero.First().Seek("whatever"); it.IsValid() {
		t.Errorf("Seek(whatever): unexected key %q=%q", it.Key(), it.Value())
	}
	zero.Clear() // don't panic

	mtest.MustPanicf(t, func() { zero.Set("bad", "mojo") },
		"Set on a zero map should panic")
}
