// Copyright (C) Michael J. Fromberger. All Rights Reserved.

package cache_test

import (
	"math/rand/v2"
	"sync"
	"testing"
	"testing/synctest"
	"time"

	"github.com/creachadair/mds/cache"
	"github.com/creachadair/mds/cache/internal/cachetest"
	gocmp "github.com/google/go-cmp/cmp"
)

func TestLRU(t *testing.T) {
	var victims []string

	wantVic := func(t *testing.T, want ...string) {
		t.Helper()
		if diff := gocmp.Diff(victims, want); diff != "" {
			t.Errorf("Victims (-got, +want):\n%s", diff)
		}
	}

	c := cache.New(cache.LRU[string, string]().
		WithLimit(25).
		WithSizeFunc(cache.Length).

		// Record evictions so we can verify they happened in the expected order.
		OnEvict(func(key, _ string) {
			victims = append(victims, key)
		}),
	)

	t.Run("New", func(t *testing.T) {
		cachetest.Run(t, c, "size = 0", "len = 0")
	})

	t.Run("Fill", func(t *testing.T) {
		cachetest.Run(t, c,
			"put k1 abcde12345 = true",
			"size = 10", "len = 1",
			"put k2 fghij67890 = true",
			"size = 20", "len = 2",
			"put k3 12345 = true",
		)
		wantVic(t)
	})

	t.Run("Evict", func(t *testing.T) {
		cachetest.Run(t, c,
			"put k4 67890 = true",
			"len = 3", "size = 20",
			"put k5 lmnop = true",
			"len = 4", "size = 25",
		)
		wantVic(t, "k1") // the eldest so far
	})

	t.Run("Check", func(t *testing.T) {
		cachetest.Run(t, c,
			"has k1 = false", // was evicted, see above
			"has k2 = true",
			"has k3 = true",
			"has k4 = true",
			"has k5 = true",
		)
	})

	t.Run("Access", func(t *testing.T) {
		cachetest.Run(t, c,
			"get k2 = fghij67890 true",
			"get k3 = 12345 true",
			"get k7 = '' false",

			// Now k4 is the least-recently accessed
		)
	})

	t.Run("EvictMore", func(t *testing.T) {
		victims = nil

		// Size is 25, we add +10. This requires us to evict 10, and the oldest
		// eligible are k4 (-5) and k5 (-5). Then we have 15, + 10 == 25 again.
		// We are left with k2, k3, and k6 (the one we just added).
		cachetest.Run(t, c,
			"put k6 appleberry = true",
			"size = 25", "len = 3",
			"has k2 = true", "has k3 = true", "has k6 = true",
		)
		wantVic(t, "k4", "k5")
	})

	t.Run("TooBig", func(t *testing.T) {
		victims = nil

		// This value is too big to be cached, make sure it is rejected and that
		// it does not throw anything else out -- even if it overlaps with an
		// existing key.
		cachetest.Run(t, c,
			"put k2 1aaaa2bbbb3cccc4ddde5eeee6ffff = false", // length 30 > 25
			"len = 3", "size = 25", // we didn't remove anything
			"get k2 = fghij67890 true", // we still have the old value for k2
		)
		wantVic(t)
	})

	t.Run("Remove", func(t *testing.T) {
		cachetest.Run(t, c, "remove k3 = true", "len = 2", "size = 20")
		wantVic(t, "k3")
	})

	t.Run("ReAdd", func(t *testing.T) {
		cachetest.Run(t, c, "put k3 stump = true", "len = 3", "size = 25")
	})

	t.Run("Clear", func(t *testing.T) {
		// Clearing evicts everything, which at this point are k6, k2, and k3 in
		// decreasing order of access time (the get of k2 above promoted it).
		victims = nil
		cachetest.Run(t, c, "clear", "len = 0", "size = 0")
		wantVic(t, "k6", "k2", "k3")
	})

	t.Run("Pop", func(t *testing.T) {
		victims = nil
		c := cache.New(cache.LRU[string, string]().
			WithLimit(10).
			OnEvict(func(key, _ string) { victims = append(victims, key) }))

		cachetest.Run(t, c,
			"put k1 abcde = true",
			"put k2 fghij = true",
			"put k3 klmno = true",
			"len = 3",
			"get k1 = abcde true",
			"pop = k2 fghij true",
			"pop = k3 klmno true",
			"pop = k1 abcde true",
			"pop = '' '' false",
			"len = 0",
		)
		wantVic(t, "k2", "k3", "k1")
	})
}

