package slice_test

import (
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
		lnds []int
		lis  []int
	}{
		{
			name: "nil",
		},
		{
			name: "empty",
			in:   []int{},
			lnds: []int{},
			lis:  []int{},
		},
		{
			name: "singleton",
			in:   []int{1},
			lnds: []int{1},
			lis:  []int{1},
		},
		{
			name: "sorted",
			in:   []int{1, 2, 3, 4},
			lnds: []int{1, 2, 3, 4},
			lis:  []int{1, 2, 3, 4},
		},
		{
			name: "backwards",
			in:   []int{4, 3, 2, 1},
			lnds: []int{1},
			lis:  []int{1},
		},
		{
			name: "organ_pipe",
			in:   []int{1, 2, 3, 4, 3, 2, 1},
			lnds: []int{1, 2, 3, 3},
			lis:  []int{1, 2, 3, 4},
		},
		{
			name: "sawtooth",
			in:   []int{0, 1, 0, -1, 0, 1, 0, -1},
			lnds: []int{0, 0, 0, 0},
			lis:  []int{-1, 0, 1},
		},
		{
			name: "A005132", // from oeis.org
			in:   []int{0, 1, 3, 6, 2, 7, 13, 20, 12, 21, 11, 22, 10},
			lnds: []int{0, 1, 3, 6, 7, 13, 20, 21, 22},
			lis:  []int{0, 1, 3, 6, 7, 13, 20, 21, 22},
		},
		{
			name: "swapped_pairs",
			in:   []int{2, 1, 4, 3, 6, 5, 8, 7},
			lnds: []int{1, 3, 5, 7},
			lis:  []int{1, 3, 5, 7},
		},
		{
			name: "run_of_equals",
			// swapped_pairs with more 3s sprinkled in.
			in:   []int{2, 1, 3, 4, 3, 6, 3, 5, 8, 3, 7},
			lnds: []int{1, 3, 3, 3, 3, 7},
			lis:  []int{1, 3, 4, 5, 7},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			lnds := slice.LNDS(tc.in)
			if diff := diff.Diff(lnds, tc.lnds); diff != "" {
				t.Logf("Input was: %v", tc.in)
				t.Logf("Got: %v", lnds)
				t.Logf("Want: %v", tc.lnds)
				t.Errorf("LNDS subsequence is wrong (-got+want):\n%s", diff)
			}

			lis := slice.LIS(tc.in)
			if diff := diff.Diff(lis, tc.lis); diff != "" {
				t.Logf("Input was: %v", tc.in)
				t.Logf("Got: %v", lis)
				t.Logf("Want: %v", tc.lis)
				t.Errorf("LIS subsequence is wrong (-got+want):\n%s", diff)
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

		gotLNDS := slice.LNDS(input)

		sorted := append([]int(nil), input...)
		slices.Sort(sorted)
		gotLCS := slice.LCS(input, sorted)

		if got, want := len(gotLNDS), len(gotLCS); got != want {
			t.Logf("Input: %v", input)
			t.Errorf("len(LNDS(x)) = %v, want len(LCS(x, sorted(x))) = %v", got, want)
		}
	}
}

func TestLISAgainstLCS(t *testing.T) {
	t.Parallel()

	// The same result from the LNDS vs. LCS test applies, but only on
	// lists of distinct elements.

	const numVals = 50
	const numIters = 100
	for i := 0; i < numIters; i++ {
		input := rand.Perm(numVals)

		gotLIS := slice.LIS(input)

		sorted := append([]int(nil), input...)
		slices.Sort(sorted)
		gotLCS := slice.LCS(input, sorted)

		if got, want := len(gotLIS), len(gotLCS); got != want {
			t.Logf("Input: %v", input)
			t.Errorf("len(LIS(x)) = %v, want len(LCS(x, sorted(x))) = %v", got, want)
		}
	}
}

func TestLNDSRandom(t *testing.T) {
	t.Parallel()

	const numVals = 50
	const numIters = 100

	for i := 0; i < numIters; i++ {
		input := randomInts(numVals)
		want := quadraticIncreasingSubsequence(input, false)
		got := slice.LNDS(input)

		if diff := diff.Diff(got, want); diff != "" {
			t.Logf("Input: %v", input)
			t.Logf("Got: %v", got)
			t.Logf("Want: %v", want)
			t.Errorf("LNDS subsequence is wrong (-got+want):\n%s", diff)
		}
	}
}

func TestLISRandom(t *testing.T) {
	t.Parallel()

	const numVals = 50
	const numIters = 100

	for i := 0; i < numIters; i++ {
		input := randomInts(numVals)
		want := quadraticIncreasingSubsequence(input, true)
		got := slice.LIS(input)

		if diff := diff.Diff(got, want); diff != "" {
			t.Logf("Input: %v", input)
			t.Logf("Got: %v", got)
			t.Logf("Want: %v", want)
			t.Errorf("LIS subsequence is wrong (-got+want):\n%s", diff)
		}
	}
}

// quadraticRisingSequence recursively scans all subsequences of lst,
// looking for the longest increasing subsequence. if
// strictlyIncreasing is true it returns the same as slice.LIS,
// otherwise it returns the same as slice.LNDS.
func quadraticIncreasingSubsequence(lst []int, strictlyIncreasing bool) []int {
	// better reports whether a is a better increasing subsequence
	// than b. a is better if it is longer, or if any of its elements
	// is smaller than its counterpart in b.
	better := func(a, b []int) bool {
		if len(a) > len(b) {
			return true
		} else if len(a) < len(b) {
			return false
		}

		// We can't use slices.Compare alone because we need list
		// length to win over list contents, and contents to matter
		// only between equal lists. But we can use it for the equal
		// case.
		return slices.Compare(a, b) < 0
	}

	canExtend := func(prev, next int) bool { return next >= prev }
	if strictlyIncreasing {
		canExtend = func(prev, next int) bool { return next > prev }
	}

	// findIS recursively constructs all possible increasing sequences
	// of vs, updating best as it discovers better candidates for
	// longest.
	var findIS func([]int, []int, []int) []int
	findIS = func(vs, acc, best []int) (bestOfTree []int) {
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
		if len(acc) == 0 || canExtend(acc[len(acc)-1], elt) {
			// elt could extend acc, try that
			best = findIS(vs, append(acc, elt), best)
		}
		// and always try skipping elt
		return findIS(vs, acc, best)
	}

	// Preallocate, so the recursion doesn't add insult to injury by
	// allocating as well.
	acc := make([]int, 0, len(lst))
	best := make([]int, 0, len(lst))

	return findIS(lst, acc, best)
}

func randomInts(N int) []int {
	ret := make([]int, N)
	for i := range ret {
		ret[i] = rand.Intn(2 * N)
	}
	return ret
}
