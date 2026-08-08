// Copyright (C) Michael J. Fromberger. All Rights Reserved.

// Package mstr defines utility functions for strings.
package mstr

import (
	"cmp"
	"slices"
	"strings"

	"github.com/creachadair/mds/value"
)

// Trunc returns a prefix of s having length no greater than n bytes.  If s
// exceeds this length, it is truncated at a point ≤ n so that the result does
// not end in a partial UTF-8 encoding. Trunc does not verify that s is valid
// UTF-8, but if it is the result will remain valid after truncation.
func Trunc[String ~string | ~[]byte](s String, n int) String {
	if n >= len(s) {
		return s
	}

	// Back up until we find the beginning of a UTF-8 encoding.
	for n > 0 && s[n-1]&0xc0 == 0x80 { // 0b10... is a continuation byte
		n--
	}

	// If we're at the beginning of a multi-byte encoding, back up one more to
	// skip it. It's possible the value was already complete, but it's simpler
	// if we only have to check in one direction.
	//
	// Otherwise, we have a single-byte code (0b00... or 0b01...).
	if n > 0 && s[n-1]&0xc0 == 0xc0 { // 0b11... starts a multibyte encoding
		n--
	}
	return s[:n]
}

// Lines splits its argument on newlines. It is a convenience function for
// [strings.Split], except that it returns empty if s == "" and treats a
// trailing newline as the end of the file rather than an empty line.
func Lines(s string) []string {
	if s == "" {
		return nil
	}
	return strings.Split(strings.TrimSuffix(s, "\n"), "\n")
}

// Split splits its argument on sep. It is a convenience function for
// [strings.Split], except that it returns empty if s == "".
func Split(s, sep string) []string {
	if s == "" {
		return nil
	}
	return strings.Split(s, sep)
}

// CompareNatural compares its arguments lexicographically, but treats runs of
// decimal digits as the spellings of natural numbers and compares their values
// instead of the individual digits.
//
// For example, "a2b" is after "a12b" under ordinary lexicographic comparison,
// but before under CompareNatural, because 2 < 12.  However, if one argument
// has digits and the other has non-digits at that position (see for example
// "a" vs. "12") the comparison falls back to lexicographic.
//
// CompareNatural returns -1 if a < b, 0 if a == b, and +1 if a > b.
// It does not allocate memory.
//
// Note that CompareNatural is non-strict, as certain lexically distinct
// strings may compare equal; for example "a01b" and "a1b". If a strict order
// is necessary, use [CompareNaturalStrict].
func CompareNatural(a, b string) int {
	for a != "" && b != "" {
		va, ra, aok := parseInt(a)
		vb, rb, bok := parseInt(b)

		if aok && bok {
			// Both begin with runs of digits, compare them numerically.
			if c := cmp.Compare(va, vb); c != 0 {
				return c
			}
			a, b = ra, rb

			// Reaching here, neither suffix can begin with digits (or we would
			// have consumed them above), so fall through to the non-digit case.
		} else if aok != bok {
			// One begins with digits, the other does not.
			// They cannot be equal, so compare them lexicographically.
			return cmp.Compare(a, b)
		}

		// Neither begins with digits. Compare runs of non-digits.
		pa, ra := parseStr(a)
		pb, rb := parseStr(b)
		if c := cmp.Compare(pa, pb); c != 0 {
			return c
		}
		a, b = ra, rb
	}
	return cmp.Compare(a, b)
}

// CompareNaturalStrict behaves as [CompareNatural], but lexically distinct
// strings that are equal under the natural comparison (for example, "a01b" and
// "a1b") are ordered non-decreasing by length.
func CompareNaturalStrict(a, b string) int {
	c := CompareNatural(a, b)
	if c == 0 {
		return cmp.Compare(len(a), len(b)) // shorter first
	}
	return c
}

// parseInt reports whether s begins with a run of one or more decimal digits,
// and if so returns the value of that run, along with the unconsumed tail of
// the string.
func parseInt(s string) (int, string, bool) {
	var i, v int
	for i < len(s) && isDigit(s[i]) {
		v = (v * 10) + int(s[i]-'0')
		i++
	}
	return v, s[i:], i > 0
}

