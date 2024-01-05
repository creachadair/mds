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
