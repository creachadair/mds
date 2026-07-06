// Copyright (C) Michael J. Fromberger. All Rights Reserved.

package value_test

import (
	"testing"

	"github.com/creachadair/mds/value"
)

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

func TestAt(t *testing.T) {
	tests := []struct {
		input *string
		want  string
	}{
		{nil, ""},
		{value.Ptr("foo"), "foo"},
	}
	for _, tc := range tests {
		if got := value.At(tc.input); got != tc.want {
			t.Errorf("At(%p): got %q, want %q", tc.input, got, tc.want)
		}
	}
}

func TestAtDefault(t *testing.T) {
	tests := []struct {
		input *string
		dflt  string
		want  string
	}{
		{nil, "", ""},
		{nil, "foo", "foo"},
		{value.Ptr("foo"), "bar", "foo"},
	}
	for _, tc := range tests {
		if got := value.AtDefault(tc.input, tc.dflt); got != tc.want {
			t.Errorf("AtDefault(%p, %q): got %q, want %q", tc.input, tc.dflt, got, tc.want)
		}
	}
}

func TestCond(t *testing.T) {
	type altBool bool
	tests := []struct {
		flag bool
		x, y string
		want string
	}{
		{true, "a", "b", "a"},
		{false, "a", "b", "b"},
		{true, "", "q", ""},
		{false, "", "q", "q"},
		{true, "z", "", "z"},
		{false, "z", "", ""},
	}
	for _, tc := range tests {
		if got := value.Cond(tc.flag, tc.x, tc.y); got != tc.want {
			t.Errorf("Cond bool(%v, %v, %v): got %v, want %v", tc.flag, tc.x, tc.y, got, tc.want)
		}
		if got := value.Cond(altBool(tc.flag), tc.x, tc.y); got != tc.want {
			t.Errorf("Cond boolish(%v, %v, %v): got %v, want %v", tc.flag, tc.x, tc.y, got, tc.want)
		}
	}
}

func TestEqual(t *testing.T) {
	checkEqual(t, 0, 0, true)
	checkEqual(t, 0, 1, false)
	checkEqual(t, byte(15), byte(15), true)
	checkEqual(t, byte(11), byte(0), false)
	checkEqual(t, "", "", true)
	checkEqual(t, "yes", "yes", true)
	checkEqual(t, "yes", "no", false)
	checkEqual(t, '\x00', '\x00', true)
	checkEqual(t, '\x01', '\x00', false)

	type thing struct{ S string }
	checkEqual(t, thing{"foo"}, thing{"foo"}, true)
	checkEqual(t, thing{}, thing{"bar"}, false)
	checkEqual(t, thing{"baz"}, thing{"quux"}, false)
}

func checkEqual[T comparable](t *testing.T, v, w T, want bool) {
	t.Helper()
	equal := value.Equal(v)
	if got := equal(w); got != want {
		t.Errorf("Equal %T %v %v: got %v, want %v", v, v, w, got, want)
	}
}
