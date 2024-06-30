package value_test

import (
	"fmt"
	"math/rand"

	"github.com/creachadair/mds/value"
)

var rng = rand.New(rand.NewSource(104))

func ExampleMaybe() {
	even := make([]value.Maybe[int], 5)
	for i := range even {
		if r := rng.Intn(20); r%2 == 0 {
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
