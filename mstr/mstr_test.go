package mstr_test

import (
	"testing"

	"github.com/creachadair/mds/mstr"
	gocmp "github.com/google/go-cmp/cmp"
)

func TestString(t *testing.T) {
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
