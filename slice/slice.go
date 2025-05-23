// Package slice implements some useful functions for slices.
package slice

import (
	"iter"
)

// Partition rearranges the elements of vs in-place so that all the elements v
// for which keep(v) is true precede all those for which it is false.  It
// returns the prefix of vs that contains the kept elements.  It takes time
// proportional to len(vs) and does not allocate storage outside the slice.
//
// The input order of the kept elements is preserved, but the unkept elements
// are permuted arbitrarily. For example, given the input:
//
//	[6, 1, 3, 2, 8, 4, 5]
//
// and
//
//	func keep(v int) bool { return v%2 == 0 }
//
// after partition vs looks like:
//
//	[6, 2, 8, 4, ...]
//
// where "..." contains the elements 1, 3, and 5 in unspecified order, and the
// returned slice is:
//
//	[6, 2, 8, 4]
//
// The capacity of the slice returned is clipped to its length, so that
// appending to it will not modify the elements of vs after those kept.
func Partition[T any](vs []T, keep func(T) bool) []T {
	if len(vs) == 0 {
		return vs
	}

	// Invariant: Everything to the left of i is kept.
	// Initialize left cursor (i) by scanning forward for an unkept element.
	i := 0
	for i < len(vs) && keep(vs[i]) {
		i++
	}
	// Initialize right cursor (j). If there is an out-of-place kept element,
	// it must be after i.
	j := i + 1

	for i < len(vs) {
		// Right: Scan forward for a kept element.
		for j < len(vs) && !keep(vs[j]) {
			j++
		}
		// If the right cursor reached the end, we're done: Everything left of i
		// is kept, everything ≥ i is unkept.
		if j == len(vs) {
			return vs[:i:i]
		}

		// Reaching here, the elements under both cursors are out of
		// order. Swap to put them in order, then advance the cursors.
		// After swapping, we have:
		//
		//    [+ + + + + + - - - - ? ? ? ?]
		//     0         i       j         n
		//
		// where + denotes a kept element, - unkept, and ? unknown.
		// The next unkept element (if any) must therefore be at i+1, and the
		// next candidate to replace it must be > j.

		vs[i], vs[j] = vs[j], vs[i]
		i++
		j++
	}
	return vs[:i:i]
}

// Zero sets all the elements of vs to their zero value.
//
// Deprecated: Use the built-in clear function instead.
func Zero[T any, Slice ~[]T](vs Slice) { clear(vs) }

// MapKeys extracts a slice of the keys from a map.  The resulting slice is in
// arbitrary order.
func MapKeys[T comparable, U any](m map[T]U) []T {
	if len(m) == 0 {
		return nil
	}
	keys := make([]T, 0, len(m))
	for key := range m {
		keys = append(keys, key)
	}
	return keys
}

func sliceCheck(i, n int) (int, bool) {
	if i < 0 {
		i += n
	}
	return i, i >= 0 && i <= n
}

func indexCheck(i, n int) (int, bool) {
	if i < 0 {
		i += n
	}
	return i, i >= 0 && i < n
}

// At returns the element of ss at offset i. Negative offsets count backward
// from the end of the slice. If i is out of range, At will panic.
func At[T any, Slice ~[]T](ss Slice, i int) T {
	b, ok := indexCheck(i, len(ss))
	if !ok {
		panic("index out of range")
	}
	return ss[b]
}

// PtrAt returns a pointer to the element of ss at offset i.  Negative offsets
// count backward from the end of the slice.  If i is out of range, PtrAt
// returns nil.
func PtrAt[T any, Slice ~[]T](ss Slice, i int) *T {
	if pos, ok := indexCheck(i, len(ss)); ok {
		return &ss[pos]
	}
	return nil
}

// MatchingKeys returns an iterator over the keys k of m for which f(m[k]) is
// true.  The results are delivered in arbitrary order.
func MatchingKeys[T comparable, U any](m map[T]U, f func(U) bool) iter.Seq[T] {
	return func(yield func(T) bool) {
		for k, v := range m {
			if f(v) {
				if !yield(k) {
					return
				}
			}
		}
	}
}

