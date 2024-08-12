package slice_test

import (
	"cmp"
	"flag"
	"fmt"
	"math/rand/v2"
	"testing"

	"github.com/creachadair/mds/slice"
)

var (
	lhsSize = flag.Int("lhs", 500, "LHS input size (number of elements)")
	rhsSize = flag.Int("rhs", 500, "RHS input size (number of elements)")

	lisSize = flag.Int("lis", 0, "LIS input size (number of elements)")
	// lisCountCmps is optional because it requires inserting some
	// accounting in the algorithm's inner loop, so while you get a
	// number that's independent of the nonlinear log(n) time factor,
	// the uncompensated numbers get thrown off.
	lisCountCmps = flag.Bool("lis-count-cmp", false, "report ns/compare stats")
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
		for range b.N {
			_ = slice.LCS(lhs, rhs)
		}
	})
	b.Run("EditScript", func(b *testing.B) {
		b.ReportAllocs()
		for range b.N {
			_ = slice.EditScript(lhs, rhs)
		}
	})
}

func BenchmarkLNDSFunc(b *testing.B) {
	buckets := []int{100, 1000, 20000}
	if *lisSize != 0 {
		buckets = []int{*lisSize}
	}

	for _, bucket := range buckets {
		b.Run(fmt.Sprint("items=", bucket), func(b *testing.B) {
			input := randomInts(bucket)

			var comparisons uint64
			cmpFn := cmp.Compare[int]
			if *lisCountCmps {
				cmpFn = func(a, b int) int {
					comparisons++
					return cmp.Compare(a, b)
				}
			}

			b.ReportAllocs()
			for range b.N {
				_ = slice.LNDSFunc(input, cmpFn)
			}

			if *lisCountCmps {
				perCmp := float64(b.Elapsed().Nanoseconds()) / float64(comparisons)
				b.ReportMetric(perCmp, "ns/cmp")
			}
		})
	}
}

func BenchmarkLISFunc(b *testing.B) {
	buckets := []int{100, 1000, 20000}
	if *lisSize != 0 {
		buckets = []int{*lisSize}
	}

	for _, bucket := range buckets {
		b.Run(fmt.Sprint("items=", bucket), func(b *testing.B) {
			input := randomInts(bucket)

			var comparisons uint64
			cmpFn := cmp.Compare[int]
			if *lisCountCmps {
				cmpFn = func(a, b int) int {
					comparisons++
					return cmp.Compare(a, b)
				}
			}

			b.ReportAllocs()
			for range b.N {
				_ = slice.LISFunc(input, cmpFn)
			}

			if *lisCountCmps {
				perCmp := float64(b.Elapsed().Nanoseconds()) / float64(comparisons)
				b.ReportMetric(perCmp, "ns/cmp")
			}
		})
	}
}
