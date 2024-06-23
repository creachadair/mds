package slice

import (
	"cmp"
	"slices"
)

// Editorial note: "longest increasing subsequence" and "longest
// non-decreasing subsequence" are a mouthful, so in this file
// "subsequence" and "longest subsequence" imply the ordering, unless
// explicitly specified otherwise.

// LNDS computes a longest non-decreasing subsequence of vs.
//
// This implementation takes O(P·log(n)) time and O(n) space for
// inputs of length n = len(vs) and longest subsequence length P. If
// the longest subsequence is the entire input, it takes O(n) time and
// O(n) space.
func LNDS[T cmp.Ordered, Slice ~[]T](vs Slice) Slice {
	return LNDSFunc(vs, cmp.Compare)
}

// LNDSFunc computes a longest non-decreasing subsequence of vs, in
// the order determined by the cmp function. cmp must return a
// negative number when a < b, a positive number when a > b, and zero
// when a == b.
//
// This implementation takes O(P·log(n)) time and O(n) space for
// inputs of length n = len(vs) and longest subsequence length P. If
// the longest subsequence is the entire input, it takes O(n) time and
// O(n) space.
func LNDSFunc[T any, Slice ~[]T](vs Slice, cmp func(a, b T) int) Slice {
	if len(vs) == 0 {
		return vs
	}

	// At its core, the algorithm considers every possible
	// non-decreasing subsequence of vs, and picks one of the
	// longest. The naive implementation is quadratic in either time
	// or space (see lis_test.go for the former). Thankfully, four
	// optimizations let us achieve O(n·log(n)):
	//
	//   - Each element only gets to participate in creating the
	//     longest subsequences it can, we discard all shorter
	//     options. It might still participate in many subsequences of
	//     that length, but that brings us to...
	//   - We only need to remember one subsequence of every length,
	//     the one whose final element is the smallest. This is the
	//     tails array, and means that every new element will
	//     contribute to exactly 0 or 1 subsequence.
	//   - Elements always appear in the tails array in non-decreasing
	//     order, so we can use a binary search to find the one
	//     subsequence that a new element might contribute to. This
	//     gets us O(log(n)) time per element, instead of O(n).
	//   - Successive non-decreasing new elements always contribute to
	//     the longest currently known subsequence, which is the final
	//     entry of tails. If we check for this trivial case before
	//     embarking on the binary search, elements that appear in
	//     non-decreasing order can be processed in O(1) time rather
	//     than O(log(n)).
	//
	// There are truly marvelous proofs of these optimizations, which
	// this comment is too short to contain.

	var (
		// tails[L] is the index into vs for the final element of a
		// subsequence of length L. If several such subsequences
		// exist, tails keeps whichever has the smallest final
		// element, according to cmp.
		tails = make([]int, 1, len(vs))

		// prev[i] is the index into vs for the element that comes
		// before vs[i], in some subsequence tracked by tails. If
		// vs[i] is the first element of that subsequence, prev[i] is
		// -1.
		//
		// It's effectively the pointers of linked lists whose heads
		// are tracked in tails.
		prev = make([]int, len(vs))
	)

	// The loop is cleaner if it can assume that at least 1 element
	// has already been processed. Run the first iteration directly.
	prev[0] = -1
	tails[0] = 0

	for i := range vs[1:] {
		// While the loop is cleaner if we lift out the first
		// iteration, it's confusing to humans to have i be off-by-one
		// from vs's natural indexing. Correct that here so the rest
		// of the loop is easier to follow.
		i++

		idxOfBestTail := tails[len(tails)-1]
		if cmp(vs[i], vs[idxOfBestTail]) >= 0 {
			// Fast path: the i-th element extends the currently known
			// longest subsequence.
			prev[i] = idxOfBestTail
			tails = append(tails, i)
			continue
		}

		// Otherwise, the i-th element _must_ be an improvement over a
		// shorter subsequence currently being tracked in tails. Find
		// and replace it.
		//
		// Note we run the search over tails minus its final element,
		// which avoids repeating the fast path's compare during the
		// search. It doesn't change the outcome since the fast path
		// eliminated the "beyond the end of tails" edge case.
		replaceIdx := bisectRight(tails[:len(tails)-1], vs[i], func(idx int, target T) int {
			return cmp(vs[idx], target)
		})

		// The new element is extending the subsequence tracked in
		// replaceIdx-1. If we're replacing the singleton subsequence,
		// we have to avoid the out of bounds read of tails.
		if replaceIdx == 0 {
			prev[i] = -1
		} else {
			prev[i] = tails[replaceIdx-1]
		}
		tails[replaceIdx] = i
	}

	// We can now iterate back through the longest subsequence and
	// partition the input.
	ret := make([]T, len(tails))
	seqIdx := tails[len(tails)-1] // current longest subsequence element
	for i := range ret {
		ret[len(ret)-1-i] = vs[seqIdx]
		seqIdx = prev[seqIdx]
	}

	return ret
}