// Rotate permutes the elements of ss in-place by k positions.
// If k > 0, elements are rotated rightward.
// If k < 0, elements are rotated leftward.
// If k is out of range, Rotate will panic.
//
// For example, if
//
//	ss := []string{"a", "b", "c", "d"}
//
// then slice.Rotate(ss, 1) produces
//
//	{"d", "a", "b", "c"}
//
// while slice.Rotate(ss, -1) produces
//
//	{"b", "c", "d", "a"}
//
// The rotation operation takes time proportional to len(ss) but does not
// allocate storage outside the input slice.
func Rotate[T any, Slice ~[]T](ss Slice, k int) {
	k, ok := sliceCheck(k, len(ss))
	if !ok {
		panic("offset out of range")
	} else if k == 0 || k == len(ss) {
		return
	}

	// There are (k, n) cycles of the rotation permutation, and we must chase
	// them all to complete the rotation. The residues of the GCD can be used as
	// starting points. Despite the nested loop here, we will visit each element
	// of the slice only once (on its cycle).
	g := gcd(k, len(ss))
	for j := range g {
		i, cur := j, ss[j]
		for {
			next := (i + k) % len(ss)
			nextv := ss[next]
			ss[next] = cur
			if next == j {
				break
			}
			i, cur = next, nextv
		}
	}
}

func gcd(a, b int) int {
	for b != 0 {
		a, b = b, a%b
	}
	return a
}

// Chunks iterates a sequence of contiguous subslices ("chunks") of vs, each
// having length at most n and together covering the input.  All slices except
// the last will have length exactly n; the last may have fewer. The slices
// yielded share storage with the input.
//
// Chunks will panic if n < 0. If n == 0, Chunks yields no chunks.
//
// Deprecated: Use [slices.Chunk] instead.
func Chunks[T any, Slice ~[]T](vs Slice, n int) iter.Seq[Slice] {
	if n < 0 {
		panic("max must be positive")
	} else if n == 0 {
		return func(func(Slice) bool) {}
	}
	return func(yield func(Slice) bool) {
		i := 0
		for i < len(vs) {
			end := min(i+n, len(vs))
			if !yield(vs[i:end:end]) {
				return
			}
			i = end
		}
	}
}

// Batches iterates a sequence of up to n contiguous subslices ("batches") of
// vs, each having nearly as possible to equal length and together covering the
// input. The slices returned share storage with the input. If n > len(vs), the
// number of batches is capped at len(vs); otherwise exactly n are constructed.
//
// Batches will panic if n < 0. If n == 0 Batches yields no batches.
func Batches[T any, Slice ~[]T](vs Slice, n int) iter.Seq[Slice] {
	if n < 0 {
		panic("n out of range")
	} else if n == 0 {
		return func(func(Slice) bool) {}
	} else if n > len(vs) {
		n = len(vs)
	}
	return func(yield func(Slice) bool) {
		i, size, rem := 0, len(vs)/n, len(vs)%n
		for i < len(vs) {
			end := i + size
			if rem > 0 {
				end++
				rem--
			}
			if !yield(vs[i:end:end]) {
				return
			}
			i = end
		}
	}
}

// Stripe returns a "stripe" of the ith elements of each slice in vs.  Any
// slice that does not have an ith element is skipped. If none of the slices
// has an ith element, the result is empty.
func Stripe[T any, Slice ~[]T](vs []Slice, i int) Slice {
	var out Slice
	for _, v := range vs {
		if i < len(v) {
			out = append(out, v[i])
		}
	}
	return out
}

// Head returns a subslice of up to n elements from the head (front) of vs.  If
// vs has fewer than n elements, the whole slice is returned.
func Head[T any, Slice ~[]T](vs Slice, n int) Slice {
	if len(vs) < n {
		return vs
	}
	return vs[:n]
}

// Tail returns a subslice of up to n elements from the tail (end) of vs. If vs
// has fewer than n elements, the whole slice is returned.
func Tail[T any, Slice ~[]T](vs Slice, n int) Slice {
	if len(vs) < n {
		return vs
	}
	return vs[len(vs)-n:]
}

// Select returns an iterator over the elements v of vs for which f(v) is true,
// in the same order they occur in the input.
func Select[T any, Slice ~[]T](vs Slice, f func(T) bool) iter.Seq[T] {
	return func(yield func(T) bool) {
		for _, v := range vs {
			if f(v) && !yield(v) {
				return
			}
		}
	}
}