func TestSieve(t *testing.T) {
	var victims []string

	wantVic := func(t *testing.T, want ...string) {
		t.Helper()
		if diff := gocmp.Diff(victims, want); diff != "" {
			t.Errorf("Victims (-got, +want):\n%s", diff)
		}
	}

	c := cache.New(cache.Sieve[string, string]().
		WithLimit(3).

		// Record evictions so we can verify they happened in the expected order.
		OnEvict(func(key, _ string) {
			victims = append(victims, key)
		}),
	)

	t.Run("New", func(t *testing.T) {
		cachetest.Run(t, c, "size = 0", "len = 0")
	})

	t.Run("Fill", func(t *testing.T) {
		cachetest.Run(t, c,
			"put k1 A = true",
			"put k2 B = true",
			"put k3 C = true",
			"size = 3", "len = 3",
		)
		wantVic(t)
	})

	t.Run("Evict", func(t *testing.T) {
		cachetest.Run(t, c,
			"put k4 D = true",
			"len = 3", "size = 3",
		)
		wantVic(t, "k1") // the eldest so far
	})

	t.Run("Check", func(t *testing.T) {
		cachetest.Run(t, c,
			"has k1 = false", // was evicted, see above
			"has k2 = true",
			"has k3 = true",
			"has k4 = true",
		)
	})

	t.Run("Access", func(t *testing.T) {
		cachetest.Run(t, c,
			"get k2 = B true",
			"get k3 = C true",
			"get k6 = '' false",

			// Now k4 is the oldest unvisited entry.
		)
	})

	t.Run("EvictMore", func(t *testing.T) {
		cachetest.Run(t, c,
			"put k5 F = true",
			"size = 3", "len = 3",
			"has k2 = true", "has k3 = true", "has k5 = true",
		)
		wantVic(t, "k1", "k4")
	})

	t.Run("Remove", func(t *testing.T) {
		cachetest.Run(t, c,
			"remove k3 = true",
			"len = 2", "size = 2",
			"has k2 = true", "has k5 = true",
		)
		wantVic(t, "k1", "k4", "k3")
	})

	t.Run("ReAdd", func(t *testing.T) {
		cachetest.Run(t, c,
			"put k3 stump = true",
			"len = 3", "size = 3",
			"has k3 = true", "has k5 = true",
		)
		wantVic(t, "k1", "k4", "k3")
	})

	t.Run("Clear", func(t *testing.T) {
		// Clearing evicts everything, which at this point are k2, k5, and k3 in
		// decreasing order of visit time (the get of k2 above promoted it).
		cachetest.Run(t, c, "clear", "len = 0", "size = 0")
		wantVic(t, "k1", "k4", "k3", "k2", "k5", "k3")
	})
}

