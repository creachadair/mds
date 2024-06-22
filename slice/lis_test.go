package slice_test

import (
	"cmp"
	"math/rand"
	"slices"
	"testing"

	"github.com/creachadair/mds/slice"
	diff "github.com/google/go-cmp/cmp"
)

func TestLNDSAndLIS(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		in   []int
		want []int
	}{
		{
			name: "nil",
		},
		{
			name: "empty",
			in:   []int{},
			want: []int{},
		},
		{
			name: "singleton",
			in:   []int{1},
			want: []int{1},
		},
		{
			name: "sorted",
			in:   []int{1, 2, 3, 4},
			want: []int{1, 2, 3, 4},
		},
		{
			name: "backwards",
			in:   []int{4, 3, 2, 1},
			want: []int{1},
		},
		{
			name: "organ_pipe",
			in:   []int{1, 2, 3, 4, 3, 2, 1},
			want: []int{1, 2, 3, 3},
		},
		{
			name: "sawtooth",
			in:   []int{0, 1, 0, -1, 0, 1, 0, -1},
			want: []int{0, 0, 0, 0},
		},
		{
			name: "A005132", // from oeis.org
			in:   []int{0, 1, 3, 6, 2, 7, 13, 20, 12, 21, 11, 22, 10},
			want: []int{0, 1, 3, 6, 7, 13, 20, 21, 22},
		},
		{
			name: "swapped_pairs",
			in:   []int{2, 1, 4, 3, 6, 5, 8, 7},
			want: []int{1, 3, 5, 7},
		},
		{
			name: "run_of_equals",
			// swapped_pairs with more 3s sprinkled in.
			in:   []int{2, 1, 3, 4, 3, 6, 3, 5, 8, 3, 7},
			want: []int{1, 3, 3, 3, 3, 7},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got := slice.LNDS(tc.in)
			if diff := diff.Diff(got, tc.want); diff != "" {
				t.Logf("Input was: %v", tc.in)
				t.Logf("Got: %v", got)
				t.Logf("Want: %v", tc.want)
				t.Errorf("LNDS subsequence is wrong (-got+want):\n%s", diff)
			}
		})
	}
}

func TestLNDSAgainstLCS(t *testing.T) {
	t.Parallel()

	// A result from literature relates LNDS and LCS:
	//
	//   len(LNDS(lst)) == len(LCS(lst, Sorted(lst)))
	//
	// Check that this holds true. Ideally we could also compare the
	// actual resultant lists, but there's no guarantee that LNDS and
	// LCS will return the _same_ longest increasing subsequence, if
	// multiple options are available.

	const numVals = 50
	const numIters = 100
	for i := 0; i < numIters; i++ {
		input := randomInts(numVals)

		gotLNDS := slice.LNDSFunc(input, cmp.Compare)

		sorted := append([]int(nil), input...)
		slices.Sort(sorted)
		gotLCS := slice.LCS(input, sorted)

		if got, want := len(gotLNDS), len(gotLCS); got != want {
			t.Logf("Input: %v", input)
			t.Errorf("len(LNDS(x)) = %v, want len(LCS(x, sorted(x))) = %v", got, want)
		}
	}
}

func TestLNDSRandom(t *testing.T) {
	t.Parallel()

	const numVals = 50
	const numIters = 100

	for i := 0; i < numIters; i++ {
		input := randomInts(numVals)
		want := quadraticLNDS(input)
		got := slice.LNDSFunc(input, cmp.Compare)

		if diff := diff.Diff(got, want); diff != "" {
			t.Logf("Input: %v", input)
			t.Logf("Got: %v", got)
			t.Logf("Want: %v", want)
			t.Errorf("LNDS subsequence is wrong (-got+want):\n%s", diff)
		}
	}
}

// quadraticLNDS returns the same longest increasing subsequence of
// lst that slice.LNDSFunc returns, but using a quadratic recursive
// search that is much slower, but more obviously correct by
// inspection.
func quadraticLNDS(lst []int) []int {
	// better reports whether a is a better LNDS than b.
	//
	// a wins if it is longer, or if any of its elements is smaller
	// than its counterpart in b.
	better := func(a, b []int) bool {
		if len(a) > len(b) {
			return true
		} else if len(a) < len(b) {
			return false
		}

		for i := range a {
			if a[i] < b[i] {
				return true
			} else if a[i] > b[i] {
				return false
			}
		}
		// a and b are completely equal, which can happen in the
		// quadratic algorithm since we might generate permutations of
		// indistinguishable equal elements. Returning false avoids a
		// pointless copy.
		return false
	}

	// findLNDS recursively constructs all possible increasing
	// sequences of vs, updating best as it discovers better LNDS
	// candidates.
	var findLNDS func([]int, []int, []int) []int
	findLNDS = func(vs, acc, best []int) (bestOfTree []int) {
		if len(vs) == 0 {
			if better(acc, best) {
				best = append(best[:0], acc...)
			}
			return best
		}

		lnBest := len(best)
		if lnBest > 0 && len(vs)+len(acc) < lnBest {
			// can't possibly do better than what's already known,
			// give up early.
			return best
		}

		elt, vs := vs[0], vs[1:]
		if len(acc) == 0 || elt >= acc[len(acc)-1] {
			// elt could extend acc, try that
			best = findLNDS(vs, append(acc, elt), best)
		}
		// and always try skipping elt
		return findLNDS(vs, acc, best)
	}

	// Preallocate, so the recursion doesn't add insult to injury by
	// allocating as well.
	acc := make([]int, 0, len(lst))
	best := make([]int, 0, len(lst))

	return findLNDS(lst, acc, best)
}

func randomInts(N int) []int {
	ret := make([]int, N)
	for i := range ret {
		ret[i] = rand.Intn(2 * N)
	}
	return ret
}
