package mapset_test

import (
	"fmt"
	"strings"

	"github.com/creachadair/mds/mapset"
)

func Example() {
	s := mapset.New(strings.Fields("a man a plan")...)

	// Add individual elements.
	s.Add("panama", "canal")

	// Add the contents of another set.
	t := mapset.New("plan", "for", "the", "future")
	s.AddAll(t)

	// Remove items and convert to a slice.
	elts := s.Remove("a", "an", "the", "for").Slice()

	// Clone and make other changes.
	u := s.Clone().Remove("future", "plans")

	// Do some basic comparisons.
	fmt.Println("t intersects u:", t.Intersects(u))
	fmt.Println("t equals u:", t.Equals(u))
	fmt.Println()

	// The slice is unordered, so impose some discipline.
	fmt.Println(strings.Join(elts, "\n"))
	// Unordered output:
	// t intersects u: true
	// t equals u: false
	//
	// canal
	// future
	// man
	// panama
	// plan
}

func ExampleKeys() {
	s := mapset.Keys(map[string]int{
		"apple":  1,
		"pear":   2,
		"plum":   3,
		"cherry": 4,
	})

	fmt.Println(strings.Join(s.Slice(), "\n"))
	// Unordered output:
	// apple
	// cherry
	// pear
	// plum
}

func ExampleValues() {
	s := mapset.Values(map[string]int{
		"apple":  5,
		"pear":   4,
		"plum":   4,
		"cherry": 6,
	})

	for _, v := range s.Slice() {
		fmt.Println(v)
	}
	// Unordered output:
	// 4
	// 5
	// 6
}
