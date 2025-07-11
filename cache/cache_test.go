package cache_test

import (
	"testing"

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

	c := cache.New(cache.LRU[string, string](25).
		WithSize(cache.Length).

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
}

func TestSieve(t *testing.T) {
	var victims []string

	wantVic := func(t *testing.T, want ...string) {
		t.Helper()
		if diff := gocmp.Diff(victims, want); diff != "" {
			t.Errorf("Victims (-got, +want):\n%s", diff)
		}
	}

	c := cache.New(cache.Sieve[string, string](3).
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
		victims = nil
		cachetest.Run(t, c,
			"put k5 F = true",
			"size = 3", "len = 3",
			"has k2 = true", "has k3 = true", "has k5 = true",
		)
		wantVic(t, "k4")
	})

	t.Run("Remove", func(t *testing.T) {
		t.Skip()
		cachetest.Run(t, c, "remove k3 = true", "len = 2", "size = 20")
		wantVic(t, "k3")
	})

	t.Run("ReAdd", func(t *testing.T) {
		t.Skip()
		cachetest.Run(t, c, "put k3 stump = true", "len = 3", "size = 25")
	})

	t.Run("Clear", func(t *testing.T) {
		// Clearing evicts everything, which at this point are k2, k3, and k6 in
		// decreasing order of access time (the get of k2 above promoted it).
		victims = nil
		cachetest.Run(t, c, "clear", "len = 0", "size = 0")
		wantVic(t, "k2", "k3", "k5")
	})
}
