// Copyright (C) Michael J. Fromberger. All Rights Reserved.

package mstr_test

import (
	"cmp"
	"math"
	"path"
	"regexp"
	"strings"
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
		a, b                string
		wantNat, wantStrict int
	}{
		{"", "", 0, 0},

		// Non-empty vs. empty with non-digits and digits.
		{"x", "", 1, 1},
		{"", "x", -1, -1},
		{"0", "", 1, 1},
		{"", "0", -1, -1},

		// Leading zeroes do not change the value.
		{"1", "1", 0, 0},
		{"01", "1", 0, 1},
		{"1", "01", 0, -1},

		// Mixed values.
		{"a1", "a1", 0, 0},
		{"a2", "a1", 1, 1},
		{"a1", "a2", -1, -1},
		{"6c", "06c", 0, -1},
		{"06c", "6c", 0, 1},
		{"5c", "06c", -1, -1},
		{"07c", "6c", 1, 1},

		// Multi-digit numeric runs.
		{"a2b", "a25b", -1, -1},
		{"a12b", "a2", 1, 1},
		{"a25b", "a21b", 1, 1},
		{"a025b", "a25b", 0, 1},

		// Non-matching types compare lexicographically.
		// Note it is not possible for these to be equal.
		{"123", "a", -1, -1},     // because 'a' > '1'
		{"123", ".", 1, 1},       // because '.' < '1'
		{"12c9", "12cv", -1, -1}, // because 'v' > '9'

		// Normal lexicographic comparison, without digits.
		{"a-b-c", "a-b-c", 0, 0},
		{"a-b-c", "a-b-d", -1, -1},
		{"a-b-c-d", "a-b-d", -1, -1},
		{"a-q", "a-b-c", 1, 1},
		{"a-q-c", "a-b-c", 1, 1},

		// Complicated cases ("v" indicates the point of divergence).
		//         v                v
		{"test1-143a19", "test01-143b13", -1, -1},
		//    v                v
		{"test5-143a21", "test04-999", 1, 1},
		//      v               v           'w' > '9'
		{"test5-word-5", "test5-999-5", 1, 1},
	}
	for _, tc := range tests {
		if got := mstr.CompareNatural(tc.a, tc.b); got != tc.wantNat {
			t.Errorf("CompareNatural(%q, %q): got %v, want %v", tc.a, tc.b, got, tc.wantNat)
		}
		if got := mstr.CompareNaturalStrict(tc.a, tc.b); got != tc.wantStrict {
			t.Errorf("CompareNaturalStrict(%q, %q): got %v, want %v", tc.a, tc.b, got, tc.wantStrict)
		}
	}

	t.Run("NoAlloc/Natural", func(t *testing.T) {
		const numRuns = 5000

		// We want the test probe to have a difference, but make it go all the way to the end before discovering it.
		// In between there are some numeric spans with leading zeroes that we expect to compare equal.
		const lhs = "a 2 b 034 c 567 d 89-1 e f 23456.78 ghijk 9abc225 10 11 121 lmnopq 999 end"
		const rhs = "a 02 b 34 c 567 d 89-01 e f 23456.78 ghijk 9abc225 010 11 121 lmnopq 999 end EXTRA"

		na := testing.AllocsPerRun(numRuns, func() {
			if c := mstr.CompareNatural(lhs, rhs); c >= 0 {
				t.Fatalf("wrong comparison result: %d", c)
			}
		})
		if na != 0 {
			t.Fatalf("Saw %f allocations, want 0", na)
		}
	})
	t.Run("NoAlloc/Strict", func(t *testing.T) {
		const numRuns = 5000

		// We want the test probe to have a difference, but make it go all the way to the end before discovering it.
		// In between there are some numeric spans with leading zeroes that we expect to compare equal.
		const lhs = "a 2 b 034 c 567 d 89-1 e f 23456.78 ghijk 9abc225 10 11 121 lmnopq 999 end"
		const rhs = "a 02 b 34 c 567 d 89-01 e f 23456.78 ghijk 9abc225 010 11 121 lmnopq 999 end EXTRA"

		na := testing.AllocsPerRun(numRuns, func() {
			if c := mstr.CompareNaturalStrict(lhs, rhs); c >= 0 {
				t.Fatalf("wrong comparison result: %d", c)
			}
		})
		if na != 0 {
			t.Fatalf("Saw %f allocations, want 0", na)
		}
	})
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
		const text = "ohai aaaX_XaaaY_YaaaZ_ZaaaP_PaaaD_DaaaQ_QaaaZ_ZaaaV_VaaaM_MaaaO_OaaaM_MaaaG_GaaaW_WaaaT_TaaaF_Faaa_aaa_aaa kthxbai"
		const pattern = "*a*a*a*a*a*a*a*a*a*a*a*a*a*"

		na := testing.AllocsPerRun(numRuns, func() {
			if !mstr.Match(text, pattern) {
				t.Fatal("no match")
			}
		})
		if na != 0 {
			t.Fatalf("Saw %f allocations, want 0", na)
		}
	})
}

