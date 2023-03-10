// Package slice implements some useful functions for slices.
package slice

// Partition rearranges the elements of vs in-place so that all the elements v
// for which keep(v) is true precede all those for which it is false.  It
// returns the prefix of vs that contains the kept elements.  It takes time
// proportional to len(vs) and does not allocate storage outside the slice.
//
// The input order of the kept elements s preserved, but the unkept elements
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
			return vs[:i]
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
	return vs[:i]
}

// Dedup rearranges the elements of vs in-place to deduplicate consecutive runs
// of identical elements.  It returns a prefix of vs that contains the first
// element of each run found, in their original relative order.  It takes time
// proportional to len(vs) and does not allocate storage outside the slice.
//
// The returned slice will contain (non-consecutive) duplicates if and only if
// vs is not in sorted order at input. If vs is sorted at input (in either
// direction), the elements of the prefix are exactly the unique first elements
// of the input.
func Dedup[T comparable](vs []T) []T {
	if len(vs) <= 1 {
		return vs
	}

	// Setup:
	//   i  : the location of the first element of the next run
	//   j  : runs forward from i looking for the end of the run
	//
	i, j := 0, 1
	for {
		// Scan forward from i for an element different from vs[i].
		for j < len(vs) && vs[i] == vs[j] {
			j++
		}

		// If there are no further distinct elements, we're done.  The item at
		// position i is the beginning of the last run in the slice.
		if j == len(vs) {
			return vs[:i+1]
		}

		// Reaching here, the slice looks like this:
		//
		//   [a b c d d d d d d d e ? ? ?]
		//    0     i             j       n
		//
		// where a, b, c, d are distinct from their neighbors.

		// Otherwise, we have found the first item of a new run at j.
		// Move i forward to the next slot and (if necessary) swap vs[j] into it.
		i++
		if j > i {
			// A swap is unnecessary (though harmless) if j is already the next slot.
			vs[i], vs[j] = vs[j], vs[i]
		}
		j++

		// Now:
		//               swapped
		//            v-----------v
		//   [a b c d e d d d d d d ? ? ?]
		//    0       i             j     n
	}
}

// Remove removes the element at position i of vs, slides the following
// elements down to fill the space, and returns the shortened slice.
//
// If i == len(vs), the input slice is returned without change.
// Remove will panic if i < 0 or i > len(vs).
func Remove[T any](vs []T, i int) []T {
	n := len(vs)
	if i == n {
		return vs
	}
	if i < n-1 {
		copy(vs[i:], vs[i+1:])
	}
	return vs[:n-1]
}

// Insert inserts v at position i of vs, moving the following elements up and
// extending the slice by one element. It returns the extended slice.
//
// If i == len(vs), this is equivalent to append(vs, v).
// Insert will panic if i < 0 or i > len(vs).
func Insert[T any](vs []T, i int, v T) []T {
	if i < 0 || i > len(vs) {
		panic("index out of range")
	}
	out := append(vs, v)
	for j := len(vs); j > i; j-- {
		out[j], out[j-1] = out[j-1], out[j]
	}
	return out
}

// Reverse reverses the contents of vs in-place.
func Reverse[T any](vs []T) {
	for i, j := 0, len(vs)-1; i < j; i++ {
		vs[i], vs[j] = vs[j], vs[i]
		j--
	}
}

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
