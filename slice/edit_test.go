package slice_test

import (
	"math/rand"
	"slices"
	"strconv"
	"strings"
	"testing"

	"github.com/creachadair/mds/slice"
)

func TestLCS(t *testing.T) {
	tests := []struct {
		a, b string
		want string
	}{
		{"", "", ""},

		{"a", "", ""},
		{"", "b", ""},

		{"a b c", "", ""},
		{"", "d e f", ""},

		{"a", "a b c", "a"},
		{"b", "a b c", "b"},
		{"c", "a b c", "c"},
		{"d", "a b c", ""},

		{"a b c", "a b c", "a b c"},
		{"a b c", "a b", "a b"},
		{"b c", "a b c", "b c"},

		{"you will be lucky to get this to work at all",
			"will we be so lucky as to get this to work in the end",
			"will be lucky to get this to work"},

		{"a foolish consistency is the hobgoblin of little minds",
			"four foolish fat hens ate the hobgoblin who is little and minds not",
			"foolish the hobgoblin little minds"},
	}
	for _, tc := range tests {
		as, bs := strings.Fields(tc.a), strings.Fields(tc.b)
		want := strings.Fields(tc.want)
		got := slice.LCS(as, bs)
		if !slices.Equal(got, want) {
			t.Errorf("LCS(%s, %s):\ngot:  %v\nwant: %v", tc.a, tc.b, got, want)
		}
	}
}

func TestLCSRandom(t *testing.T) {
	// Append n randomly generated letters from alpha to *ss.
	pad := func(ss *[]string, n int, alpha string) {
		for i := 0; i < n; i++ {
			j := rand.Intn(len(alpha))
			*ss = append(*ss, alpha[j:j+1])
		}
	}

	// Append 0-4 randomly generated letters from alpha before and after each
	// word in want, and return the resulting sequence.
	input := func(want []string, alpha string) []string {
		var out []string
		for _, w := range want {
			pad(&out, rand.Intn(4), alpha)
			out = append(out, w)
		}
		pad(&out, rand.Intn(4), alpha)
		return out
	}

	// Generate a longest common subsequence of length i, and inputs constructed
	// to have that as their LCS, and verify that they do.
	for i := 0; i < 200; i += 20 {
		var want []string
		pad(&want, i, "abcdefghijklmonpqrstuvwxyz")

		// N.B. The alphabets used by the probe string must not overlap with the
		// inputs, nor the inputs with each other.
		//
		// Probe string: lower-case
		// LHS: digits
		// RHS: upper-case

		lhs := input(want, "0123456789")
		rhs := input(want, "ABCDEFGHIJKLMNOPQRSTUVWXYZ")
		got := slice.LCS(lhs, rhs)
		if !slices.Equal(got, want) {
			t.Errorf("LCS(%q, %q):\ngot:  %q\nwant: %q", lhs, rhs, got, want)
		}
	}
}

func TestEditScript(t *testing.T) {
	tests := []struct {
		a, b string
		want []slice.Edit
	}{
		{"", "", nil},

		{"a", "", pedit(t, "-1")},
		{"", "b", pedit(t, "+1:0")},

		{"a b c", "", pedit(t, "-3")},
		{"", "d e f", pedit(t, "+3:0")},

		{"a", "a b c", pedit(t, "=1 +2:1")},
		{"b", "a b c", pedit(t, "+1:0 =1 +1:2")},
		{"c", "a b c", pedit(t, "+2:0")},
		{"d", "a b c", pedit(t, "x1:0 +2:1")},

		{"a b c", "a b c", pedit(t, "")},
		{"a b c", "a x c", pedit(t, "=1 x1:1")},
		{"a b c", "a b", pedit(t, "=2 -1")},
		{"b c", "a b c", pedit(t, "+1:0")},

		{"a x b x c", "1 x b x 2", pedit(t, "x1:0 =3 x1:4")},
		{"fly you fools", "to fly you must not be fools", pedit(t, "+1:0 =2 +3:3")},

		{"have the best time it is possible to have under the circumstances",
			"I hope you have the time of your life in the forest",
			pedit(t, "+3:0 =2 -1 =1 x4:6 -2 =1 x1:11"),
		},
	}
	for _, tc := range tests {
		as, bs := strings.Fields(tc.a), strings.Fields(tc.b)
		got := slice.EditScript(as, bs)
		if !slices.Equal(got, tc.want) {
			t.Errorf("EditScript(%q, %q):\ngot:  %v\nwant: %v", tc.a, tc.b, got, tc.want)
		}
		checkApply(t, as, bs, got)
	}
}

// checkApply verifies that applying the specified edit script to lhs produces rhs.
func checkApply[T comparable, Slice ~[]T](t *testing.T, lhs, rhs Slice, edit []slice.Edit) {
	t.Helper()

	var out Slice
	i := 0
	for _, e := range edit {
		switch e.Op {
		case slice.OpDelete:
			i += e.N
		case slice.OpInsert:
			out = append(out, rhs[e.X:e.X+e.N]...)
		case slice.OpCopy:
			out = append(out, lhs[i:i+e.N]...)
			i += e.N
		case slice.OpReplace:
			out = append(out, rhs[e.X:e.X+e.N]...)
			i += e.N
		default:
			t.Fatalf("Unexpected edit operation: %v", e)
		}
	}
	out = append(out, lhs[i:]...)
	if !slices.Equal(out, rhs) {
		t.Errorf("Apply %v:\ngot:  %v\nwant: %v", edit, out, rhs)
	} else {
		t.Logf("Apply L %v E %v OK: %v", lhs, edit, out)
	}
}

// pedit parses a string of space-separated edit strings matching the string
// format rendered by the String method of a slice.Edit.
func pedit(t *testing.T, ss string) (out []slice.Edit) {
	t.Helper()
	for _, s := range strings.Fields(ss) {
		var next slice.Edit
		switch s[0] {
		case '-', '=', '+', 'x':
			next.Op = slice.EditOp(s[0])
		default:
			t.Fatalf("Invalid edit op: %c", s[0])
		}
		var err error
		fst, snd, ok := strings.Cut(s[1:], ":")
		next.N, err = strconv.Atoi(fst)
		if err != nil {
			t.Fatalf("Invalid N: %v", err)
		}
		if ok {
			next.X, err = strconv.Atoi(snd)
			if err != nil {
				t.Fatalf("Invalid X: %v", err)
			}
		}
		out = append(out, next)
	}
	return
}