// parseStr returns the longest prefix of s not containing decimal digits,
// along with the remaining suffix of s.
func parseStr(s string) (pfx, sfx string) {
	var i int
	for i < len(s) && !isDigit(s[i]) {
		i++
	}
	return s[:i], s[i:]
}

func isDigit(b byte) bool { return b >= '0' && b <= '9' }

// Match reports whether s matches the specified pattern.  An occurrence of "*"
// in the pattern matches zero or more arbitrary bytes in s; otherwise the
// corresponding positions of s and the pattern must be equal.
//
// Match takes time proportional to the lengths of its arguments, and does not
// allocate memory.
func Match(s, pattern string) bool {
	head, rest, ok := strings.Cut(pattern, "*")
	if !ok {
		// No wildcards, the entire pattern must match exactly.
		return s == pattern
	}
	tail, ok := strings.CutPrefix(s, head)
	if !ok {
		return false
	}
	return matchSuffix(tail, rest)
}

// matchSuffix reports whether a suffix of s matches pattern.
// As with Match, the "*" wildcard matches zero or more bytes.
func matchSuffix(s, pattern string) bool {
	for pattern != "" {
		phead, prest, ok := strings.Cut(pattern, "*")
		if !ok {
			// No further globs, the entire remaining pattern must be a suffix.
			return strings.HasSuffix(s, phead)
		}
		_, stail, ok := strings.Cut(s, phead)
		if !ok {
			// The static prefix of the pattern is not found, match is impossible.
			return false
		}

		// Reaching here:
		// pattern == <phead> *|<prest>
		// s       == <phead>  |<stail>
		//
		// so we can recur on the pieces after the |.
		s, pattern = stail, prest
	}
	return true
}

// Equal returns a function that reports whether its argument is equal to s.
// See also [EqualFold].
func Equal(s string) func(string) bool { return value.Equal(s) }

// EqualFold returns a function that reports whether its argument is equal to s
// up to case folding. See also [Equal].
func EqualFold(s string) func(string) bool {
	return func(t string) bool { return strings.EqualFold(t, s) }
}

// Similarity reports a similarity score between A and B.
// The result is in the closed interval [0..1], where 0 means A and B
// have nothing in common, and 1 means A == B.
//
// This is intended for use with strings representing "words" or "terms" rather
// than long documents. While it will work for strings of arbitrary length, it
// is not an efficient way to compute document similarity.
func Similarity(A, B string) float64 { return jaroWinkler(A, B, true) }

