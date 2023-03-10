package heapq_test

import (
	"math/rand"
	"sort"
	"testing"

	"github.com/creachadair/mds/heapq"
	"github.com/google/go-cmp/cmp"
)

func intLess(a, b int) bool { return a < b }

func TestHeap(t *testing.T) {
	t.Run("New", func(t *testing.T) {
		runTests(t, heapq.New(intLess))
	})
	t.Run("NewWithData", func(t *testing.T) {
		buf := make([]int, 0, 64)
		runTests(t, heapq.NewWithData(intLess, buf))
	})
}

func runTests(t *testing.T, q *heapq.Queue[int]) {
	t.Helper()

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
	checkPop(3, true)

	checkPop(4, true)
	checkPop(5, true)
	checkPop(10, true)
	checkPop(0, false)
	check()

	q.Set([]int{1, 2, 3, 4, 5})
	check(1, 2, 3, 4, 5)
	checkPop(1, true)

	q.Clear()
	check()

	q.Set([]int{15, 3, 9, 4, 8, 2, 11, 20, 11, 17, 1})
	check(1, 3, 2, 4, 8, 9, 11, 20, 11, 17, 15) // constructed by hand
	if got := extract(q); !sort.IntsAreSorted(got) {
		t.Errorf("Queue contents are out of order: %v", got)
	}
}

func TestOrder(t *testing.T) {
	const inputSize = 5000
	const inputRange = 100000

	makeInput := func() []int {
		input := make([]int, inputSize)
		for i := range input {
			input[i] = rand.Intn(inputRange) - (inputRange / 2)
		}
		return input
	}

	t.Run("Ascending", func(t *testing.T) {
		q := heapq.New(func(i, j int) bool { return i < j })
		q.Set(makeInput())
		if got := extract(q); !sort.IntsAreSorted(got) {
			t.Errorf("Queue contents are out of order: %v", got)
		}
	})

	t.Run("Descending", func(t *testing.T) {
		q := heapq.New(func(i, j int) bool { return i > j })
		q.Set(makeInput())
		got := extract(q)
		if !sort.IsSorted(sort.Reverse(sort.IntSlice(got))) {
			t.Errorf("Queue contents are out of order: %v", got)
		}
	})

	t.Run("Reorder", func(t *testing.T) {
		q := heapq.New(func(i, j int) bool { return i < j })
		q.Set([]int{17, 3, 11, 2, 7, 5, 13})
		if got, want := q.Front(), 2; got != want {
			t.Errorf("Front: got %v, want %v", got, want)
		}

		q.Reorder(func(i, j int) bool { return i > j })
		if got, want := q.Front(), 17; got != want {
			t.Errorf("Front: got %v, want %v", got, want)
		}

		got := extract(q)
		if !sort.IsSorted(sort.Reverse(sort.IntSlice(got))) {
			t.Errorf("Results are out of order: %v", got)
		}
	})
}

func TestNewWithData(t *testing.T) {
	const bufSize = 100 // N.B. must be even, so we can fill halves

	// Preallocate a buffer and populate part of it with some data.
	buf := make([]int, 0, bufSize)

	var want []int
	for i := 0; i < bufSize/2; i++ {
		z := rand.Intn(500) - 250
		buf = append(buf, z)
		want = append(want, z) // keep track of what we added.
	}

	// Give buf over to the queue, then add more stuff so we can check that the
	// queue took over the array correctly.
	q := heapq.NewWithData(intLess, buf)

	// Add some more stuff via the queue.
	for i := 0; i < bufSize/2; i++ {
		z := rand.Intn(500) - 250
		q.Add(z)
		want = append(want, z)
	}

	// Check that the queue used the same array.  You are specifically NOT
	// supposed to do this, messing with the array outside the queue, but here
	// we need to check that the queue did the right thing.
	got := buf[:len(want)]
	sort.Ints(got)
	sort.Ints(want)

	if diff := cmp.Diff(want, got); diff != "" {
		t.Errorf("Queue contents (+want, -got):\n%s", diff)
	}
}

func TestSort(t *testing.T) {
	longIn := make([]int, 50)
	for i := range longIn {
		longIn[i] = rand.Intn(1000) - 250
	}
	longOut := make([]int, len(longIn))
	copy(longOut, longIn)
	sort.Ints(longOut)

	lt := func(a, b int) bool { return a < b }
	gt := func(a, b int) bool { return a > b }
	_, _ = lt, gt
	tests := []struct {
		name        string
		cmp         func(a, b int) bool
		input, want []int
	}{
		{"Nil", lt, nil, nil},
		{"Empty", lt, []int{}, nil},
		{"Single-LT", lt, []int{11}, []int{11}},
		{"Single-GT", gt, []int{11}, []int{11}},
		{"Ascend", lt, []int{9, 1, 4, 11}, []int{1, 4, 9, 11}},
		{"Descend", gt, []int{9, 1, 4, 11}, []int{11, 9, 4, 1}},
		{"Long", lt, longIn, longOut},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			in := append([]int(nil), tc.input...)
			heapq.Sort(tc.cmp, in)
			if diff := cmp.Diff(tc.want, in); diff != "" {
				t.Errorf("Sort (-want, +got):\n%s", diff)
			}
		})
	}
}

func extract[T any](q *heapq.Queue[T]) []T {
	all := make([]T, 0, q.Len())
	for !q.IsEmpty() {
		all = append(all, q.Front())
		q.Pop()
	}
	return all
}
