package mlink_test

import (
	"testing"

	"github.com/creachadair/mds/mlink"
	"github.com/creachadair/mtest"
)

func eq(z int) func(int) bool {
	return func(n int) bool { return n == z }
}

func TestList(t *testing.T) {
	lst := mlink.NewList[int]()
	checkList := checker(t, lst)
	advance := func(c *mlink.Cursor[int], wants ...int) {
		t.Helper()
		for _, want := range wants {
			c.Next()
			if got := c.Get(); got != want {
				t.Errorf("Get: got %v, want %v", got, want)
			}
		}
	}
	checkAt := func(c *mlink.Cursor[int], want int) {
		t.Helper()
		if got := c.Get(); got != want {
			t.Errorf("Get: got %v, want %v", got, want)
		}
	}

	// A new list is initially empty.
	checkList()

	if !lst.At(0).AtEnd() {
		t.Error("At(0): should be at end for empty list")
	}
	if !lst.Last().AtEnd() {
		t.Error("Last: should be at end for empty list")
	}

	// Add advances after insertion.
	first := lst.At(0)
	first.Add(1, 2)
	checkList(1, 2)

	// Push does not advance after insertion.
	first.Push(3)
	first.Push(4)
	checkList(1, 2, 4, 3)
	if got, want := first.Remove(), 4; got != want {
		t.Errorf("Remove: got %v, want %v", got, want)
	}
	checkList(1, 2, 3)

	// Check adding at the end.
	lst.End().Add(4)
	checkList(1, 2, 3, 4)

	// At and Last should work.
	checkAt(lst.At(0), 1)
	checkAt(lst.At(1), 2)
	checkAt(lst.At(2), 3)
	checkAt(lst.At(3), 4)
	checkAt(lst.Last(), 4)

	// At past the end of the list should pin to the end.
	if e := lst.At(1000); !e.AtEnd() {
		t.Error("At(big) should go to the end of the list")
	}
	// Peek past the end should report failure and a zero.
	if v, ok := lst.Peek(1000); ok || v != 0 {
		t.Errorf("Peek(big): got (%v, %v), want (0, false)", v, ok)
	}

	// Exercise navigation with a cursor.
	c := lst.At(0)
	checkAt(c, 1)
	advance(c, 2)

	if got, want := c.Remove(), 2; got != want {
		t.Errorf("Remove: got %v, want %v", got, want)
	}
	checkList(1, 3, 4)
	checkAt(c, 3)

	// Add at the ends of the list.
	lst.End().Add(5)
	lst.At(0).Add(6)
	checkList(6, 1, 3, 4, 5)

	// The cursor should still be valid, and see the changes.
	advance(c, 4)

	// Add in the middle of the list.
	c.Push(7)
	checkList(6, 1, 3, 7, 4, 5)

	// Exercise moving in a list.
	c = lst.At(0)
	checkAt(c, 6)
	advance(c, 1, 3, 7)

	// Copy c and verify it moves separately.
	cp := c.Copy()

	lst.End().Add(8)
	advance(cp, 4, 5, 8, 0)
	if !cp.AtEnd() {
		t.Error("Cursor copy should be at the end")
	}

	c.Truncate()
	checkList(6, 1, 3)
	if !c.AtEnd() {
		t.Error("Cursor should be at the end")
	}

	// Push at the end does the needful.
	lst.End().Push(9)
	checkList(6, 1, 3, 9)
	checkAt(c, 9) // c sees the new value

	// Setting the last element doesn't add any mass.
	lst.Last().Set(10)
	checkList(6, 1, 3, 10)
	checkAt(c, 10) // c sees the new value

	// Setting elsewhere works too.
	lst.At(0).Set(11)
	checkList(11, 1, 3, 10)

	// Setting at the end pushes a new item.
	tail := lst.End()
	tail.Set(12)
	checkAt(tail, 12) // tail is no longer at the end
	checkList(11, 1, 3, 10, 12)
	checkAt(c, 10) // c hasn't moved

	// Finding things.
	checkAt(lst.Find(eq(11)), 11)            // first
	checkAt(lst.Find(eq(3)), 3)              // middle
	checkAt(lst.Find(eq(12)), 12)            // last
	if q := lst.Find(eq(-999)); !q.AtEnd() { // missing
		t.Errorf("FInd: got %v, wanted no result", q.Get())
	}

	// Remove an item, and verify that a cursor to the following item is
	// correctly invalidated.
	d := c.Copy()
	d.Next()
	checkAt(d, 12)
	if got, want := c.Remove(), 10; got != want {
		t.Errorf("Remove: got %v, want %v", got, want)
	}
	checkList(11, 1, 3, 12)
	if !d.AtEnd() {
		t.Errorf("Sequent cursor should be AtEnd after deletion")
	}

	// We can remove the first and last elements.
	if got, want := lst.At(0).Remove(), 11; got != want {
		t.Errorf("Remove: got %v, want %v", got, want)
	}
	checkList(1, 3, 12)
	if got, want := lst.Last().Remove(), 12; got != want {
		t.Errorf("Remove: got %v, want %v", got, want)
	}
	checkList(1, 3)

	lst.Clear()
	checkList()
}

func mustPanic(f func()) func(*testing.T) {
	return func(t *testing.T) {
		t.Helper()
		mtest.MustPanic(t, f)
	}
}

func TestPanics(t *testing.T) {
	var lst mlink.List[bool]

	t.Run("At(-1)", mustPanic(func() {
		lst.At(-1)
	}))
	t.Run("Peek(-1)", mustPanic(func() {
		lst.Peek(-1)
	}))
	t.Run("NilCursor", mustPanic(func() {
		var nc *mlink.Cursor[bool]
		nc.Get()
	}))
}
