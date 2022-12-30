package heapq_test

import (
	"fmt"
	"math/rand"
	"sort"

	"github.com/creachadair/mds/heapq"
)

func Example_inPlaceHeapSort() {
	// Fill a buffer with some random data for example purposes.
	buf := make([]int, 25)
	for i := range buf {
		buf[i] = rand.Intn(100) - 25
	}
	fmt.Println("sorted before:", sort.IntsAreSorted(buf))

	// Put the data into a heap. Note that the comparison function here puts
	// greater elements at the top.
	q := heapq.NewWithData(func(a, b int) bool {
		return a > b
	}, buf)

	// Pull the items off the heap. Because the heap is using the buffer we
	// already allocated, each item removed leaves an empty slot at the end of
	// the array.
	for i := len(buf) - 1; !q.IsEmpty(); i-- {
		v, _ := q.Pop()
		buf[i] = v // N.B. Order matters here: Pop before update!
	}
	fmt.Println("sorted after:", sort.IntsAreSorted(buf))
	// Output:
	// sorted before: false
	// sorted after: true
}
