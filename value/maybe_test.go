package value_test

import (
	"strconv"
	"testing"

	"github.com/creachadair/mds/value"
)

func TestMaybe(t *testing.T) {
	t.Run("Zero", func(t *testing.T) {
		var v value.Maybe[int]
		if v.Present() {
			t.Error("Zero maybe should not be present")
		}
		if got := v.Get(); got != 0 {
			t.Errorf("Get: got %d, want 0", got)
		}
	})

	t.Run("Present", func(t *testing.T) {
		v := value.Just("apple")
		if got, ok := v.GetOK(); !ok || got != "apple" {
			t.Errorf("GetOK: got %q, %v; want apple, true", got, ok)
		}
		if got := v.Get(); got != "apple" {
			t.Errorf("Get: got %q, want apple", got)
		}
		if !v.Present() {
			t.Error("Value should be present")
		}
	})

	t.Run("Or", func(t *testing.T) {
		v := value.Just("pear")
		absent := value.Absent[string]()
		tests := []struct {
			lhs       value.Maybe[string]
			rhs, want string
		}{
			{absent, "", ""},
			{v, "", "pear"},
			{absent, "plum", "plum"},
			{v, "plum", "pear"},
		}
		for _, tc := range tests {
			if got := tc.lhs.Or(tc.rhs); got != value.Just(tc.want) {
				t.Errorf("%v.Or(%v): got %v, want %v", tc.lhs, tc.rhs, got, tc.want)
			}
		}
	})

	t.Run("Ptr", func(t *testing.T) {
		t.Run("Present", func(t *testing.T) {
			v := value.Just("plum")
			if p := v.Ptr(); p == nil {
				t.Errorf("Ptr(%v): got nil, want non-nil", v)
			} else if *p != "plum" {
				t.Errorf("*Ptr(%v): got %q, want %q", v, *p, "plum")
			}
		})

		t.Run("Absent", func(t *testing.T) {
			v := value.Absent[int]()
			if p := v.Ptr(); p != nil {
				t.Errorf("Ptr(%v): got %p (%d), want nil", v, p, *p)
			}
		})
	})

	t.Run("String", func(t *testing.T) {
		v := value.Just("pear")
		if got := v.String(); got != "pear" {
			t.Errorf("String: got %q, want pear", got)
		}

		var w value.Maybe[string]
		if got, want := w.String(), "Absent[string]"; got != want {
			t.Errorf("String: got %q, want %q", got, want)
		}
	})
}

func TestCheck(t *testing.T) {
	t.Run("OK", func(t *testing.T) {
		got := value.Check(strconv.Atoi("1"))
		if want := value.Just(1); got != want {
			t.Errorf("Check(1): got %v, want %v", got, want)
		}
	})
	t.Run("Error", func(t *testing.T) {
		got := value.Check(strconv.Atoi("bogus"))
		if got.Present() {
			t.Errorf("Check(bogus): got %v, want absent", got)
		}
	})
}

func TestAtMaybe(t *testing.T) {
	tests := []struct {
		input *string
		want  value.Maybe[string]
	}{
		{nil, value.Absent[string]()},
		{value.Ptr("foo"), value.Just("foo")},
		{value.Ptr(""), value.Just("")},
	}
	for _, tc := range tests {
		if got := value.AtMaybe(tc.input); got != tc.want {
			t.Errorf("MaybeAt(%p): got %q, want %q", tc.input, got, tc.want)
		}
	}
}
