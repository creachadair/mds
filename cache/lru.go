package cache

import (
	"cmp"
	"fmt"

	"github.com/creachadair/mds/heapq"
)

// lruStore is an implementation of the [Store] interface.
// Eviction chooses the least-recently accessed elements first.
type lruStore[Key comparable, Value any] struct {
	present map[Key]int // :: Key → offset in access
	access  *heapq.Queue[prioKey[Key, Value]]
	clock   int64

	// A linked list is asymptotically better than a heap, but the heap avoids
	// all the pointer indirections, allocates less, and leaves less garbage.
}

type prioKey[Key comparable, Value any] struct {
	lastAccess int64
	key        Key
	value      Value
}

func comparePrio[Key comparable, Value any](a, b prioKey[Key, Value]) int {
	return cmp.Compare(a.lastAccess, b.lastAccess) // logical time order
}

// LRU constructs a [Config] with a cache store with the specified capacity
// limit that manages entries with a least-recently used eviction policy.
func LRU[Key comparable, Value any](limit int64) Config[Key, Value] {
	lru := &lruStore[Key, Value]{
		present: make(map[Key]int),
		access:  heapq.New(comparePrio[Key, Value]),
	}
	lru.access.Update(func(v prioKey[Key, Value], pos int) {
		lru.present[v.key] = pos
	})
	return Config[Key, Value]{limit: limit, store: lru}
}

// Check implements part of the [Store] interface.
func (c *lruStore[Key, Value]) Check(key Key) (Value, bool) {
	pos, ok := c.present[key]
	if !ok {
		var zero Value
		return zero, false
	}
	elt, ok := c.access.Peek(pos)
	return elt.value, ok
}

// Access implements part of the [Store] interface.
func (c *lruStore[Key, Value]) Access(key Key) (Value, bool) {
	pos, ok := c.present[key]
	if !ok {
		var zero Value
		return zero, false
	}
	c.clock++ // this counts as an access

	// Remove the item at its existing priority, and re-add it as the most
	// recent access. Only the timestamp matters for order.
	out, _ := c.access.Remove(pos) // cannot fail
	out.lastAccess = c.clock
	c.access.Add(out)
	return out.value, true
}

// Store implements part of the [Store] interface.
func (c *lruStore[Key, Value]) Store(key Key, val Value) {
	if _, ok := c.present[key]; ok {
		panic(fmt.Sprintf("lru store: unexpected key %v", key))
	}

	c.clock++
	pos := c.access.Add(prioKey[Key, Value]{
		lastAccess: c.clock,
		key:        key,
		value:      val,
	})
	c.present[key] = pos
}

// Remove implements part of the [Store] interface.
func (c *lruStore[Key, _]) Remove(key Key) {
	pos, ok := c.present[key]
	if ok {
		c.access.Remove(pos)
		delete(c.present, key)
	}
}

// Evict implements part of the [Store] interface.
func (c *lruStore[Key, Value]) Evict() (Key, Value) {
	out, ok := c.access.Pop()
	if !ok {
		panic("lru evict: no entries left")
	}
	delete(c.present, out.key)
	return out.key, out.value
}
