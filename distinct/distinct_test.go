package distinct_test

import (
	"flag"
	"fmt"
	"math"
	"testing"

	"math/rand/v2"

	"github.com/creachadair/mds/distinct"
	"github.com/creachadair/mds/mapset"
)

var (
	errRate  = flag.Float64("error-rate", 0.06, "Error rate")
	failProb = flag.Float64("fail-probability", 0.02, "Failure probability")
)

func fill(c *distinct.Counter[int], n int) mapset.Set[int] {
	actual := mapset.New[int]()
	for range n {
		r := rand.Int()
		actual.Add(r)
		c.Add(r)
	}
	return actual
}

func TestCounter(t *testing.T) {
	t.Run("Empty", func(t *testing.T) {
		// An empty counter should report no elements.
		c := distinct.NewCounter[int](100)
		if got := c.Count(); got != 0 {
			t.Errorf("Empty count: got %d, want 0", got)
		}
	})

	t.Run("Small", func(t *testing.T) {
		// A counter that has seen fewer values than its buffer size should count
		// perfectly.
		c := distinct.NewCounter[int](100)
		want := len(fill(c, 50))
		if got := c.Len(); got != want {
			t.Errorf("Small count: got %d, want %d", got, want)
		}
	})

	t.Logf("Error rate: %g%%", 100**errRate)
	t.Logf("Failure probability: %g%%", 100**failProb)
	for _, tc := range []int{9_999, 100_000, 543_210, 1_000_000, 1_048_576} {
		name := fmt.Sprintf("Large/%d", tc)
		t.Run(name, func(t *testing.T) {
			size := distinct.BufferSize(*errRate, *failProb, tc)
			t.Logf("Buffer size estimate: %d", size)

			c := distinct.NewCounter[int](size)
			actual := fill(c, tc)

			t.Logf("Actual count:    %d", actual.Len())
			t.Logf("Estimated count: %d", c.Count())
			t.Logf("Buffer size:     %d", c.Len())

			e := float64(c.Count()-int64(actual.Len())) / float64(actual.Len())
			t.Logf("Error:           %.4g%%", 100*e)

			if math.Abs(e) > *errRate {
				t.Errorf("Error rate = %f, want ≤ %f", e, *errRate)
			}
			if c.Len() > size {
				t.Errorf("Buffer size is %d > %d", c.Len(), size)
			}

			// After counting, a reset should leave the buffer empty.
			c.Reset()
			if got := c.Len(); got != 0 {
				t.Errorf("After reset: buffer size is %d, want 0", got)
			}
		})
	}

	t.Run("Saturate", func(t *testing.T) {
		// To achieve a ± 5% error rate for 1M inputs, we theoretically need
		// about 142K buffer slots. With 10K buffer slots the expected error rate
		// for 1M inputs is about ± 18.8%. The predicted bound is correct, but is
		// very conservative: With high probability the error will be much less,
		// even when we greatly exceed the predicted load.
		//
		// In several hundred runs with random inputs, this configuration did not
		// exceed 5% error, although it could have done so.

		c := distinct.NewCounter[int](10_000)
		var actual mapset.Set[int]
		var maxErr float64
		for i := 0; i < 1_000_000; i += 500 {
			actual.AddAll(fill(c, 500))
			e := float64(c.Count()-int64(actual.Len())) / float64(actual.Len())
			if math.Abs(e) > math.Abs(maxErr) {
				maxErr = e
				t.Logf("At %d unique items, max error is %.4g%%", actual.Len(), 100*maxErr)
			}
		}
		t.Logf("Actual count:    %d", actual.Len())
		t.Logf("Estimated count: %d", c.Count())
		t.Logf("Buffer size:     %d", c.Len())
		t.Logf("Max error:       %.4g%%", 100*maxErr)
	})
}