// LIS computes a longest strictly increasing subsequence of vs.
//
// This implementation takes O(P·log(n)) time and O(n) space for
// inputs of length n = len(vs) and longest subsequence length P. If
// the longest subsequence is the entire input, it takes O(n) time and
// O(n) space.
func LIS[T cmp.Ordered, Slice ~[]T](vs Slice) Slice {
	return LISFunc(vs, cmp.Compare)
}

// LISFunc computes a longest strictly increasing subsequence of vs,
// in the order determined by the cmp function. cmp must return a
// negative number when a < b, a positive number when a > b, and zero
// when a == b.
//
// This implementation takes O(P·log(n)) time and O(n) space for
// inputs of length n = len(vs) and longest subsequence length P. If
// the longest subsequence is the entire input, it takes O(n) time and
// O(n) space.
func LISFunc[T any, Slice ~[]T](vs Slice, cmp func(a, b T) int) Slice {
	// LISFunc is almost exactly the same as LNDSFunc, except that
	// comparisons are > instead of >= and the binary search leans
	// left. Further comments have been omitted for brevity.
	if len(vs) == 0 {
		return vs
	}

	var (
		tails = make([]int, 1, len(vs))
		prev  = make([]int, len(vs))
	)
	prev[0] = -1
	tails[0] = 0

	for i := range vs[1:] {
		i++

		idxOfBestTail := tails[len(tails)-1]
		if cmp(vs[i], vs[idxOfBestTail]) > 0 {
			prev[i] = idxOfBestTail
			tails = append(tails, i)
			continue
		}

		replaceIdx, _ := slices.BinarySearchFunc(tails[:len(tails)-1], vs[i], func(idx int, target T) int {
			return cmp(vs[idx], target)
		})

		if replaceIdx == 0 {
			prev[i] = -1
		} else {
			prev[i] = tails[replaceIdx-1]
		}
		tails[replaceIdx] = i
	}

	ret := make([]T, len(tails))
	seqIdx := tails[len(tails)-1]
	for i := range ret {
		ret[len(ret)-1-i] = vs[seqIdx]
		seqIdx = prev[seqIdx]
	}

	return ret
}

// bisectRight returns the position where target should be inserted in
// a sorted slice. If target is already present in the slice, the
// returned position is one past the final existing occurrence.
//
// This is effectively a right-leaning variant of
// slices.BinarySearch. It doesn't return a found bool, since by
// definition it will never return an index equivalent to target.
func bisectRight[T, U any, Slice ~[]T](vs Slice, target U, cmp func(T, U) int) (idx int) {
	ln := len(vs)
	low, high := uint(0), uint(ln)
	for low < high {
		mid := (low + high) / 2
		if cmp(vs[mid], target) > 0 {
			high = mid
		} else {
			low = mid + 1
		}
	}
	ret := int(low)
	return ret
}
