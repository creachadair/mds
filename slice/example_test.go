package slice_test

import (
	"fmt"
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

func ExampleSplit() {
	vs := []int{1, 2, 3, 4, 5, 6, 7, 8}

	fmt.Println("input:", vs)
	first3, tail := slice.Split(vs, 3)
	head, last3 := slice.Split(vs, -3)
	fmt.Println("first3:", first3, "tail:", tail)
	fmt.Println("last3:", last3, "head:", head)
	// Output:
	// input: [1 2 3 4 5 6 7 8]
	// first3: [1 2 3] tail: [4 5 6 7 8]
	// last3: [6 7 8] head: [1 2 3 4 5]
}

func ExampleMatchingKeys() {
	vs := map[string]int{"red": 3, "yellow": 6, "blue": 4, "green": 5}

	for key := range slice.MatchingKeys(vs, isEven) {
		fmt.Println(key, vs[key])
	}
	// Unordered output:
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
	lhs := strings.Fields("if you mix red with green you get blue")
	rhs := strings.Fields("red mixed with green does not give blue at all")

	fmt.Println("start", lhs)
	var out []string
	for _, e := range slice.EditScript(lhs, rhs) {
		switch e.Op {
		case slice.OpDrop:
			fmt.Println("drop", e.X)
		case slice.OpEmit:
			fmt.Println("emit", e.X)
			out = append(out, e.X...)
		case slice.OpCopy:
			fmt.Println("copy", e.Y)
			out = append(out, e.Y...)
		case slice.OpReplace:
			fmt.Println("replace", e.X, "with", e.Y)
			out = append(out, e.Y...)
		default:
			panic("invalid")
		}
	}
	fmt.Println("end", out)
	// Output:
	// start [if you mix red with green you get blue]
	// drop [if you mix]
	// emit [red]
	// copy [mixed]
	// emit [with green]
	// replace [you get] with [does not give]
	// emit [blue]
	// copy [at all]
	// end [red mixed with green does not give blue at all]
}