func TestEqual(t *testing.T) {
	tests := []struct {
		s, t    string
		eq, eqf bool
	}{
		{"", "", true, true},
		{"", "x", false, false},
		{"", "X", false, false},
		{"x", "", false, false},
		{"X", "", false, false},
		{"y", "y", true, true},
		{"z", "y", false, false},
		{"ABC", "abc", false, true},
		{"def", "DEF", false, true},
		{"GHI", "ghi", false, true},
		{"JKL", "JKL", true, true},
	}
	for _, tc := range tests {
		equal := mstr.Equal(tc.s)
		equalFold := mstr.EqualFold(tc.s)
		label := cmp.Or(tc.s, "ε") + "_" + cmp.Or(tc.t, "ε")
		t.Run("Equal/"+label, func(t *testing.T) {
			got := equal(tc.t)
			if got != tc.eq {
				t.Errorf("Equal(%q, %q): got %v, want %v", tc.s, tc.t, got, tc.eq)
			}

			// Consistency check: All lexically equal strings are fold-equal too.
			if got && !equalFold(tc.t) {
				t.Errorf("Equal(%q, %q) is true but EqualFold is not", tc.s, tc.t)
			}
		})
		t.Run("EqualFold/"+label, func(t *testing.T) {
			if got := equalFold(tc.t); got != tc.eqf {
				t.Errorf("EqualFold(%q, %q): got %v, want %v", tc.s, tc.t, got, tc.eqf)
			}
		})
	}
}