func TestExpiration(t *testing.T) {
	t.Run("Correctness", func(t *testing.T) {
		synctest.Test(t, func(t *testing.T) {
			var vmu sync.Mutex
			var victims []string
			wantVic := func(want ...string) {
				t.Helper()
				vmu.Lock()
				defer vmu.Unlock()
				if diff := gocmp.Diff(victims, want); diff != "" {
					t.Errorf("Evictions (-got, +want):\n%s", diff)
				}
			}

			c := cache.New(cache.LRU[string, string]().
				WithLimit(10).
				OnEvict(func(_, val string) {
					vmu.Lock()
					defer vmu.Unlock()
					victims = append(victims, val)
				}))

			t0 := time.Now()

			// An ordinary key that does not expire.
			cachetest.Run(t, c, "put a apple = true")

			// A key with an expiration assigned at birth to t0+5s.
			c.PutWithExpiration("b", "banana", t0.Add(5*time.Second))

			// More unexpiring keys.
			cachetest.Run(t, c,
				"put c cherry = true",
				"put d durian = true",
				"put e elderberry = true",
			)
			wantVic()

			// At t0+2s, add an expiration to c at t0+15s.
			time.Sleep(2 * time.Second)
			c.SetExpiration("c", time.Now().Add(15*time.Second))
			wantVic()

			// After t0+6s, "b" should have expired.
			time.Sleep(4 * time.Second)
			wantVic("banana")

			// After t0+26s, "c" should have expired.
			time.Sleep(20 * time.Second)
			wantVic("banana", "cherry")

			// Add an expiration for "a", then remove it before t fires
			// Verify that "a" did not expire.
			if !c.SetExpiration("a", t0.Add(5*time.Minute)) {
				t.Error("Key a should have been present")
			}
			time.Sleep(time.Second)
			if !c.SetExpiration("a", time.Time{}) {
				t.Error("Key a should have been present")
			}

			time.Sleep(10 * time.Minute)
			wantVic("banana", "cherry")

			// Proposing a new key with an expiration in the past does not add anything.
			if c.PutWithExpiration("q", "quince", time.Now().Add(-time.Hour)) {
				t.Error("Key q should not have been added")
			}

			// Setting an expiration in the past evicts the key (at the next sync point).
			c.SetExpiration("a", time.Now().Add(-time.Second))

			synctest.Wait()
			wantVic("banana", "cherry", "apple")

			// Update "d" with a new value and expiration (in the future).
			// This generates an eviction for "durian".
			if !c.PutWithExpiration("d", "dragonfruit", time.Now().Add(time.Minute)) {
				t.Error("Key d should have been updated")
			}
			wantVic("banana", "cherry", "apple", "durian")

			// Update "d" with a new value and no expiration before that expires.
			// This generates an eviction for "dragonfruit".
			time.Sleep(time.Second)
			if !c.Put("d", "dingleberry") {
				t.Error("Key d should have been updated")
			}

			synctest.Wait()
			wantVic("banana", "cherry", "apple", "durian", "dragonfruit")

			// More time passes, but "d" should remain.
			time.Sleep(5 * time.Minute)
			wantVic("banana", "cherry", "apple", "durian", "dragonfruit") // i.e., not dingleberry

			if !c.Has("d") {
				t.Error("Key d should be present")
			}

			// If we remove a key (e) that has an expiration, the expiration should
			// go with it.  Adding the same key back won't inherit the expiration.
			c.SetExpiration("e", time.Now().Add(10*time.Minute))
			time.Sleep(5 * time.Minute)
			cachetest.Run(t, c, "remove e = true", "put e evilfruit = true")
			wantVic("banana", "cherry", "apple", "durian", "dragonfruit", "elderberry")

			time.Sleep(20 * time.Minute)
			if !c.Has("e") {
				t.Error("Key e should be present")
			}
			wantVic("banana", "cherry", "apple", "durian", "dragonfruit", "elderberry")

			// A key not present should not affect state.
			if c.SetExpiration("q", time.Now().Add(time.Hour)) {
				t.Error("Key q should not be present")
			}
		})
	})

	t.Run("Cleanup", func(t *testing.T) {
		synctest.Test(t, func(t *testing.T) {
			c := cache.New(cache.Sieve[int, string]().WithLimit(10))

			// Add some pending expirations, to exercise that calling Clear releases them.
			// If it does not, the synctest bubble will fail at exit because the maintenance
			// goroutine will still be blocked.

			c.PutWithExpiration(5, "apple", time.Now().Add(1*time.Minute))
			c.PutWithExpiration(4, "pear", time.Now().Add(2*time.Minute))
			c.PutWithExpiration(6, "cherry", time.Now().Add(3*time.Minute))

			time.Sleep(90 * time.Second)

			c.Clear()
			if n := c.Len(); n != 0 {
				t.Errorf("After clear: got %d items, want 0", n)
			}
		})
	})

	t.Run("Rolling", func(t *testing.T) {
		synctest.Test(t, func(t *testing.T) {
			const maxEntries = 100
			const numKeys = 2_000

			c := cache.New(cache.LRU[int, int]().WithLimit(maxEntries))

			// Push a bunch of keys through the cache, some with and some without
			// expirations, to give the race detector something to work with.
			var cur int
			var elts int
			for i := range numKeys {
				cur++
				if rand.Float64() <= 0.6 {
					c.PutWithExpiration(i%(2*maxEntries), cur, time.Now().Add(rand.N(5*time.Second)))
				} else {
					c.Put(i%200, cur)
				}
				elts = max(elts, c.Len())
				time.Sleep(rand.N(800 * time.Millisecond))
			}
			t.Logf("Done priming cache, %d items present (max %d)", c.Len(), elts)

			// Sleep long enough that all expiring keys should go away.
			time.Sleep(time.Hour)
			after := c.Len()
			t.Logf("After expiration is complete, %d items remain", after)

			// Verify that no more items disappear after a longer wait.
			time.Sleep(time.Hour)
			if got := c.Len(); got != after {
				t.Errorf("Expired count changed: got %d, want %d", got, after)
			}

			c.Clear()
			if n := c.Len(); n != 0 {
				t.Errorf("After clear: have %d elements, want 0", n)
			}
		})
	})
}
