// Copyright (C) Michael J. Fromberger. All Rights Reserved.

package mctx_test

import (
	"context"
	"testing"

	"github.com/creachadair/mds/mctx"
	"github.com/google/go-cmp/cmp"
)

func TestNewUnique(t *testing.T) {
	// All the keys in this slice must compare unequal to each other,
	// even if they were constructed with the same argument.
	var zero mctx.Key[bool]
	keys := []mctx.Key[bool]{
		zero,
		mctx.New[bool]("apple"),
		mctx.New[bool]("apple"),
		mctx.New[bool]("pear"),
	}

	for i, a := range keys {
		for j, b := range keys {
			if j == i {
				continue
			}
			if a == b {
				t.Errorf("Key %d (%v) == key %d (%v)", i, a, j, b)
			}
		}
	}
}

func TestKeyRoundTrip(t *testing.T) {
	type V struct {
		A string
		B int
	}

	var k1 mctx.Key[V]
	var k2 = mctx.New[V]("apple")
	var k3 = mctx.New[V]("pear")

	tests := []struct {
		name  string
		key   mctx.Key[V]
		value V
	}{
		{"empty", k1, V{A: "hello", B: 1}},
		{"nonempty-1", k2, V{A: "apple", B: 2}},
		{"nonempty-2", k3, V{A: "pear", B: 3}},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			ctx := tc.key.Attach(t.Context(), tc.value)
			got, ok := tc.key.Lookup(ctx).GetOK()
			if !ok {
				t.Errorf("%s: not found", tc.key)
			}
			if diff := cmp.Diff(got, tc.value); diff != "" {
				t.Errorf("%s value (-got, +want):\n%s", tc.key, diff)
			}

			// Verify that an unrelated context is not afflicted.
			if got := tc.key.Lookup(t.Context()); got.Present() {
				t.Errorf("%s on base context: got %v; want absent", tc.key, got)
			}
		})
	}
}

func TestKeyNesting(t *testing.T) {
	type V struct{ S string }

	var vkey mctx.Key[V]
	var wkey = mctx.New[V]("alt")

	base := t.Context()
	c1 := vkey.Attach(base, V{S: "apple"})
	c2 := vkey.Attach(c1, V{S: "pear"})
	c3 := wkey.Attach(c2, V{S: "plum"})
	c4 := wkey.Attach(c2, V{S: "cherry"})
	c5 := vkey.Attach(c4, V{S: "quince"})

	tests := []struct {
		input   context.Context
		key     mctx.Key[V]
		want    V
		present bool
	}{
		{base, vkey, V{}, false},

		{c1, vkey, V{S: "apple"}, true},
		{c1, wkey, V{}, false},

		{c2, vkey, V{S: "pear"}, true},
		{c2, wkey, V{}, false},

		{c3, vkey, V{S: "pear"}, true}, // from parent
		{c3, wkey, V{S: "plum"}, true}, // directly attached

		{c4, vkey, V{S: "pear"}, true},   // from parent
		{c4, wkey, V{S: "cherry"}, true}, // directly attached

		{c5, vkey, V{S: "quince"}, true}, // directly attached
		{c5, wkey, V{S: "cherry"}, true}, // from parent
	}

	for _, tc := range tests {
		got := tc.key.Lookup(tc.input)
		if ok := got.Present(); ok != tc.present {
			t.Errorf("%s present: got %v, want %v", tc.key, ok, tc.present)
		}
		if diff := cmp.Diff(got.Get(), tc.want); diff != "" {
			t.Errorf("%s value (-got, +want):\n%s", tc.key, diff)
		}
	}
}
