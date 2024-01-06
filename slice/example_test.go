package slice_test

import (
	"fmt"
	"sort"
	"strings"

	"github.com/creachadair/mds/slice"
)

func isOdd(v int) bool  { return v%2 == 1 }
func isEven(v int) bool { return v%2 == 0 }

func ExamplePartition() {
	vs := []int{3, 1, 8, 4, 2, 6, 9, 10, 5, 7}
	odd := slice.Partition(vs, isOdd)
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

func ExampleReverse() {
	vs := []string{"red", "yellow", "blue", "green"}
	fmt.Println("before:", strings.Join(vs, " "))
	slice.Reverse(vs)
	fmt.Println("after:", strings.Join(vs, " "))
	// Ouput:
	// before: red yellow blue green
	// after: green blue yellow red
}

func ExampleSplit() {
	vs := []int{1, 2, 3, 4, 5, 6, 7, 8}

	fmt.Println("input:", vs)
	lhs, rhs := slice.Split(vs, 3)
	fmt.Println("lhs:", lhs)
	fmt.Println("rhs:", rhs)
	// Output:
	// input: [1 2 3 4 5 6 7 8]
	// lhs: [1 2 3]
	// rhs: [4 5 6 7 8]
}

func ExampleMatchingKeys() {
	vs := map[string]int{"red": 3, "yellow": 6, "blue": 4, "green": 5}

	keys := slice.MatchingKeys(vs, isEven)
	sort.Strings(keys) // sort so test output is stable
	for _, key := range keys {
		fmt.Println(key, vs[key])
	}
	// Output:
	// blue 4
	// yellow 6
}

func ExampleChunks() {
	vs := strings.Fields("my heart is a fish hiding in the water grass")

	for _, c := range slice.Chunks(vs, 3) {
		fmt.Println(c)
	}
	// Output:
	// [my heart is]
	// [a fish hiding]
	// [in the water]
	// [grass]
}

func ExampleBatches() {
	vs := strings.Fields("the freckles in our eyes are mirror images that when we kiss are perfectly aligned")

	for _, b := range slice.Batches(vs, 4) {
		fmt.Println(b)
	}
	// Output:
	// [the freckles in our]
	// [eyes are mirror images]
	// [that when we kiss]
	// [are perfectly aligned]
}
