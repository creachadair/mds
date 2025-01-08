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
	for newSize > c.limit {
		ek, ev := c.store.Evict()
		c.onEvict(ek, ev)
		c.count--
		newSize -= c.sizeOf(ev)
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

	for c.count > 0 {
		ek, ev := c.store.Evict()
		c.onEvict(ek, ev)
		c.size -= c.sizeOf(ev)
		c.count--
	}
	if c.size != 0 || c.count != 0 {
		panic(fmt.Sprintf("cache: after clear size=%d count=%d", c.size, c.count))
	}
}

// Size reports the current size of the items in c.
func (c *Cache[K, V]) Size() int64 {
	c.μ.Lock()
	defer c.μ.Unlock()
	return c.size
}

// New constructs a new empty cache with the specified settings.
// The store and capacity limits of config must be set or New will panic.
func New[K comparable, V any](config Config[K, V]) *Cache[K, V] {
	if config.limit <= 0 {
		panic("cache: limit must be positive")
	}
	if config.store == nil {
		panic("cache: no store implementation")
	}
	return &Cache[K, V]{
		store:   config.store,
		limit:   config.limit,
		sizeOf:  config.sizeFunc(),
		onEvict: config.onEvictFunc(),
	}
}

// A Config carries the settings for a cache implementation.
// To set options:
//
//   - Use [Config.WithLimit] to set the capacity.
//   - Use [Config.WithStore] to set the storage implementation.
//   - Use [Config.WithSize] to set the size function.
//   - Use [Config.OnEvict] to set the eviction callback.
//
// A zero Config is invalid; at least the store field must be set.
type Config[Key comparable, Value any] struct {
	// limit is the capacity limit for the cache.
	// It must be positive. The interpretation depends on sizeOf.
	limit int64

	// store is the storage implementation used by the cache.
	// It must be non-nil.
	store Store[Key, Value]

	// sizeOf reports the effective size of v in the cache. If nil, the default
	// size is 1, meaning the limit is a number of cache entries.
	sizeOf func(v Value) int64

	// onEvict, if non-nil, is called for each entry evicted from the cache.
	onEvict func(key Key, val Value)
}

// WithLimit returns a copy of c with its capacity set to n.
// The limit implementation must be positive, or [New] will panic.
func (c Config[K, V]) WithLimit(n int64) Config[K, V] { c.limit = n; return c }

// WithStore returns a copy of c with its storage implementation set to s.
// The storage implementation must be set, or [New] will panic.
func (c Config[K, V]) WithStore(s Store[K, V]) Config[K, V] { c.store = s; return c }

// WithSize returns a copy of c with its size function set to sizeOf.
//
// If no size function is set, the default size of an entry is 1, meaning the
// limit is based on the number of entries in the cache.
func (c Config[K, V]) WithSize(sizeOf func(V) int64) Config[K, V] { c.sizeOf = sizeOf; return c }

// OnEvict returns a copy of c with its eviction callback set to f.
//
// If an eviction callback is set, it is called for each entry removed or
// evicted from the cache.
func (c Config[K, V]) OnEvict(f func(K, V)) Config[K, V] { c.onEvict = f; return c }

func (c Config[K, V]) sizeFunc() func(V) int64 {
	if c.sizeOf != nil {
		return c.sizeOf

		// TODO(creachadair): Maybe defensively take max(_, 1)?
	}
	return func(V) int64 { return 1 }
}

func (c Config[K, V]) onEvictFunc() func(K, V) {
	if c.onEvict != nil {
		return c.onEvict
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
	// That condition should not be possible when used from a Cache.
	Store(key Key, val Value)

	// Remove removes the specified key from the cache.  If key is not present,
	// Remove should do nothing.
	Remove(key Key)

	// Evict evicts an entry from the cache, chosen by the Store, and returns
	// the key and value evicted.
	//
	// If there are no items in the store, it should panic.
	// That condition should not be possible when used from a Cache.
	Evict() (Key, Value)
}

// Length is a convenience function for using the length of a string or byte
// slice as its size in a cache. It returns len(v).
func Length[T ~[]byte | ~string](v T) int64 { return int64(len(v)) }
