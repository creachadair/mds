// Package slice implements some useful functions for slices.
package slice

// Partition rearranges the elements of vs in-place so that all the elements v
// for which keep(v) is true precede all those for which it is false.  It
// returns the number i of kept elements. It takes time proportional to len(vs)
// and does not allocate storage outside the slice.
//
// The input order of the kept elements (those at indexes < i) is preserved,
// but the unkept elements (indexes ≥ i) are permuted arbitrarily. For example,
// given the input:
//
//	[6, 1, 3, 2, 8, 4, 5]
//
// and
//
//	func keep(v int) bool { return v%2 == 0 }
//
// the resulting partition returns 4 and the resulting slice looks like
//
//	[6, 2, 8, 4, ...]
//
// where "..." containts the elements 1, 3, and 5 in unspecified order.
func Partition[V any](vs []V, keep func(V) bool) int {
	n := len(vs)

	// Invariant: Everything to the left of i is kept.
	// Initialize left cursor (i) by scanning forward for an unkept element.
	i := 0
	for i < n && keep(vs[i]) {
		i++
	}
	// Initialize right cursor (j). If there is an out-of-place kept element,
	// it must be after i.
	j := i + 1

	for i < n && j < n {
		// Right: Scan forward for a kept element.
		for !keep(vs[j]) {
			j++

			// If the right cursor reached the end, we're done: Everything left
			// of i is kept, everything ≥ i is unkept.
			if j == n {
				return i
			}
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
	return i
}
