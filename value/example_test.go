package value_test

import (
	"fmt"

	"github.com/creachadair/mds/value"
)

var randomValues = []int{1, 6, 16, 19, 4}

func ExampleMaybe() {
	even := make([]value.Maybe[int], 5)
	for i, r := range randomValues {
		if r%2 == 0 {
			even[i] = value.Just(r)
		}
	}

	var count int
	for _, v := range even {
		if v.Present() {
			count++
		}
	}

	fmt.Println(count)
	fmt.Println(even)
	// Output:
	// 3
	// [Absent[int] 6 16 Absent[int] 4]
}
