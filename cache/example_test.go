package cache_test

import (
	"fmt"

	"github.com/creachadair/mds/cache"
)

func Example() {
	c := cache.New(10, cache.LRU[string, int]())
	for i := range 50 {
		c.Put(fmt.Sprint(i+1), i+1)
	}

	fmt.Println("size:", c.Size())

	fmt.Println("has 1:", c.Has("1"))
	fmt.Println("has 40:", c.Has("40"))
	fmt.Println("has 41:", c.Has("41"))
	fmt.Println("has 50:", c.Has("50"))

	fmt.Println(c.Get("41")) // access the value

	c.Put("51", 51)

	fmt.Println("has 42:", c.Has("42")) // gone now
	fmt.Println(c.Get("41"))            // still around

	c.Clear()
	fmt.Println("size:", c.Size())

	// Output:
	// size: 10
	// has 1: false
	// has 40: false
	// has 41: true
	// has 50: true
	// 41 true
	// has 42: false
	// 41 true
	// size: 0
}