func TestSimilarity(t *testing.T) {
	const ε = 0.009
	eqf := func(got, want float64) bool {
		return math.Abs(got-want) <= ε
	}
	tests := []struct {
		A, B  string
		want  float64
		exact bool
	}{
		// Exact values for boundary cases.
		{"", "", 1, true},
		{"x", "", 0, true},
		{"", "x", 0, true},
		{"foo", "foo", 1, true},
		{"abc", "def", 0, true}, // no matches

		{"fo", "of", 0.83, false},             // 2 matches, 1 transp
		{"abc", "cba", 0.55, false},           // 1 match (b), 0 transp (a and c are too far apart)
		{"qarcsb", "abc", 0.66, false},        // 2 matches (a, c), 0 transp
		{"qcrasb", "abc", 0.5, false},         // 2 matches (a, c), 1 transp
		{"aqrcsb", "abc", 0.75, false},        // 2 matches (a, c), 0 transp; prefix 1
		{"acqb", "abc", 0.85, false},          // 3 matches (a, b, c), 1 transp; prefix 1
		{"garbage", "cabbage", 0.81, false},   // 5 matches, 0 transp
		{"babbage", "cabbage", 0.85, false},   // 6 matches, 1 transp
		{"carbage", "cabbage", 0.92, false},   // 6 matches, 0 transp; prefix 2
		{"cabbage", "southwest", 0.41, false}, // 1 match (e), 0 transp
		{"alien", "predator", 0.38, false},    // 2 matches (a, e), 1 transp
		{"pike", "puzzlement", 0.65, false},   // 2 matches (p, e), 0 transp; prefix 1
		{"puke", "puzzlement", 0.81, false},   // 3 matches (p, u, e), 0 trans; prefix 2
		{"flaky", "flukes", 0.8, false},       // 3 matches (f, l, k), 0 trans; prefix 2
		{"mispsell", "imsspell", 0.91, false}, // 8 matches, 2 transp
		{"mispsell", "misspell", 0.97, false}, // 8 matches, 1 transp; prefix 3
	}
	for _, tc := range tests {
		got := mstr.Similarity(tc.A, tc.B)
		t.Logf("Similarity(%q, %q) = %.3f", tc.A, tc.B, got)
		if tc.exact && got != tc.want {
			t.Errorf("want %v", tc.want)
		} else if !tc.exact && !eqf(got, tc.want) {
			t.Errorf("want approximately %v", tc.want)
		}

		// Similarity should always commute. We compare exactly here, because we
		// really want the same value, regardless whether it was correct.
		comm := mstr.Similarity(tc.B, tc.A) // N.B. order reversed
		if comm != got {
			t.Errorf("Similarity(%q, %q) = %v, want %v", tc.B, tc.A, comm, got)
		}
	}
}

func BenchmarkMatch(b *testing.B) {
	const text = "ohai aaaX_XaaaY_YaaaZ_ZaaaP_PaaaD_DaaaQ_QaaaZ_ZaaaV_VaaaM_MaaaO_OaaaM_MaaaG_GaaaW_WaaaT_TaaaF_Faaa_aaa_aaa kthxbai"
	const pattern = "*a*a*a*a*a*a*a*a*a*a*a*a*a*"

	b.Run("Match", func(b *testing.B) {
		for b.Loop() {
			_ = mstr.Match(text, pattern)
		}
	})

	b.Run("Regexp", func(b *testing.B) {
		// Don't charge the cost of compiling the expression against the match.
		parts := strings.Split(pattern, "*")
		for i, p := range parts {
			parts[i] = regexp.QuoteMeta(p)
		}
		m := regexp.MustCompile(`^` + strings.Join(parts, ".*") + `$`)

		for b.Loop() {
			_ = m.MatchString(text)
		}
	})

	b.Run("PathMatch", func(b *testing.B) {
		// The pattern grammar for path.Match has more operators than mstr.Match,
		// but the test probes we are using here relies only on the "*".
		// We should expect mstr.Match performance to be comparable with path.Match.
		for b.Loop() {
			_, _ = path.Match(pattern, text)
		}
	})
}

func BenchmarkSimilarity(b *testing.B) {
	const A = "The thousand injuries of Fortunato I had borne as I best could but when he ventured upon insult I vowed revenge"
	const Bsim = "The ten thousand injuries of Importunato I took as I best would but when he chose violence I vowed revenge"
	const Bdiff = "xxxxxx xxx yy yyyyy yyyyy zz zzzz wwwwwwwww www wwww www wwww vvv vvvv yyyyyyy yy qqqq qqqqqqqqq q xxx xx"

	b.Logf("A vs. Bsim: %v", mstr.Similarity(A, Bsim))
	b.Logf("A vs. Bdiff: %v", mstr.Similarity(A, Bdiff))

	b.Run("Similar", func(b *testing.B) {
		for b.Loop() {
			mstr.Similarity(A, Bsim)
		}
	})
	b.Run("Dissimilar", func(b *testing.B) {
		for b.Loop() {
			mstr.Similarity(A, Bdiff)
		}
	})
}
