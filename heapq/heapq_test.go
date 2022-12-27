package heapq_test

import (
	"math/rand"
	"sort"
	"testing"

	"github.com/creachadair/mds/heapq"
	"github.com/google/go-cmp/cmp"
)

func TestHeap(t *testing.T) {
	q := heapq.New(func(a, b int) bool {
		return a < b
	})
	check := func(want ...int) {
		t.Helper()
		var got []int
		q.Each(func(v int) bool {
			got = append(got, v)
			return true
		})
		if diff := cmp.Diff(want, got); diff != "" {
			t.Errorf("Queue contents (+want, -got):\n%s", diff)
			t.Logf("Got:  %v", got)
			t.Logf("Want: %v", want)
		}
		if len(want) != 0 {
			if got := q.Front(); got != want[0] {
				t.Errorf("Front: got %v, want %v", got, want[0])
			}
		}
		if got := q.Len(); got != len(want) {
			t.Errorf("Len: got %v, want %v", got, len(want))
		}
	}
	checkAdd := func(v, want int) {
		if got := q.Add(v); got != want {
			t.Errorf("Add(%v): got %v, want %v", v, got, want)
		}
	}
	checkPop := func(want int, wantok bool) {
		got, ok := q.Pop()
		if got != want || ok != wantok {
			t.Errorf("Pop: got (%v, %v), want (%v, %v)", got, ok, want, wantok)
		}
	}

	check()
	checkPop(0, false)

	checkAdd(10, 0)
	check(10)
	checkAdd(5, 0)
	check(5, 10)
	checkAdd(3, 0)
	check(3, 5, 10)
	checkAdd(4, 1)
	check(3, 4, 10, 5)
	q.Update(1, 0)
	check(0, 3, 10, 5)
	checkPop(0, true)

	checkPop(3, true)
	checkPop(5, true)
	checkPop(10, true)
	checkPop(0, false)
	check()

	q.Set([]int{1, 2, 3, 4, 5})
	check(1, 2, 3, 4, 5)
	q.Update(2, 0)
	check(0, 1, 2, 4, 5)
	q.Update(1, 6)
	check(0, 4, 2, 6, 5)
	checkPop(0, true)

	q.Clear()
	check()

	q.Set([]int{15, 3, 9, 4, 8, 2, 11, 20, 11, 17, 1})
	check(1, 3, 2, 4, 8, 9, 11, 20, 11, 17, 15) // constructed by hand
	checkOrdered(t, q)
}

func TestOrder(t *testing.T) {
	const inputSize = 500
	const inputRange = 1000
	const numTests = 5

	for i := 0; i < numTests; i++ {
		input := make([]int, inputSize)
		for i := range input {
			input[i] = rand.Intn(inputRange) - (inputRange / 2)
		}

		q := heapq.New(func(i, j int) bool { return i < j })
		q.Set(input)

		checkOrdered(t, q)
	}
}

func checkOrdered(t *testing.T, q *heapq.Queue[int]) {
	t.Helper()
	var all []int
	for !q.IsEmpty() {
		all = append(all, q.Front())
		q.Pop()
	}
	if !sort.IntsAreSorted(all) {
		t.Errorf("Queue contents out of order: %v", all)
	}
}
