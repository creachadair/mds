package mapset_test

import (
	"fmt"
	"sort"
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
	sort.Strings(elts)
	fmt.Println(strings.Join(elts, "\n"))
	// Output:
	// t intersects u: true
	// t equals u: false
	//
	// canal
	// future
	// man
	// panama
	// plan
}
