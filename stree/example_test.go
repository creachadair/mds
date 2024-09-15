package stree_test

import (
	"cmp"
	"fmt"
	"strings"

	"github.com/creachadair/mds/stree"
)

type Pair struct {
	X string
	V int
}

func (p Pair) Compare(q Pair) int { return cmp.Compare(p.X, q.X) }

func ExampleTree_Add() {
	tree := stree.New(200, strings.Compare)

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
	tree := stree.New(1, strings.Compare)

	fmt.Println("inserted:", tree.Add(key))
	fmt.Println("removed:", tree.Remove(key))
	fmt.Println("re-removed:", tree.Remove(key))
	// Output:
	// inserted: true
	// removed: true
	// re-removed: false
}

func ExampleTree_Get() {
	tree := stree.New(1, Pair.Compare,
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
	tree := stree.New(15, strings.Compare, "eat", "those", "bloody", "vegetables")
	for key := range tree.Inorder {
		fmt.Println(key)
	}
	// Output:
	// bloody
	// eat
	// those
	// vegetables
}

func ExampleTree_Min() {
	tree := stree.New(50, cmp.Compare[int], 1814, 1956, 955, 1066, 2016)

	fmt.Println("len:", tree.Len())
	fmt.Println("min:", tree.Min())
	fmt.Println("max:", tree.Max())
	// Output:
	// len: 5
	// min: 955
	// max: 2016
}

func ExampleKV() {
	// For brevity, it can be helpful to define a type alias for your items.

	type item = stree.KV[int, string]

	tree := stree.New(100, item{}.Compare(cmp.Compare))
	tree.Add(item{1, "one"})
	tree.Add(item{2, "two"})
	tree.Add(item{3, "three"})
	tree.Add(item{4, "four"})

	for _, i := range []int{1, 3, 2} {
		fmt.Println(tree.Cursor(item{Key: i}).Key().Value)
	}
	// Output:
	// one
	// three
	// two
}
