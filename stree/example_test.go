package stree_test

import (
	"fmt"

	"github.com/creachadair/mds/stree"
)

func stringLess(a, b string) bool { return a < b }

type Pair struct {
	X string
	V int
}

func (p Pair) Less(q Pair) bool { return p.X < q.X }

func ExampleTree_Add() {
	tree := stree.New(200, stringLess)

	fmt.Println("inserted:", tree.Add("never"))
	fmt.Println("inserted:", tree.Add("say"))
	fmt.Println("re-inserted:", tree.Add("never"))
	fmt.Println("items:", tree.Len())
	// Output:
	// inserted: true
	// inserted: true
	// re-inserted: false
	// items: 2
}

func ExampleTree_Remove() {
	const key = "Aloysius"
	tree := stree.New(1, stringLess)

	fmt.Println("inserted:", tree.Add(key))
	fmt.Println("removed:", tree.Remove(key))
	fmt.Println("re-removed:", tree.Remove(key))
	// Output:
	// inserted: true
	// removed: true
	// re-removed: false
}

func ExampleTree_Get() {
	tree := stree.New(1, Pair.Less,
		Pair{X: "angel", V: 5},
		Pair{X: "devil", V: 7},
		Pair{X: "human", V: 13},
	)

	for _, key := range []string{"angel", "apple", "human"} {
		hit, ok := tree.Get(Pair{X: key})
		fmt.Println(hit.V, ok)
	}
	// Output:
	// 5 true
	// 0 false
	// 13 true
}

func ExampleTree_Inorder() {
	tree := stree.New(15, stringLess, "eat", "those", "bloody", "vegetables")
	tree.Inorder(func(key string) bool {
		fmt.Println(key)
		return true
	})
	// Output:
	// bloody
	// eat
	// those
	// vegetables
}

func ExampleTree_Min() {
	tree := stree.New(50, intLess, 1814, 1956, 955, 1066, 2016)

	fmt.Println("len:", tree.Len())
	fmt.Println("min:", tree.Min())
	fmt.Println("max:", tree.Max())
	// Output:
	// len: 5
	// min: 955
	// max: 2016
}