// jaroWinkler computes a Jaro or Jaro-Winkler similarity score for A and B.
// The result is in the closed interval [0..1], where 1 indicates A == B.
// If winkle == false, it returns the plain Jaro score.
// Otherwise it returns the weighted Jaro-Winkler value.
//
// In contrast with the paper, which uses P = 4 and ρ = 0.1 for prefix
// weighting, this implementation computes the longest matching prefix up to
// the length of the shorter input.
//
// See: https://files.eric.ed.gov/fulltext/ED325505.pdf
func jaroWinkler(A, B string, winkle bool) float64 {
	if A == B {
		return 1
	} else if A == "" || B == "" {
		return 0
	}
	// Reaching here, both A and B are non-empty.

	if len(B) > len(A) {
		A, B = B, A // ensure A is the longer, if they differ
	}

	buf := make([]bool, len(A)+len(B))
	// ma[i] is whether A[i] has a δ match
	// mb[j] is whether B[j] has a δ match
	ma, mb := buf[:len(A)], buf[len(A):]

	// Count δ-matches.
	//
	// A "match" is a pair (i, j) with i-δ ≤ j < i+δ+1 and A[i] == B[j],
	// where for any 0 ≤ j´ < j having B[j´] == B[j] there exists an 0 ≤ i´ < i
	// with A[i´] == B[j´]. Informally, this means a match is a leftmost pair of
	// equal bytes within δ positions of each other in their respective strings,
	// that are not claimed by a previous (further left) match.

	δ := len(A) / 2
	var m float64
	for i := range A {
		for j := max(0, i-δ); j < min(i+δ+1, len(B)); j++ {
			if !mb[j] && A[i] == B[j] {
				// Reaching here, A[i] has a δ-match at B[j].
				ma[i] = true
				mb[j] = true
				m++
				break
			}
		}
	}
	if m == 0 {
		return 0 // no matches
	}

	// Count transpositions, i.e., bounded matches that are out of order.
	// We know at this point there is at least one match.
	//
	// Scan matching positions of A from left to right. For each such i,
	// consider the next unclaimed matching position j in B.  If A[i] ≠ B[j], it
	// means that A and B disagree on the order of that byte, i.e., there is a
	// transposition. Count how many times this occurs.
	//
	// For example, given A = "acqb" and B = "abc" with δ = 2, there are three
	// matches
	//
	//    (0, 0):"a"  (1, 2):"c"  (3, 1):"b"
	//
	// At i = 0 in A, the next matching position in B is j = 0.
	// Since A[0] == B[0] == "a", there is no transposition.
	//
	// At i = 1 in A, the next matching position in B is j = 1.
	// Since A[1] == "c" and B[1] == "b", there is a transposition.
	//
	// At i = 2 in A, there is no match.
	//
	// At i = 3 in A, the next matching position in B is j = 2.
	// Since A[3] == "b" and B[2] == "c", there is a transposition.
	//
	// So here we have 3 matches (a, b, c) and 1 transposition (bc / cb).
	// But note that we double-counted the transposition, since we checked
	// both "c-b" (in A) and "bc" (in B). So we will divide the number of
	// observed disparities by 2 to get the "real" number.
	//
	// Although we have two nested loops here, each position is considered only
	// once.
	var t float64
	j := 0
	for i := range A {
		if !ma[i] {
			continue // no match at this position in A
		}

		// Find the next unclaimed match in B.
		for j < len(B) && !mb[j] {
			j++
		}
		if j >= len(B) {
			break // no further matches in B
		}

		if A[i] != B[j] {
			t++
		}
		j++ // this position is consumed
	}
	t /= 2

	// Jaro:
	//         1    /  m     m    m - t  \
	//   sJ = --- · | --- + --- + ------ |
	//         3    \ |A|   |B|     m    /
	//
	sim := (m/float64(len(A)) + m/float64(len(B)) + (m-t)/m) / 3
	if !winkle {
		return sim
	}

	// Winkler:
	//
	//   sW = sJ + λ·ρ·(1 - sJ)
	//
	// Where λ is the length of a common prefix of A and B and ρ is a
	// normalization factor.
	//
	// This increases the similarity score for inputs with a common prefix.
	// Typically one picks constants P > 0 and ρ ≤ 1/P and considers prefixes of
	// up to length P. Here, however, we'll just use the whole of B, the shorter
	// string.
	var lp float64
	for i := range B {
		if A[i] != B[i] {
			break
		}
		lp++
	}
	win := lp / float64(len(B)+1) * (1 - sim)
	return sim + win
}

// Next returns the next string in lexicographic order after s.
func Next(s string) string {
	next := []byte(s)
	for i, b := range slices.Backward(next) {
		if b < 255 {
			next[i]++
			return string(next)
		}
		next[i] = 0 // carry
	}
	next = append(next, 0)
	return string(next)
}

// Next returns the previous string in lexicographic order before s.
// If s == "", it returns "".
func Prev(s string) string {
	if s == "" {
		return ""
	}
	prev := []byte(s)
	for i, b := range slices.Backward(prev) {
		if b > 0 {
			prev[i]--
			return string(prev)
		}
		prev[i] = 255 // borrow
	}

	// All digits are 255; reduce length by 1 (either end is fine).
	return string(prev[1:])
}

// WithPrefix returns a copy of s beginning with prefix. If s already begins
// with prefix, it is returned unchanged.
func WithPrefix(s, prefix string) string {
	if !strings.HasPrefix(s, prefix) {
		return prefix + s
	}
	return s
}

// WithSuffix returns a copy of s ending with suffix. If s already ends with
// suffix, it is returned unchanged
func WithSuffix(s, suffix string) string {
	if !strings.HasSuffix(s, suffix) {
		return s + suffix
	}
	return s
}
