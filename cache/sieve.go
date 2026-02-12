package cache

import (
	"fmt"

	"github.com/creachadair/mds/ring"
)

// An entry holds the state of a single cached key-value pair.
type entry[Key comparable, Value any] struct {
	Key     Key
	Value   Value
	Visited bool
}

type handle[Key comparable, Value any] = *ring.Ring[entry[Key, Value]]

// sieveStore is an implementation of the [Store] interface.
type sieveStore[Key comparable, Value any] struct {
	present map[Key]handle[Key, Value]
	queue   handle[Key, Value]
	hand    handle[Key, Value]
}

// Sieve constructs a [Config] with a cache store that manages entries with the
// [SIEVE] eviction algorithm.
//
// [SIEVE]: https://junchengyang.com/publication/nsdi24-SIEVE.pdf
func Sieve[Key comparable, Value any]() Config[Key, Value] {
	s := &sieveStore[Key, Value]{
		present: make(map[Key]handle[Key, Value]),
		queue:   ring.New[entry[Key, Value]](1), // sentinel
	}
	s.hand = s.queue
	return Config[Key, Value]{store: s}
}

// Check implements part of the [Store] interface.
func (s *sieveStore[Key, Value]) Check(key Key) (Value, bool) {
	e, ok := s.present[key]
	if ok {
		return e.Value.Value, true
	}
	var zero Value
	return zero, false
}

// Access implements part of the [Store] interface.
func (s *sieveStore[Key, Value]) Access(key Key) (Value, bool) {
	e, ok := s.present[key]
	if ok {
		e.Value.Visited = true
		return e.Value.Value, true
	}
	var zero Value
	return zero, false
}

// Store implements part of the [Store] interface.
func (s *sieveStore[Key, Value]) Store(key Key, val Value) {
	if _, ok := s.present[key]; ok {
		panic(fmt.Sprintf("sieve store: unexpected key %v", key))
	}

	e := ring.Of(entry[Key, Value]{
		Key:   key,
		Value: val,
	})
	s.queue.Join(e)
	s.present[key] = e
}

// Remove implements part of the [Store] interface.
func (s *sieveStore[Key, _]) Remove(key Key) {
	e, ok := s.present[key]
	if ok {
		if s.hand == e {
			s.hand = s.hand.Next()
		}
		e.Pop()
		delete(s.present, key)
	}
}

// Evict implements part of the [Store] interface.
func (s *sieveStore[Key, Value]) Evict() (Key, Value) {
	for s.hand.Value.Visited {
		s.hand.Value.Visited = false
		s.hand = s.hand.Prev()
	}
	if s.hand == s.queue {
		s.hand = s.hand.Prev()
	}
	out := s.hand
	s.hand = s.hand.Prev()
	out.Pop()
	delete(s.present, out.Value.Key)
	return out.Value.Key, out.Value.Value
}
