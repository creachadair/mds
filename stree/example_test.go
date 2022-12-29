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

func (p Pair) less(q Pair) bool { return p.X < q.X }

func ExampleTree_Add() {
	tree := stree.New[string](200, stringLess)
	tree.Add("never")
	tree.Add("say")
	tree.Add("never")
	fmt.Println("tree.Len() =", tree.Len())
	// Output:
	// tree.Len() = 2
}

func ExampleTree_Remove() {
	const key = "Aloysius"
	tree := stree.New[string](1, stringLess)
	fmt.Println("inserted:", tree.Add(key))
	fmt.Println("removed:", tree.Remove(key))
	fmt.Println("re-removed:", tree.Remove(key))
	// Output:
	// inserted: true
	// removed: true
	// re-removed: false
}

func ExampleTree_Get() {
	tree := stree.New(1, Pair.less, []Pair{{
		X: "mom",
		V: 1,
	}}...)
	hit, ok := tree.Get(Pair{X: "mom"})
	fmt.Printf("%v, %v\n", hit.V, ok)
	miss, ok := tree.Get(Pair{X: "dad"})
	fmt.Printf("%v, %v\n", miss.V, ok)
	// Output:
	// 1, true
	// 0, false
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
