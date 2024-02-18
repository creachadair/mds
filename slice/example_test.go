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

func ExampleRotate() {
	vs := []int{8, 6, 7, 5, 3, 0, 9}

	slice.Rotate(vs, -3)
	fmt.Println(vs)
	// Output:
	// [5 3 0 9 8 6 7]
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

func ExampleLCS() {
	fmt.Println(slice.LCS(
		[]int{1, 0, 3, 4, 2, 7, 9, 9},
		[]int{1, 3, 5, 7, 9, 11},
	))
	// Output:
	// [1 3 7 9]
}

func ExampleEditScript() {
	lhs := strings.Fields("a stitch in time saves nine")
	rhs := strings.Fields("we live in a time of nine lives")

	i := 0
	fmt.Println("start", lhs)
	var out []string
	for _, e := range slice.EditScript(lhs, rhs) {
		switch e.Op {
		case slice.OpDrop:
			i += e.N
		case slice.OpEmit:
			fmt.Println("emit", lhs[i:i+e.N])
			out = append(out, lhs[i:i+e.N]...)
			i += e.N
		case slice.OpCopy:
			fmt.Println("copy", rhs[e.X:e.X+e.N])
			out = append(out, rhs[e.X:e.X+e.N]...)
		case slice.OpReplace:
			fmt.Println("replace", lhs[i:i+e.N], "with", rhs[e.X:e.X+e.N])
			out = append(out, rhs[e.X:e.X+e.N]...)
			i += e.N
		default:
			panic("invalid")
		}
	}
	fmt.Println("end", out)
	// Output:
	// start [a stitch in time saves nine]
	// copy [we live in]
	// emit [a]
	// emit [time]
	// replace [saves] with [of]
	// emit [nine]
	// copy [lives]
	// end [we live in a time of nine lives]
}
