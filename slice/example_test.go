package slice_test

import (
	"fmt"

	"github.com/creachadair/mds/slice"
)

func ExamplePartition() {
	vs := []int{3, 1, 8, 4, 2, 6, 9, 10, 5, 7}
	odd := slice.Partition(vs, func(v int) bool {
		return v%2 == 1
	})
	fmt.Println(odd)
	// Output:
	// [3 1 9 5 7]
}

func ExampleDedup() {
	vs := []int{1, 1, 3, 2, 2, 4, 4, 5, 1, 3, 3, 3}
	fmt.Println(slice.Dedup(vs))
	// Output:
	// [1 3 2 4 5 1 3]
}

func ExampleRemove() {
	vs := []string{"three", "fuzzy", "kittens"}
	fmt.Println(slice.Remove(vs, 1))
	// Output:
	// [three kittens]
}

func ExampleInsert() {
	vs := []string{"three", "kittens"}
	fmt.Println(slice.Insert(vs, "fuzzy", 1))
	// Output:
	// [three fuzzy kittens]
}

func ExampleInsert_remove() {
	vs := []string{"three", "cute", "kittens"}
	ws := slice.Remove(vs, 1)
	fmt.Println(ws)
	xs := slice.Insert(ws, "fuzzy", 1)
	fmt.Println(xs)
	// Output:
	// [three kittens]
	// [three fuzzy kittens]
}
