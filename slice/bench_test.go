package slice_test

import (
	"flag"
	"math/rand/v2"
	"testing"

	"github.com/creachadair/mds/slice"
)

var (
	lhsSize = flag.Int("lhs", 500, "LHS input size (number of elements)")
	rhsSize = flag.Int("rhs", 500, "RHS input size (number of elements)")
)

func BenchmarkEdit(b *testing.B) {
	lhs := make([]int, *lhsSize)
	for i := range lhs {
		lhs[i] = rand.IntN(1000000)
	}
	rhs := make([]int, *rhsSize)
	for i := range rhs {
		rhs[i] = rand.IntN(10000000)
	}

	b.Run("LCS", func(b *testing.B) {
		b.ReportAllocs()
		for i := 0; i < b.N; i++ {
			_ = slice.LCS(lhs, rhs)
		}
	})
	b.Run("EditScript", func(b *testing.B) {
		b.ReportAllocs()
		for i := 0; i < b.N; i++ {
			_ = slice.EditScript(lhs, rhs)
		}
	})
}
