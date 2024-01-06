package heapq_test

import (
	"fmt"

	"github.com/creachadair/mds/heapq"
)

func ExampleNew() {
	q := heapq.New(intCompare)

	q.Add(8)
	q.Add(6)
	q.Add(7)
	q.Add(5)
	q.Add(3)
	q.Add(0)
	q.Add(9)

	fmt.Println("length before", q.Len())
	fmt.Println(q.Pop())
	fmt.Println(q.Pop())
	fmt.Println(q.Pop())
	fmt.Println("length after", q.Len())
	fmt.Println("front", q.Front())

	// Output:
	// length before 7
	// 0 true
	// 3 true
	// 5 true
	// length after 4
	// front 6
}
