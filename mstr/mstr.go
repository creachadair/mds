// Package mstr defines utility functions for strings.
package mstr

import "strings"

// Trunc returns a prefix of s having length no greater than n bytes.  If s
// exceeds this length, it is truncated at a point ≤ n so that the result does
// not end in a partial UTF-8 encoding. Trunc does not verify that s is valid
// UTF-8, but if it is the result will remain valid after truncation.
func Trunc(s string, n int) string {
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
// trailing newline as the end of the file.
func Lines(s string) []string {
	if s == "" {
		return nil
	}
	return strings.Split(strings.TrimSuffix(s, "\n"), "\n")
}
