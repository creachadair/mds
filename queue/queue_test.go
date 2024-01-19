package queue_test

import (
	"flag"
	"math/rand"
	"testing"

	"github.com/creachadair/mds/queue"
	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
)

func checkQueue[T any](t *testing.T, q *queue.Queue[T], want []T) {
	t.Helper()
	got := q.Slice()
	if diff := cmp.Diff(want, got, cmpopts.EquateEmpty()); diff != "" {
		t.Errorf("Wrong contents (-got, +want):\n%s", diff)
	}
	if n := q.Len(); n != len(got) || n != len(want) {
		t.Errorf("Wrong length: got %d, want %d == %d", n, len(got), len(want))
	}
}

func TestQueue(t *testing.T) {
	var q queue.Queue[int]
	check := func(want ...int) { checkQueue(t, &q, want) }

	// Front and Pop of an empty queue report no value.
	if v := q.Front(); v != 0 {
		t.Errorf("Front: got %v, want 0", v)
	}
	if v, ok := q.Pop(); ok {
		t.Errorf("Pop: got (%v, %v), want (0, false)", v, ok)
	}

	check()
	if !q.IsEmpty() {
		t.Error("IsEmpty is incorrectly false")
	}
	if n := q.Len(); n != 0 {
		t.Errorf("Len: got %d, want 0", n)
	}

	q.Add(1)
	if q.IsEmpty() {
		t.Error("IsEmpty is incorrectly true")
	}
	check(1)

	q.Add(2)
	check(1, 2)

	q.Add(3)
	check(1, 2, 3)
	if n := q.Len(); n != 3 {
		t.Errorf("Len: got %d, want 3", n)
	}

	front := q.Front()
	if front != 1 {
		t.Errorf("Front: got %v, want 1", front)
	}
	if v, ok := q.Peek(0); !ok || v != front {
		t.Errorf("Peek(0): got (%v, %v), want (%v, true)", v, ok, front)
	}
	if v, ok := q.Peek(1); !ok || v != 2 {
		t.Errorf("Peek(1): got (%v, %v), want (2, true)", v, ok)
	}
	if v, ok := q.Peek(10); ok {
		t.Errorf("Peek(10): got (%v, %v), want (0, false)", v, ok)
	}

	if v, ok := q.Pop(); !ok || v != front {
		t.Errorf("Pop: got (%v, %v), want (%v, true)", v, ok, front)
	}
	check(2, 3)

	q.Clear()
	check()
}

var doDebug = flag.Bool("debug", false, "Enable debug logging")

func TestQueueRandom(t *testing.T) {
	var q queue.Queue[int]

	debug := func(msg string, args ...any) {
		if *doDebug {
			t.Logf(msg, args...)
		}
	}

	// The "has" slice is an "awful" queue that grows indefinitely with use, but
	// serves to confirm that the real implementation gets the right order.
	var has []int
	var stats struct {
		MaxLen   int
		NumAdd   int
		NumPop   int
		NumClear int
	}
	get := func(z int) int {
		if z < 0 || z >= len(has) {
			return -1
		}
		return has[z]
	}

	// Run a bunch of operations at random on the q, and verify that we get the
	// right values out of its methods.
	const (
		doAdd   = 45
		doPop   = doAdd + 45
		doPeek  = doPop + 3
		doFront = doPeek + 3
		doLen   = doFront + 3
		doClear = doLen + 1

		doTotal = doClear
	)

	for i := 0; i < 5000; i++ {
		if len(has) > stats.MaxLen {
			stats.MaxLen = len(has)
		}
		checkQueue(t, &q, has)
		switch op := rand.Intn(doTotal); {
		case op < doAdd:
			stats.NumAdd++
			r := rand.Intn(1000)
			has = append(has, r)
			debug("Add(%d)", r)
			q.Add(r)
		case op < doPop:
			stats.NumPop++
			debug("Pop exp=%d", get(0))
			got, ok := q.Pop()
			if len(has) == 0 {
				if ok {
					t.Errorf("Pop: got (%v, %v), want (0, false)", got, ok)
				}
				continue
			}
			want := has[0]
			has = has[1:]
			if !ok || got != want {
				t.Errorf("Pop: got (%v, %v), want (%v, true)", got, ok, want)
			}
		case op < doLen:
			debug("Len n=%d", len(has))
			if got := q.Len(); got != len(has) {
				t.Errorf("Len: got %d, want %d", got, len(has))
			}
		case op < doFront:
			debug("Front exp=%d", get(0))
			if got := q.Front(); len(has) != 0 && got != has[0] {
				t.Errorf("Front: got %d, want %d", got, has[0])
			}
		case op < doPeek:
			if len(has) != 0 {
				r := rand.Intn(len(has))
				debug("Peek(%d) exp=%d", r, has[r])
				if got, ok := q.Peek(r); !ok || got != has[r] {
					t.Errorf("Peek(%d): got (%d, %v), want (%d, true)", r, got, ok, has[r])
				}
			}
		case op < doClear:
			stats.NumClear++
			debug("Clear n=%d", len(has))
			has = has[:0]
			q.Clear()
		default:
			panic("unexpected")
		}
	}
	t.Logf("Queue at exit (n=%d): %v", q.Len(), q.Slice())
	t.Logf("Stats: %+v", stats)
}
