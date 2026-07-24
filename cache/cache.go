// Copyright (C) Michael J. Fromberger. All Rights Reserved.

// Package cache implements a keyed cache for arbitrary values.
//
// # Overview
//
// Call [New] to construct a new, empty [Cache] from a [Config]. A minimal
// config specifies the storage implemetation ([Store]) and a size limit for
// the cache. For example, to create an LRU cache for int keys and string
// values, with 10 slots, write:
//
//	c := cache.New(cache.LRU[int, string]().WithLimit(10))
//
// By default the limit is a number of entries in the cache. To use a different
// value size measurement, call [Config.WithSizeFunc]. For example, to use the
// length of the value string as its size, and to set the size limit to 1MiB,
// write:
//
//	cfg := cache.LRU[int, string]().
//	    WithLimit(1 << 20).
//	    WithSizeFunc(cache.Length)
//
// The storage implementation determines the eviction policy. This package
// provides [LRU] and [Sieve] implementations; you may also plug in your own
// implementation with a different policy using [Config.WithStore].
// A [Cache] is safe for concurrent use by multiple goroutines.
//
// ## Key Methods
//
// Add keys to the cache with [Cache.Put] or [Cache.PutWithExpiration].
// Remove keys from the cache with [Cache.Remove] or [Cache.Clear].
// Check whether a key is present with [Cache.Has]. This does not count as an
// access of the key for accounting purposes.
// Access the value of a key with [Cache.Get]. This counts as an access.
// Use [Cache.Len] to report the number of entries present in the cache.
// Use [Cache.Size] to report the total size of all entries in the cache.
//
// # Eviction
//
// By default, entries are evicted silently from the cache as needed to make
// room for new entries. If you need to know when entries are evicted, use
// [Config.OnEvict] to set a callback that will be executed synchronously for
// each key/value pair evicted or removed from the cache. Note that replacing a
// key with [Cache.Put] also counts as an eviction for the old value being
// replaced.
//
// You can explicitly evict the most-eligible entry using [Cache.Pop].
//
// # Automatic Expiration
//
// The caller may request that a key should expire and be automatically removed
// from the cache at or after a certain [time.Time].  Entries added by
// [Cache.Put] do not expire. Use [Cache.PutWithExpiration] to specify that the
// key being added should be automatically removed after the given time. Use
// [Cache.SetExpiration] to add or update the expiration time of an existing
// key in the cache.
//
// Expirations are processed promptly, but are not guaranteed to occur at
// exactly the time specified.
//
// While any expirations are pending, a cache maintains a background goroutine
// to remove expired keys. The goroutine exits automatically when no further
// expirations are needed.  When you are finished using a cache that has
// expiration times set, you should call [Cache.Clear] to discard all the keys
// in the cache, so that this goroutine can be cleaned up. If you do not set
// any expiration times, this is not necessary.
package cache

