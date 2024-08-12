package distinct_test

import (
	"fmt"
	"math/rand/v2"
	"os"

	"github.com/creachadair/mds/distinct"
	"github.com/creachadair/mds/mapset"
)

func Example() {
	// Suggest how big a buffer we need to have an estimate within ± 10% of the
	// true value with 95% probability, given we expect to see 60000 inputs.
	bufSize := distinct.BufferSize(0.1, 0.05, 60000)

	// Construct a counter with the specified buffer size limit.
	c := distinct.NewCounter[int](bufSize)

	// For demonstration purposes, keep track of the actual count.
	// This will generally be impractical for "real" workloads.
	var unique mapset.Set[int]

	// Observe some (50,000) random inputs...
	for range 50000 {
		r := rand.IntN(80000)
		c.Add(r)

		unique.Add(r)
	}

	fmt.Printf("Buffer limit: %d\n", bufSize)
	fmt.Fprintf(os.Stderr, "Unique:       %d\n", unique.Len())
	fmt.Fprintf(os.Stderr, "Estimate:     %d\n", c.Count())
	fmt.Fprintf(os.Stderr, "Buffer used:  %d\n", c.Len())

	// N.B.: Counter results are intentionally omitted here. The exact values
	// are not stable even if the RNG is fixed, because the counter uses map
	// iteration during update.

	// Output:
	// Buffer limit: 27834
}
