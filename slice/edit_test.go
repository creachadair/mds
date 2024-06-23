package slice_test

import (
	"math/rand/v2"
	"regexp"
	"slices"
	"strings"
	"testing"

	"github.com/creachadair/mds/slice"

	_ "embed"
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
		{"a b c d e", "e b c d a", "b c d"},
		{"x y z", "p d q a b", ""},
		{"b a r a t a", "a b a t e", "a a t"},

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
			j := rand.IntN(len(alpha))
			*ss = append(*ss, alpha[j:j+1])
		}
	}

	// Append 0-4 randomly generated letters from alpha before and after each
	// word in want, and return the resulting sequence.
	input := func(want []string, alpha string) []string {
		var out []string
		for _, w := range want {
			pad(&out, rand.IntN(4), alpha)
			out = append(out, w)
		}
		pad(&out, rand.IntN(4), alpha)
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
		want []slice.Edit[string]
	}{
		{"", "", nil},

		{"a", "", pedit(t, "-[a]")},
		{"", "b", pedit(t, "+[b]")},

		{"a b c", "", pedit(t, "-[a b c]")},
		{"", "d e f", pedit(t, "+[d e f]")},

		{"a", "a b c", pedit(t, "=[a] +[b c]")},
		{"b", "a b c", pedit(t, "+[a] =[b] +[c]")},
		{"c", "a b c", pedit(t, "+[a b] =[c]")},
		{"d", "a b c", pedit(t, "![d:a b c]")},

		{"c d", "a b c d", pedit(t, "+[a b] =[c d]")},
		{"a b c", "a b c", nil},
		{"a b c", "a x c", pedit(t, "=[a] ![b:x] =[c]")},
		{"a b c", "a b", pedit(t, "=[a b] -[c]")},
		{"b c", "a b c", pedit(t, "+[a] =[b c]")},
		{"a b c d e", "e b c d a", pedit(t, "![a:e] =[b c d] ![e:a]")},
		{"1 2 3 4", "4 3 2 1", pedit(t, "+[4 3 2] =[1] -[2 3 4]")},
		{"a b c 4", "1 2 4", pedit(t, "![a b c:1 2] =[4]")},
		{"a b 3 4", "0 1 2 3 4", pedit(t, "![a b:0 1 2] =[3 4]")},
		{"1 2 3 4", "1 2 3 5 6", pedit(t, "=[1 2 3] ![4:5 6]")},
		{"1 2 3 4", "1 2 q", pedit(t, "=[1 2] ![3 4:q]")},

		{"a x b x c", "1 x b x 2", pedit(t, "![a:1] =[x b x] ![c:2]")},
		{"fly you fools", "to fly you must not be fools",
			pedit(t, "+[to] =[fly you] +[must not be] =[fools]")},
		{"have the best time it is possible to have under the circumstances",
			"I hope you have the time of your life in the forest",
			pedit(t, "+[I hope you] =[have the] -[best] =[time] "+
				"![it is possible to have under:of your life in] "+
				"=[the] ![circumstances:forest]")},
	}
	for _, tc := range tests {
		as, bs := strings.Fields(tc.a), strings.Fields(tc.b)
		got := slice.EditScript(as, bs)
		if !equalEdits(got, tc.want) {
			t.Errorf("EditScript(%q, %q):\ngot:  %v\nwant: %v", tc.a, tc.b, got, tc.want)
		}
		checkApply(t, as, bs, got)
	}
}

func equalEdits[T comparable](a, b []slice.Edit[T]) bool {
	if len(a) != len(b) {
		return false
	}
	for i := 0; i < len(a); i++ {
		if a[i].Op != b[i].Op ||
			!slices.Equal(a[i].X, b[i].X) ||
			!slices.Equal(a[i].Y, b[i].Y) {
			return false
		}
	}
	return true
}

// checkApply verifies that applying the specified edit script to lhs produces rhs.
func checkApply[T comparable, Slice ~[]T](t *testing.T, lhs, rhs Slice, edit []slice.Edit[T]) {
	t.Helper()

	var out Slice
	for _, e := range edit {
		switch e.Op {
		case slice.OpDrop:
			// nothing to do
		case slice.OpCopy, slice.OpReplace:
			out = append(out, e.Y...)
		case slice.OpEmit:
			out = append(out, e.X...)
		default:
			t.Fatalf("Unexpected edit operation: %v", e)
		}
	}
	if len(edit) == 0 {
		out = rhs
	}
	if !slices.Equal(out, rhs) {
		t.Errorf("Apply %v:\ngot:  %v\nwant: %v", edit, out, rhs)
	} else {
		t.Logf("Apply L %v E %v OK: %v", lhs, edit, out)
	}
}

var argsRE = regexp.MustCompile(`([-+=!])\[([^\]]*)\](?:\s|$)`)

// pedit parses a string of space-separated edit strings matching the string
// format rendered by the String method of a slice.Edit.
func pedit(t *testing.T, ss string) (out []slice.Edit[string]) {
	t.Helper()
	ms := argsRE.FindAllStringSubmatch(ss, -1)
	if ms == nil {
		t.Fatalf("Invalid argument %q", ss)
	}
	for _, m := range ms {
		fs := strings.Fields(m[2])
		var next slice.Edit[string]
		switch m[1] {
		case "+":
			next.Op = slice.OpCopy
			next.Y = fs
		case "-":
			next.Op = slice.OpDrop
			next.X = fs
		case "=":
			next.Op = slice.OpEmit
			next.X = fs
		case "!":
			next.Op = slice.OpReplace
			pre, post, ok := strings.Cut(m[2], ":")
			if !ok {
				t.Fatalf("Missing separator in argument %q", m[2])
			}
			next.X = strings.Fields(pre)
			next.Y = strings.Fields(post)
		default:
			t.Fatalf("Invalid edit op %q", m[1])
		}
		out = append(out, next)
	}
	return
}

//go:embed testdata/bad-lhs.txt
var badLHS string

//go:embed testdata/bad-rhs.txt
var badRHS string

func TestRegression(t *testing.T) {
	// The original implementation appended path elements to a slice, which
	// could in some circumstances lead to paths clobbering each other.  Test
	// that this does not regress.
	t.Run("ShiftLCS", func(t *testing.T) {
		lhs := strings.Split(strings.TrimSpace(badLHS), "\n")
		rhs := strings.Split(strings.TrimSpace(badRHS), "\n")

		// If we overwrite the path improperly, this will panic.  The output was
		// generated from a production value, but the outputs were hashed since
		// only the order matters.
		_ = slice.EditScript(lhs, rhs)
	})
}
