// Package mbits provides functions for manipulating bits and bytes.
package mbits

import (
	"unsafe"
)

// Zero sets the contents of data to zero and returns len(data).
func Zero(data []byte) int { clear(data); return len(data) }

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
	m := n &^ 7 // end of full 64-bit chunks spanned by data

	// Note that we still take 64-bit chunks starting from offset 0, because the
	// allocator is likely to have aligned the slice. We count the ragged tail
	// first (if there is one) and then process strides, as above, but in
	// reverse order.

	var nz int
	for n > m && data[n-1] == 0 {
		nz++
		n--
	}

	for i := n - 8; i >= 0; i -= 8 {
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
	return nz
}
