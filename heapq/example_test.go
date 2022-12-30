package heapq_test

import (
	"fmt"
	"math/rand"
	"sort"

	"github.com/creachadair/mds/heapq"
)

func Example_InPlaceHeapSort() {
	buf := make([]int, 25)
	for i := range buf {
		buf[i] = rand.Intn(100) - 25
	}

	q := heapq.NewWithData(func(a, b int) bool {
		return a > b
	}, buf)

	for i := len(buf) - 1; !q.IsEmpty(); i-- {
		v, _ := q.Pop()
		buf[i] = v
	}
	fmt.Println(sort.IntsAreSorted(buf))
	// Output:
	// true
}
