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
		{"", 0, ""},                               // n == length
		{"", 1000, ""},                            // n > length
		{"abc", 4, "abc"},                         // n > length
		{"abc", 3, "abc"},                         // n == length
		{"abcdefg", 4, "abcd"},                    // n < length, safe
		{"abcdefg", 0, ""},                        // n < length, safe
		{"abc\U0001f60a", 3, "abc"},               // n < length, at boundary
		{"abc\U0001f60a", 4, "abc"},               // n < length, mid-rune
		{"abc\U0001f60a", 5, "abc"},               // n < length, mid-rune
		{"abc\U0001f60a", 6, "abc"},               // n < length, mid-rune
		{"abc\U0001f60axxx", 7, "abc"},            // n < length, cut multibyte
		{"abc\U0001f60axxx", 8, "abc\U0001f60ax"}, // n < length, keep multibyte
	}

	for _, tc := range tests {
		t.Logf("Input %q len=%d n=%d", tc.input, len(tc.input), tc.size)
		got := mstr.Trunc(tc.input, tc.size)
		if got != tc.want {
			t.Errorf("Trunc(%q, %d) [string]: got %q, want %q", tc.input, tc.size, got, tc.want)
		}
		if got := mstr.Trunc([]byte(tc.input), tc.size); string(got) != tc.want {
			t.Errorf("Trunc(%q, %d) [bytes]: got %q, want %q", tc.input, tc.size, got, tc.want)
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

func TestMatch(t *testing.T) {
	tests := []struct {
		s, pattern string
		want       bool
	}{
		{"", "", true},
		{"", "*", true},
		{"", "**", true},
		{"*", "*", true},
		{"*", "**", true},
		{"", "abc", false},
		{"abc", "", false},
		{"abc", "abc", true},
		{"abc", "abc*", true},
		{"abc", "*abc", true},
		{"abc", "a*c", true},
		{"abc", "a*cd", false},
		{"abcd", "a*cd", true},
		{"abXcd", "a*cd", true},
		{"abcdef", "ab**ef", true},
		{"abc_def", "abc**def", true},
		{"____xyz", "*xyz", true},
		{"____xy", "*xyz", false},
		{"abc", "abc*", true},
		{"abc___", "abc*", true},
		{"ab___", "abc*", false},
		{"ab___", "ab*c*", false},
		{"ab__cd_", "ab*c*", true},
	}
	for _, tc := range tests {
		if got := mstr.Match(tc.s, tc.pattern); got != tc.want {
			t.Errorf("Match(%q, %q): got %v, want %v", tc.s, tc.pattern, got, tc.want)
		}
	}

	t.Run("NoAlloc", func(t *testing.T) {
		const numRuns = 5000
		const needle = "ohai aaaX_XaaaY_YaaaZ_ZaaaP_PaaaD_DaaaQ_QaaaZ_ZaaaV_VaaaM_MaaaO_OaaaM_MaaaG_GaaaW_WaaaT_TaaaF_Faaa_aaa_aaa kthxbai"
		const pattern = "*a*a*a*a*a*a*a*a*a*a*a*a*a*"

		na := testing.AllocsPerRun(numRuns, func() {
			if !mstr.Match(needle, pattern) {
				t.Fatal("no match")
			}
		})
		if na != 0 {
			t.Fatalf("Saw %f allocations, want 0", na)
		}
	})
}
