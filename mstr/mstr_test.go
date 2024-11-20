package mstr_test

import (
	"testing"

	"github.com/creachadair/mds/mstr"
	gocmp "github.com/google/go-cmp/cmp"
)

func TestTrunc(t *testing.T) {
	tests := []struct {
		input string
		size  int
		want  string
	}{
		{"", 1000, ""},                 // n > length
		{"abc", 4, "abc"},              // n > length
		{"abc", 3, "abc"},              // n == length
		{"abcdefg", 4, "abcd"},         // n < length, safe
		{"abcdefg", 0, ""},             // n < length, safe
		{"abc\U0001fc2d", 3, "abc"},    // n < length, at boundary
		{"abc\U0001fc2d", 4, "abc"},    // n < length, mid-rune
		{"abc\U0001fc2d", 5, "abc"},    // n < length, mid-rune
		{"abc\U0001fc2d", 6, "abc"},    // n < length, mid-rune
		{"abc\U0001fc2defg", 7, "abc"}, // n < length, cut multibyte
	}

	for _, tc := range tests {
		got := mstr.Trunc(tc.input, tc.size)
		if got != tc.want {
			t.Errorf("Trunc(%q, %d): got %q, want %q", tc.input, tc.size, got, tc.want)
		}
	}
}

func TestLines(t *testing.T) {
	tests := []struct {
		input string
		want  []string
	}{
		{"", nil},
		{" ", []string{" "}},
		{"\n", []string{""}},
		{"\n ", []string{"", " "}},
		{"a\n", []string{"a"}},
		{"\na\n", []string{"", "a"}},
		{"a\nb\n", []string{"a", "b"}},
		{"a\nb", []string{"a", "b"}},
		{"\n\n\n", []string{"", "", ""}},
		{"\n\nq", []string{"", "", "q"}},
		{"\n\nq\n", []string{"", "", "q"}},
		{"a b\nc\n\n", []string{"a b", "c", ""}},
		{"a b\nc\n\nd\n", []string{"a b", "c", "", "d"}},
	}
	for _, tc := range tests {
		if diff := gocmp.Diff(mstr.Lines(tc.input), tc.want); diff != "" {
			t.Errorf("Lines %q (-got, +want):\n%s", tc.input, diff)
		}
	}
}

func TestSplit(t *testing.T) {
	tests := []struct {
		input, sep string
		want       []string
	}{
		{"", "x", nil},
		{"y", "x", []string{"y"}},
		{"x", "x", []string{"", ""}},
		{"ax", "x", []string{"a", ""}},
		{"xa", "x", []string{"", "a"}},
		{"axbxc", "x", []string{"a", "b", "c"}},
		{"axxc", "x", []string{"a", "", "c"}},
		{"a,b,c,,d", ",", []string{"a", "b", "c", "", "d"}},
	}
	for _, tc := range tests {
		if diff := gocmp.Diff(mstr.Split(tc.input, tc.sep), tc.want); diff != "" {
			t.Errorf("Split %q on %q (-got, +want):\n%s", tc.input, tc.sep, diff)
		}
	}
}

func TestCompareNatural(t *testing.T) {
	tests := []struct {
		a, b string
		want int
	}{
		{"", "", 0},

		// Non-empty vs. empty with non-digits and digits.
		{"x", "", 1},
		{"", "x", -1},
		{"0", "", 1},
		{"", "0", -1},

		// Leading zeroes do not change the value.
		{"1", "1", 0},
		{"01", "1", 0},
		{"1", "01", 0},

		// Mixed values.
		{"a1", "a1", 0},
		{"a2", "a1", 1},
		{"a1", "a2", -1},
		{"6c", "06c", 0},
		{"06c", "6c", 0},
		{"5c", "06c", -1},
		{"07c", "6c", 1},

		// Multi-digit numeric runs.
		{"a2b", "a25b", -1},
		{"a12b", "a2", 1},
		{"a25b", "a21b", 1},
		{"a025b", "a25b", 0},

		// Non-matching types compare lexicographically.
		// Note it is not possible for these to be equal.
		{"123", "a", -1},     // because 'a' > '1'
		{"123", ".", 1},      // because '.' < '1'
		{"12c9", "12cv", -1}, // because 'v' > '9'

		// Normal lexicographic comparison, without digits.
		{"a-b-c", "a-b-c", 0},
		{"a-b-c", "a-b-d", -1},
		{"a-b-c-d", "a-b-d", -1},
		{"a-q", "a-b-c", 1},
		{"a-q-c", "a-b-c", 1},

		// Complicated cases ("v" indicates the point of divergence).
		//         v                v
		{"test1-143a19", "test01-143b13", -1},
		//    v                v
		{"test5-143a21", "test04-999", 1},
		//      v               v           'w' > '9'
		{"test5-word-5", "test5-999-5", 1},
	}
	for _, tc := range tests {
		if got := mstr.CompareNatural(tc.a, tc.b); got != tc.want {
			t.Errorf("Compare(%q, %q): got %v, want %v", tc.a, tc.b, got, tc.want)
		}
	}
}
