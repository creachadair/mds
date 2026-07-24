// Copyright (C) Michael J. Fromberger. All Rights Reserved.

// Package expirer implements a type that manages timed expiration of keys.
package expirer

import (
	"sync"
	"time"

	"github.com/creachadair/mds/heapq"
	"github.com/creachadair/msync/trigger"
)

// An Expirer manages expiration deadlines for a collection of keys.
// A nil value is ready for use (its methods silently do nothing).
type Expirer[Key comparable] struct {
	check     *trigger.Cond // signalled when the queue changes
	removeKey func(Key)     // remove the specified key from the cache

	μ       sync.Mutex
	present map[Key]int // :: Key → offset in queue
	queue   *heapq.Queue[timerKey[Key]]
	running bool

	// The queue is ordered in nondecreasing order by deadline, so that the
	// earliest deadline is always at the front.
}

// New constructs a new empty [Expirer] that uses removeKey to report expired keys.
// The removeKey callback must not be nil.
func New[Key comparable](removeKey func(Key)) *Expirer[Key] {
	if removeKey == nil {
		panic("nil removeKey callback")
	}
	m := make(map[Key]int)
	q := heapq.New(timerKey[Key].Compare).SetUpdate(func(t timerKey[Key], npos int) {
		m[t.Key] = npos
	})
	return &Expirer[Key]{
		check:     trigger.New(),
		removeKey: removeKey,
		present:   m,
		queue:     q,
	}
}

// SetExpiration adds or updates the expiration for key to deadline.
// It does nothing if e == nil.
func (e *Expirer[Key]) SetExpiration(key Key, deadline time.Time) {
	if e == nil {
		return
	}
	e.μ.Lock()
	defer e.μ.Unlock()
	e.startIfNeededLocked()
	if i, ok := e.present[key]; ok {
		e.queue.Remove(i)
	}
	e.queue.Add(timerKey[Key]{Key: key, Deadline: deadline})
	e.check.Signal()
}

// RemoveExpiration removes any pending expiration for the specified key.
// It does nothing if e == nil.
func (e *Expirer[Key]) RemoveExpiration(key Key) {
	if e == nil {
		return
	}
	e.μ.Lock()
	defer e.μ.Unlock()
	if i, ok := e.present[key]; ok {
		e.queue.Remove(i)
		delete(e.present, key)
		e.check.Signal()
	}
}

// ClearExpirations discards all pending expirations for all keys.
// It does nothing if e == nil.
func (e *Expirer[Key]) ClearExpirations() {
	if e == nil {
		return
	}
	e.μ.Lock()
	defer e.μ.Unlock()
	e.queue.Clear()
	clear(e.present)
	e.check.Signal()
}

func (e *Expirer[Key]) startIfNeededLocked() {
	if e.running {
		return // already running
	}
	e.running = true

	// Start the service goroutine for the expiration queue.
	// This is started (or re-started) whenever the queue transitions from empty
	// to non-empty, and runs until it has drained the queue completely.
	go func() {
		for {
			changed := e.check.Ready()
			wake, expired, isEmpty := e.scanQueue()
			for _, key := range expired {
				e.removeKey(key)
			}
			if isEmpty {
				return // no more expirations are pending
			}
			select {
			case <-wake:
				// the deadline for the frontmost queue element has passed
			case <-changed:
				// the queue changed, so it is now possible that the last wakeup
				// time we chose in scanQueue is no longer valid.
			}
		}
	}()
}

// scanQueue removes and returnsall keys whose expiration is past from the
// front of the queue, and reports whether the remaining queue is empty.
//
// If the queue is not empty, it also returns a channel that is ready when the
// next unexpired item is expected to expire.
func (e *Expirer[Key]) scanQueue() (wake <-chan time.Time, expired []Key, isEmpty bool) {
	e.μ.Lock()
	defer e.μ.Unlock()
	now := time.Now()
	for {
		next, ok := e.queue.Peek(0)
		if !ok {
			break
		}
		if next.Deadline.After(now) {
			return time.After(next.Deadline.Sub(now)), expired, false
		}
		e.queue.Pop()
		delete(e.present, next.Key)
		expired = append(expired, next.Key)
	}

	// The service goroutine is still running at this instant, but will exit
	// without doing further work upon learning that the queue is now empty.
	// Clear the flag now, so that if a new item arrives during the interval
	// SetExpiration will know to start a new one.
	e.running = false
	return nil, expired, true
}

// A timerKey packages a cache key with an expiration deadline.
type timerKey[Key any] struct {
	Key      Key
	Deadline time.Time
}

func (t timerKey[Key]) Compare(u timerKey[Key]) int { return t.Deadline.Compare(u.Deadline) }
