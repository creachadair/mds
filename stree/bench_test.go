package stree_test

import (
	"fmt"
	"math"
	"math/rand/v2"
	"sort"
	"testing"

	"github.com/creachadair/mds/stree"
)

const benchSeed = 1471808909908695897

// Trial values of β for load-testing tree operations.
var balances = []int{0, 50, 100, 150, 200, 250, 300, 500, 800, 1000}

func intCompare(a, b int) int { return a - b }

func randomTree(b *testing.B, β int) (*stree.Tree[int], []int) {
	rng := rand.New(rand.NewPCG(benchSeed, benchSeed))
	values := make([]int, b.N)
	for i := range values {
		values[i] = rng.IntN(math.MaxInt32)
	}
	return stree.New(β, intCompare, values...), values
}

func BenchmarkNew(b *testing.B) {
	for _, β := range balances {
		b.Run(fmt.Sprintf("β=%d", β), func(b *testing.B) {
			randomTree(b, β)
		})
	}
}

func BenchmarkAddRandom(b *testing.B) {
	for _, β := range balances {
		b.Run(fmt.Sprintf("β=%d", β), func(b *testing.B) {
			_, values := randomTree(b, β)
			b.ResetTimer()
			tree := stree.New[int](β, intCompare)
			for _, v := range values {
				tree.Add(v)
			}
		})
	}
}

func BenchmarkAddOrdered(b *testing.B) {
	for _, β := range balances {
		b.Run(fmt.Sprintf("β=%d", β), func(b *testing.B) {
			tree := stree.New[int](β, intCompare)
			for i := 1; i <= b.N; i++ {
				tree.Add(i)
			}
		})
	}
}

func BenchmarkRemoveRandom(b *testing.B) {
	for _, β := range balances {
		b.Run(fmt.Sprintf("β=%d", β), func(b *testing.B) {
			tree, values := randomTree(b, β)
			b.ResetTimer()
			for _, v := range values {
				tree.Remove(v)
			}
		})
	}
}

func BenchmarkRemoveOrdered(b *testing.B) {
	for _, β := range balances {
		b.Run(fmt.Sprintf("β=%d", β), func(b *testing.B) {
			tree, values := randomTree(b, β)
			sort.Slice(values, func(i, j int) bool {
				return values[i] < values[j]
			})
			b.ResetTimer()
			for _, v := range values {
				tree.Remove(v)
			}
		})
	}
}

func BenchmarkLookup(b *testing.B) {
	for _, β := range balances {
		b.Run(fmt.Sprintf("β=%d", β), func(b *testing.B) {
			tree, values := randomTree(b, β)
			b.ResetTimer()
			for _, v := range values {
				tree.Get(v)
			}
		})
	}
}
