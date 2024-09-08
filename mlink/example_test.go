package mlink_test

import (
	"fmt"
	"strings"

	"github.com/creachadair/mds/mlink"
)

func ExampleList() {
	var lst mlink.List[string]

	lst.At(0).Add(strings.Fields("A is for Amy who fell down the stairs")...)

	// Find the first element of the list matching a predicate.
	name := lst.Find(func(s string) bool {
		return s == "Amy"
	})

	name.Set("Amélie") // change the value
	name.Next()        // move to the next element
	name.Next()        // ...and again
	name.Truncate()    // discard the rest of the list
	name.Set("ran")    // add a new last element
	name.Push("far")   // push a new element

	// Add a new element at the end.
	lst.End().Add("away")

	// Remove the element we previously pushed.
	name.Remove()

	// Print out everything in the list.
	lst.Each(func(s string) bool {
		fmt.Print(" ", s)
		return true
	})
	fmt.Println()

	// Calculate the length of the list.
	fmt.Println(lst.Len(), "items")

	// Output:
	// A is for Amélie who ran away
	// 7 items
}

func ExampleList_Find() {
	lst := mlink.NewList[int]()
	lst.At(0).Add(2, 4, 6, 7, 8, 10)

	if odd := lst.Find(isOdd); !odd.AtEnd() {
		fmt.Printf("Found %d\n", odd.Get())
		odd.Remove()
	}
	if odd := lst.Find(isOdd); odd.AtEnd() {
		fmt.Println("no more odds")
	}
	// Output:
	// Found 7
	// no more odds
}

func isOdd(z int) bool { return z%2 == 1 }

func ExampleCursor_Next() {
	var lst mlink.List[string]
	lst.At(0).Add("apples", "pears", "plums", "cherries")

	cur := lst.At(0)
	for !cur.AtEnd() {
		fmt.Print(cur.Get())
		if cur.Next() {
			fmt.Print(", ")
		}
	}
	fmt.Println()
	// Output:
	// apples, pears, plums, cherries
}
