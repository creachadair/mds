// Package mapset implements a basic set type using a built-in map.
//
// The Set type is a thin wrapper on a built-in Go map, so a Set is not safe
// for concurrent use without external synchronization.
package mapset

// A Set represents a set of distinct values. It is implemented via the
// built-in map type, and the underlying map can also be used directly to add
// and remove items and to iterate the contents.
type Set[T comparable] map[T]struct{}

// New constructs a set of the specified items. The result is never nil, even
// if no items are provided.
func New[T comparable](items ...T) Set[T] {
	m := make(Set[T], len(items))
	return m.add(items)
}

// NewSize constructs a new set preallocated to have space for n items.
func NewSize[T comparable](n int) Set[T] { return make(Set[T], n) }

// IsEmpty reports whether s is empty.
func (s Set[T]) IsEmpty() bool { return len(s) == 0 }

// Len reports the number of elements in s.
func (s Set[T]) Len() int { return len(s) }

// Clear removes all elements from s and returns s.
func (s Set[T]) Clear() Set[T] {
	for item := range s {
		// N.B.: This has the usual limitations with weird values like NaN;
		// see https://github.com/golang/go/issues/56351.
		delete(s, item)
	}
	return s
}

// Clone returns a new set with the same contents as s.
// The value returned is never nil.
func (s Set[T]) Clone() Set[T] {
	m := make(Set[T], len(s))
	for item := range s {
		m[item] = struct{}{}
	}
	return m
}

// Has reports whether t is present in the set.
func (s Set[T]) Has(t T) bool { _, ok := s[t]; return ok }

// Add adds the specified items to the set and returns s.
func (s *Set[T]) Add(items ...T) Set[T] {
	if *s == nil {
		*s = make(Set[T], len(items))
	}
	return (*s).add(items)
}

func (s Set[T]) add(items []T) Set[T] {
	for _, item := range items {
		s[item] = struct{}{}
	}
	return s
}

// AddAll adds all the elements of set t to s and returns s.
func (s *Set[T]) AddAll(t Set[T]) Set[T] {
	if *s == nil {
		*s = t.Clone()
		return *s
	}
	return (*s).addAll(t)
}

func (s Set[T]) addAll(t Set[T]) Set[T] {
	for item := range t {
		s[item] = struct{}{}
	}
	return s
}

// Remove removes the specified items from the set and returns s.
func (s Set[T]) Remove(items ...T) Set[T] {
	if len(s) != 0 {
		for _, item := range items {
			delete(s, item)
		}
	}
	return s
}

// RemoveAll removes all the elements of set t from s and returns s.
func (s Set[T]) RemoveAll(t Set[T]) Set[T] {
	if len(s) != 0 {
		for item := range t {
			delete(s, item)
		}
	}
	return s
}

// Pop removes and returns an arbitrary element of s, if s is non-empty.
// If s is empty, it returns a zero value.
func (s Set[T]) Pop() T {
	for item := range s {
		delete(s, item)
		return item
	}
	var zero T
	return zero
}

// Intersects reports whether s and t share any elements in common.
func (s Set[T]) Intersects(t Set[T]) bool {
	lo, hi := s, t
	if len(s) > len(t) {
		lo, hi = hi, lo
	}
	for item := range lo {
		if hi.Has(item) {
			return true
		}
	}
	return false
}

// HasAny reports whether s contains any element of ts. It returns false if
// len(ts) == 0.
func (s Set[T]) HasAny(ts ...T) bool {
	if len(s) == 0 {
		return false
	}
	for _, t := range ts {
		if s.Has(t) {
			return true
		}
	}
	return false
}

// IsSubset reports whether s is a subset of t.
func (s Set[T]) IsSubset(t Set[T]) bool {
	if len(s) == 0 {
		return true
	} else if len(s) > len(t) {
		return false
	}
	for item := range s {
		if !t.Has(item) {
			return false
		}
	}
	return true
}

// Equals reports whether s and t contain exactly the same elements.
func (s Set[T]) Equals(t Set[T]) bool {
	if len(s) != len(t) {
		return false
	}
	for item := range s {
		if !t.Has(item) {
			return false
		}
	}
	return true
}

// Slice returns a slice of the contents of s in arbitrary order.
func (s Set[T]) Slice() []T {
	if len(s) == 0 {
		return nil
	}
	items := make([]T, 0, len(s))
	for item := range s {
		items = append(items, item)
	}
	return items
}

// Intersect constructs a new set containing the intersection of the specified
// sets.  The result is never nil, even if the given sets are empty.
func Intersect[T comparable](s1, s2 Set[T]) Set[T] {
	n1, n2 := len(s1), len(s2)
	if n1 == 0 || n2 == 0 {
		return make(Set[T])
	}
	if n1 > n2 {
		s1, s2 = s2, s1
	}
	out := make(Set[T], len(s1))
	for v := range s1 {
		if s2.Has(v) {
			out.Add(v)
		}
	}
	return out
}
