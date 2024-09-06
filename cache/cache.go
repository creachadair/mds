// Package cache implements a keyed cache for arbitrary values.
package cache

import (
	"fmt"
	"sync"
)

// A Cache is a cache mapping keys to values, with a fixed limit on its maximum
// capacity. Any key may be present in the cache at most once. By default,
// cache capacity is a number of elements; however, the caller may specify a
// different size metric using the [Config] argument to [New].
//
// A Cache is safe for concurrent access by multiple goroutines.
type Cache[Key comparable, Value any] struct {
	// Hold μ to call any method of store and to read or set size/limit.
	μ           sync.Mutex
	store       Store[Key, Value]
	size, limit int64
	count       int

	// Set once at construction, read-only thereafter.
	sizeOf  func(Value) int64
	onEvict func(Key, Value)

	// TODO(creachadair): add metrics
}

// Has reports whether a value for key is present in c.  This does not count as
// an access of the value for cache accounting.
func (c *Cache[K, _]) Has(key K) bool {
	c.μ.Lock()
	defer c.μ.Unlock()
	_, ok := c.store.Check(key)
	return ok
}

// Get reports whether key is present in c, and if so returns the corresponding
// cached value. This counts as an access of the value for cache accounting.
func (c *Cache[K, V]) Get(key K) (V, bool) {
	c.μ.Lock()
	defer c.μ.Unlock()
	return c.store.Access(key)
}

// Put adds or replaces the value for key in c, and reports whether the value
// was successfully stored. Put reports false if the cache does not have room
// to store the provided value; otherwise, the cache is updated and Put reports
// true. If necessary, items are evicted from the cache to make room for the
// new value. Which values are evicted is determined by the cache store.
func (c *Cache[K, V]) Put(key K, val V) bool {
	c.μ.Lock()
	defer c.μ.Unlock()

	valSize := c.sizeOf(val)
	if valSize > c.limit {
		return false // this value will never fit
	}

	// If there is an existing item for this key, remove it.
	if old, ok := c.store.Check(key); ok {
		c.store.Remove(key)
		c.onEvict(key, old)
		c.size -= c.sizeOf(old)
		c.count--
	}

	// If necessary, evict items to make room.
	newSize := c.size + valSize
	if newSize > c.limit {
		want := int64(newSize - c.limit)
		got := c.store.Evict(want, func(ek K, ev V) int64 {
			c.onEvict(ek, ev)
			c.count--
			return c.sizeOf(ev)
		})
		if got < want {
			panic(fmt.Sprintf("store: evicted %d units, need %d", got, want))
		}
		newSize -= got
	}

	// Now there is room.
	c.store.Store(key, val)
	c.size = newSize
	c.count++
	return true
}

// Remove removes the specified key from c, and reports whether a value had
// been cached for that key.
func (c *Cache[K, _]) Remove(key K) bool {
	c.μ.Lock()
	defer c.μ.Unlock()

	if old, ok := c.store.Check(key); ok {
		c.store.Remove(key)
		c.onEvict(key, old)
		c.size -= c.sizeOf(old)
		c.count--
		return true
	}
	return false
}

// Len reports the number of items present in the cache.
func (c *Cache[K, V]) Len() int {
	c.μ.Lock()
	defer c.μ.Unlock()
	return c.count
}

// Clear discards the complete contents of c, leaving it empty.
func (c *Cache[K, V]) Clear() {
	c.μ.Lock()
	defer c.μ.Unlock()

	if got := c.store.Evict(c.size, func(ek K, ev V) int64 {
		c.onEvict(ek, ev)
		return c.sizeOf(ev)
	}); got != c.size {
		panic(fmt.Sprintf("store: evicted %d units, need %d", got, c.size))
	}
	c.size = 0
	c.count = 0
}

// Size reports the current size of the items in c.
func (c *Cache[K, V]) Size() int64 {
	c.μ.Lock()
	defer c.μ.Unlock()
	return c.size
}

// New constructs a new empty cache with the specified settings.  The store
// must be non-nil, and limit must be positive.  A nil [Config] is valid, and
// provides default values as described.
func New[K comparable, V any](limit int64, store Store[K, V], config *Config[K, V]) *Cache[K, V] {
	if limit <= 0 {
		panic("cache: limit must be positive")
	}
	if store == nil {
		panic("cache: no store implementation")
	}
	return &Cache[K, V]{
		store:   store,
		limit:   limit,
		sizeOf:  config.sizeOf(),
		onEvict: config.onEvict(),
	}
}

// A Config carries the settings for a cache implementation.  A nil *Config is
// ready for use and provides default values as described.
type Config[Key comparable, Value any] struct {
	// Size reports the effective size of v in the cache. If nil, or if the
	// function returns a value ≤ 0, the default is 1.
	Size func(v Value) int64

	// OnEvict, if non-nil, is called for each entry evicted from the cache.
	OnEvict func(key Key, val Value)
}

func (c *Config[K, V]) sizeOf() func(V) int64 {
	if c != nil && c.Size != nil {
		return c.Size
	}
	return func(V) int64 { return 1 }
}

func (c *Config[K, V]) onEvict() func(K, V) {
	if c != nil && c.OnEvict != nil {
		return c.OnEvict
	}
	return func(K, V) {}
}

// Store is the interface to a cache storage backend. A Store determines the
// cache eviction policy.
//
// A Cache will serialize access to the methods of Store, so it is not
// necessary for the implementation to do so separately, unless it is to be
// shared among multiple cache instances.
type Store[Key comparable, Value any] interface {
	// Access reports whether key is present, and if so returns its
	// corresponding value and records an access of the value.
	Access(key Key) (Value, bool)

	// Check reports whether key is present and, if so, returns the
	// corresponding value without recording an access.
	Check(key Key) (Value, bool)

	// Store adds the specified key, value entry to the cache.
	// This counts as an access of the value.
	//
	// If key is already present, Store should panic.
	// That condition should not be possible when used from a [Cache].
	Store(key Key, val Value)

	// Remove removes the specified key from the cache.  If key is not present,
	// Remove should do nothing.
	Remove(key Key)

	// Evict evicts entries from the cache whose total size is at least need.
	// It reports the actual size of the evicted items.  It must call sizeOf on
	// each key/value pair to be evicted, if any, to obtain its effective size.
	//
	// If Evict cannot satisfy the specified size, it should panic.
	// That condition should not be possible when used from a [Cache].
	Evict(need int64, sizeOf func(Key, Value) int64) int64
}

// Length is a convenience function for using the length of a string or byte
// slice as its size in a cache. It returns len(v).
func Length[T ~[]byte | ~string](v T) int64 { return int64(len(v)) }
