// Package mbits provides functions for manipulating bits and bytes.
package mbits

import "unsafe"

// Zero sets the contents of data to zero and returns len(data).
func Zero(data []byte) int {
	n := len(data)
	m := n &^ 7 // count of 64-bit chunks spanned by data

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
