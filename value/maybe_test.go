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
		w := value.Just("plum")
		absent := value.Absent[string]()
		tests := []struct {
			lhs, rhs value.Maybe[string]
			want     value.Maybe[string]
		}{
			{absent, absent, absent},
			{v, absent, v},
			{absent, v, v},
			{v, w, v},
			{w, v, w},
		}
		for _, tc := range tests {
			if got := tc.lhs.Or(tc.rhs); got != tc.want {
				t.Errorf("%v.Or(%v): got %v, want %v", tc.lhs, tc.rhs, got, tc.want)
			}
		}
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

func TestJust(t *testing.T) {
	odd := func(v int) (int, bool) {
		return v, v%2 == 1
	}
	t.Run("OK", func(t *testing.T) {
		got := value.JustOK(odd(1))
		if want := value.Just(1); got != want {
			t.Errorf("JustOK(1): got %v, want %v", got, want)
		}
	})
	t.Run("NotOK", func(t *testing.T) {
		got := value.JustOK(odd(2))
		if got.Present() {
			t.Errorf("JustOK(2): got %v, want absent", got)
		}
	})
	t.Run("NoErr", func(t *testing.T) {
		got := value.JustErr(strconv.Atoi("1"))
		if want := value.Just(1); got != want {
			t.Errorf("JustErr(1): got %v, want %v", got, want)
		}
	})
	t.Run("Err", func(t *testing.T) {
		got := value.JustErr(strconv.Atoi("bogus"))
		if got.Present() {
			t.Errorf("JustErr(bogus): got %v, want absent", got)
		}
	})
}

func TestMapMaybe(t *testing.T) {
	length := value.MapMaybe(func(s string) int {
		return len(s)
	})

	tests := []struct {
		input value.Maybe[string]
		want  value.Maybe[int]
	}{
		{value.Just(""), value.Just(0)},
		{value.Just("plum"), value.Just(4)},
		{value.Absent[string](), value.Absent[int]()},
	}
	for _, tc := range tests {
		if got := length(tc.input); got != tc.want {
			t.Errorf("Length %q: got %v, want %v", tc.input, got, tc.want)
		}
	}
}

func TestFirst(t *testing.T) {
	type tv = value.Maybe[int]
	var absent tv
	v1 := value.Just(1)
	v2 := value.Just(2)
	tests := []struct {
		input []tv
		want  tv
	}{
		{nil, absent},
		{[]tv{}, absent},
		{[]tv{absent}, absent},
		{[]tv{absent, absent, absent}, absent},

		{[]tv{v1}, v1},
		{[]tv{absent, v1}, v1},
		{[]tv{absent, v1, v2}, v1},
		{[]tv{absent, absent, v1, absent, v2}, v1},
	}
	for _, tc := range tests {
		if got := value.First(tc.input...); got != tc.want {
			t.Errorf("First %+v: got %v, want %v", tc.input, got, tc.want)
		}
	}
}
