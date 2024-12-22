// Package mstr defines utility functions for strings.
package mstr

import (
	"cmp"
	"strings"
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
			continue
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
