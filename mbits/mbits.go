// Package mbits provides functions for manipulating bits and bytes.
package mbits

import (
	"unsafe"
)

// Zero sets the contents of data to zero and returns len(data).
func Zero(data []byte) int {
	n := len(data)
	m := n &^ 7 // end of 64-bit chunks spanned by data

	i := 0
	for ; i < m; i += 8 {
		v := (*uint64)(unsafe.Pointer(&data[i]))
		*v = 0
	}
	for ; i < n; i++ {
		data[i] = 0
	}
	return n
}

// LeadingZeroes reports the number of leading zero bytes at the beginning of data.
func LeadingZeroes(data []byte) int {
	n := len(data)
	m := n &^ 7 // end of full 64-bit chunks spanned by data

	var i int
	for ; i < m; i += 8 {
		v := *(*uint64)(unsafe.Pointer(&data[i]))
		if v != 0 {
			// Count zeroes at the front of v.
			for data[i] == 0 {
				i++
			}
			return i
		}
	}

	// Count however many zeroes are left.
	for i < n && data[i] == 0 {
		i++
	}
	return i
}

// TrailingZeroes reports the number of trailing zero bytes at the end of data.
func TrailingZeroes(data []byte) int {
	n := len(data)

	// Find the start of the tail of data comprising only full-width chunks.
	//
	//   | < 8 | ... 8 ... | ... 8 ... | . . . | ... 8 ... |
	//         ^m                                          ^n
	//
	// Walk backward through these looking for a non-zero.
	m := n - n&^7

	i, nz := n-8, 0
	for ; i >= m; i -= 8 {
		v := *(*uint64)(unsafe.Pointer(&data[i]))
		if v != 0 {
			// Count zeroes at the end of v.
			for data[i+7] == 0 {
				i--
				nz++
			}
			return nz
		}
		nz += 8
	}

	// Count zeroes left at the tail of the ragged block at the front of data.
	for m--; m >= 0 && data[m] == 0; m-- {
		nz++
	}
	return nz
}