import (
	"fmt"
	"sync"
	"time"

	"github.com/creachadair/mds/cache/internal/expirer"
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
	expirer *expirer.Expirer[Key]

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
// true. This counts as an access of the value for accounting purposes.
//
// If necessary, items are evicted from the cache to make room for the new
// value. Which values are evicted is determined by the cache store. If Put
// replaces an existing value for key, the old value is also evicted.
//
// A key added or updated by Put does not expire. If Put replaces an existing
// key that had an expiration time, that expiration is removed.
// See [Cache.PutWithExpiration] and [Cache.SetExpiration].
func (c *Cache[K, V]) Put(key K, val V) bool {
	c.μ.Lock()
	defer c.μ.Unlock()
	return c.putLocked(key, val)
}

// PutWithExpiration behaves as [Cache.Put], and also sets the expiration time
// for the specified key.
//
// If the specified key is already present in the cache, its value and deadline
// are updated. If deadline is the zero time, an existing deadline (if any) is
// removed. This counts as an access of the
//
// If the specified key is not already present in the cache, and deadline is in
// the past (including the zero time), PutWithExpiration reports false without
// modifying the cache.
func (c *Cache[K, V]) PutWithExpiration(key K, val V, deadline time.Time) bool {
	c.μ.Lock()
	defer c.μ.Unlock()
	if _, ok := c.store.Check(key); !ok && deadline.Before(time.Now()) {
		return false
	}
	if c.putLocked(key, val) {
		c.setExpirationLocked(key, deadline)
		return true
	}
	return false
}

// SetExpiration sets or updates the expiration deadline for the specified key.
// If deadline is the zero time, any pending expiration for key is removed, and
// the key will not expire.  It reports whether the specified key is present in
// c. This does not count as an access of the value for accounting purposes,
// even if it modifies the expiration time for key.
//
// It is legal for deadline to be a non-zero time in the past, in which case
// the key will be removed at the next available opportunity. However, if you
// wish to remove a key immediately, prefer [Cache.Remove].
func (c *Cache[K, V]) SetExpiration(key K, deadline time.Time) bool {
	c.μ.Lock()
	defer c.μ.Unlock()
	if _, ok := c.store.Check(key); !ok {
		return false
	}
	c.setExpirationLocked(key, deadline)
	return true
}

// putLocked is the shared implementation of Put and PutWithExpiration.
// Precondition: Caller holds c.μ.
func (c *Cache[K, V]) putLocked(key K, val V) bool {
	valSize := c.sizeOf(val)
	if valSize > c.limit {
		return false // this value will never fit
	}

	// If there is an existing item for this key, remove it.
	if old, ok := c.store.Check(key); ok {
		c.store.Remove(key)
		c.dropKeyLocked(key, old)
	}

	// If necessary, evict items to make room.
	for c.size+valSize > c.limit {
		c.dropKeyLocked(c.store.Evict())
	}

	// Now there is room.
	c.store.Store(key, val)
	c.size += valSize
	c.count++
	return true
}

// Remove removes the specified key from c, and reports whether a value had
// been cached for that key. If Remove reports true, the removed value is
// reported as an eviction.
func (c *Cache[K, _]) Remove(key K) bool {
	c.μ.Lock()
	defer c.μ.Unlock()

	if old, ok := c.store.Check(key); ok {
		c.store.Remove(key)
		c.dropKeyLocked(key, old)
		return true
	}
	return false
}

func (c *Cache[K, _]) removeExpired(key K) {
	c.μ.Lock()
	defer c.μ.Unlock()

	if old, ok := c.store.Check(key); ok {
		c.store.Remove(key)
		c.dropKeyLockedInternal(key, old)
	}
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

	c.expirer.ClearExpirations()
	for c.count > 0 {
		c.dropKeyLocked(c.store.Evict())
	}
	if c.size != 0 || c.count != 0 {
		panic(fmt.Sprintf("cache: after clear size=%d count=%d", c.size, c.count))
	}
}

// Pop evicts the most eligible eviction candidate from c and reports whether
// anything was removed. If Pop reports false, it means c was empty.
func (c *Cache[K, V]) Pop() (_ K, _ V, ok bool) {
	c.μ.Lock()
	defer c.μ.Unlock()

	if c.count == 0 {
		return
	}
	k, v := c.store.Evict()
	c.dropKeyLocked(k, v)
	return k, v, true
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

// dropKeyLocked reports the removal of key and updates the cache size.
// It also discards any pending expiration for the key.
// Precondition: Caller holds c.μ.
func (c *Cache[K, V]) dropKeyLocked(key K, value V) {
	c.expirer.RemoveExpiration(key)
	c.dropKeyLockedInternal(key, value)
}

func (c *Cache[K, V]) dropKeyLockedInternal(key K, value V) {
	c.onEvict(key, value)
	c.size -= c.sizeOf(value)
	c.count--
}

// setExpirationLocked sets or updates the expiration time of key to deadline.
// If deadline is the zero time, any expiration for key is removed.
// Precondition: Caller holds c.μ.
func (c *Cache[K, _]) setExpirationLocked(key K, deadline time.Time) {
	if deadline.IsZero() {
		c.expirer.RemoveExpiration(key)
		return
	}
	if c.expirer == nil {
		c.expirer = expirer.New[K](c.removeExpired) // lazy init
	}
	c.expirer.SetExpiration(key, deadline)
}

// A Config carries the settings for a cache implementation.
// To set options:
//
//   - Use [Config.WithLimit] to set the capacity.
//   - Use [Config.WithStore] to set the storage implementation.
//   - Use [Config.WithSizeFunc] to set the size function.
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
// Deprecated: Use [Config.WithSizeFunc] instead.
func (c Config[K, V]) WithSize(sizeOf func(V) int64) Config[K, V] { return c.WithSizeFunc(sizeOf) }

// WithSizeFunc returns a copy of c with its size function set to sizeOf.
//
// If no size function is set, or if sizeOf == nil,the default size of an entry
// is 1, meaning the limit is based on the number of entries in the cache.
func (c Config[K, V]) WithSizeFunc(sizeOf func(V) int64) Config[K, V] { c.sizeOf = sizeOf; return c }

// OnEvict returns a copy of c with its eviction callback set to f.
//
// If an eviction callback is set, it is called for each entry removed or
// evicted from the cache. Note that updating the entry for a given key also
// records an eviction for the old value being replaced.
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
// necessary for the implementation to do so separately.
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
