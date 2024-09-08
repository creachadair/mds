package ring_test

import (
	"fmt"

	"github.com/creachadair/mds/ring"
)

func ExampleRing() {
	r := ring.Of("time", "flies", "like", "an", "arrow")

	// Set the value of an existing element.
	r.Value = "fruit"

	// Splice new elements into a ring.
	s := r.At(2).Join(ring.Of("a", "banana"))

	// Splice existing elements out of a ring.
	s.Prev().Join(r)

	// Iterate over the elements of a ring.
	r.Each(func(s string) bool {
		fmt.Println(s)
		return true
	})

	// Output:
	// fruit
	// flies
	// like
	// a
	// banana
}
