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
	cap int     // maximum allowed size of buf
	p   float64 // eviction probability
	rng rng
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
		p:   1,
		rng: rand.New(rand.NewChaCha8(seed)),
	}
}

// Len reports the number of elements currently buffered by c.
func (c *Counter[T]) Len() int { return c.buf.Len() }

// Reset resets c to its initial state, as if freshly constructed.
// The internal buffer size limit remains unchanged.
func (c *Counter[T]) Reset() { c.buf.Clear(); c.p = 1 }

// Add adds v to the counter.
func (c *Counter[T]) Add(v T) {
	if c.rng.Float64() >= c.p {
		c.buf.Remove(v)
		return
	}
	c.buf.Add(v)
	if c.buf.Len() >= c.cap {
		for elt := range c.buf {
			if c.rng.Float64() < 0.5 {
				c.buf.Remove(elt)
			}
		}
		c.p /= 2
	}
}

// Count returns the current estimate of the number of distinct elements
// observed by the counter.
func (c *Counter[T]) Count() int64 { return int64(float64(c.buf.Len()) / c.p) }

// BufferSize returns a buffer size sufficient to ensure that a counter using
// this size will produce estimates within (1 ± ε) times the true count with
// probability (1 - δ), assuming the expected total number of elements to be
// counted is expSize.
//
// The suggested buffer size guarantees these constraints, but note that the
// estimate is very conservative. In practice, the actual estimates will
// usually be much more accurate. Empirically, values of ε and δ in the 0.05
// range work well.
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

// rng is a source of uniformly-distributed random or pseudo-random values.
// This interface is satisfied by the [math/rand/v2.Rand] and [math/rand.Rand]
// types.
type rng interface {
	// Float64 returns a random value in the half-open interval [0,1).
	Float64() float64
}
