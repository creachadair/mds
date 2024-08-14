// Package distinct implements the probabilistic distinct-elements counter
// algorithm of Chakraborty, Vinodchandran, and Meel as described in the paper
// "Distinct Elements in Streams" ([CVM]).
//
// [CVM]: https://arxiv.org/pdf/2301.10191
package distinct

import (
	crand "crypto/rand"
	"fmt"
	"math"
	"math/bits"
	"math/rand/v2"

	"github.com/creachadair/mds/mapset"
)

// A Counter estimates the number of distinct comparable elements that have
// been passed to its Add method using the CVM algorithm.
//
// Add elements to a counter using [Counter.Add] method; use [Counter.Count] to
// obtain the current estimate of the number of distinct elements observed.
type Counter[T comparable] struct {
	buf mapset.Set[T]
	cap int    // maximum allowed size of buf
	p   uint64 // eviction probability (see below)
	rng rand.Source

	// To avoid the need for floating-point calculations during update, we
	// express the probability as a fixed-point threshold in 0..MaxUint64, where
	// 0 denotes probability 0 and ~0 denotes probability 1.
}

// NewCounter constructs a new empty distinct-elements counter using a buffer
// of at most size elements for estimation.
//
// A newly-constructed counter does not pre-allocate the full buffer size.  It
// begins with a small buffer that grows as needed up to the limit.
func NewCounter[T comparable](size int) *Counter[T] {
	var seed [32]byte
	if _, err := crand.Read(seed[:]); err != nil {
		panic(fmt.Sprintf("seed RNG: %v", err))
	}
	return &Counter[T]{
		buf: make(mapset.Set[T]),
		cap: size,
		p:   math.MaxUint64,
		rng: rand.NewChaCha8(seed),
	}
}

// Len reports the number of elements currently buffered by c.
func (c *Counter[T]) Len() int { return c.buf.Len() }

// Reset resets c to its initial state, as if freshly constructed.
// The internal buffer size limit remains unchanged.
func (c *Counter[T]) Reset() { c.buf.Clear(); c.p = math.MaxUint64 }

// Add adds v to the counter.
func (c *Counter[T]) Add(v T) {
	if c.p < math.MaxUint64 && c.rng.Uint64() >= c.p {
		c.buf.Remove(v)
		return
	}
	c.buf.Add(v)
	if c.buf.Len() >= c.cap {
		// Instead of flipping a coin for each element, grab blocks of 64 random
		// bits and use them directly, refilling only as needed.
		var nb, rnd uint64

		for elt := range c.buf {
			if nb == 0 {
				rnd = c.rng.Uint64() // refill
				nb = 64
			}
			if rnd&1 == 0 {
				c.buf.Remove(elt)
			}
			rnd >>= 1
			nb--
		}
		c.p >>= 1
	}
}

// Count returns the current estimate of the number of distinct elements
// observed by the counter.
func (c *Counter[T]) Count() uint64 {
	// The estimate is |X| / p, where p = 1/2^k after k eviction passes.
	// To convert our fixed-point probability, note that:
	//
	//   |X| / p == |X| * (1/p) == |X| * 2^k
	//
	// The number of leading zeroes of c.p records k, so we can do this all in
	// fixed-point arithmetic with no floating point conversion.
	p2k := uint64(1) << uint64(bits.LeadingZeros64(c.p))
	return uint64(c.buf.Len()) * p2k
}

// BufferSize returns a buffer size sufficient to ensure that a counter using
// this size will produce estimates within (1 ± ε) times the true count with
// probability (1 - δ), assuming the expected total number of elements to be
// counted is expSize.
//
// The suggested buffer size guarantees these constraints, but note that the
// Chernoff bound estimate is very conservative. In practice, the actual
// estimates will usually be much more accurate. Empirically, values of ε and δ
// in the 0.05 range work well.
func BufferSize(ε, δ float64, expSize int) int {
	if ε < 0 || ε > 1 {
		panic(fmt.Sprintf("error bound out of range: %v", ε))
	}
	if δ < 0 || δ > 1 {
		panic(fmt.Sprintf("error rate out of range: %v", δ))
	}
	if expSize <= 0 {
		panic(fmt.Sprintf("expected size must be positive: %d", expSize))
	}

	v := math.Ceil((12 / (ε * ε)) * math.Log2((8*float64(expSize))/δ))
	return int(v)
}
