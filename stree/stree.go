// Package stree implements the Scapegoat Tree, an approximately-balanced
// binary search tree.
//
// A scapegoat tree supports worst-case O(lg n) lookup and amortized O(lg n)
// insertion and delection (the worst-case cost of a single insert or delete
// operation is O(n)).
//
// Scapegoat trees are relatively memory-efficient, as interior nodes do not
// require any ancillary metadata for balancing purposes, and the tree itself
// costs only a few words of bookkeeping overhead beyond the nodes.  More
// interestingly, each rebalancing operation requires only a single contiguous
// vector allocation.
//
// The scapegoat tree algorithm is described by the paper:
//
//	I. Galperin, R. Rivest: "Scapegoat Trees"
//	https://people.csail.mit.edu/rivest/pubs/GR93.pdf
package stree

import (
	"math"
	"sort"
)

const (
	maxBalance = 1000
	fracLimit  = 2 * maxBalance
)

// New returns a new tree with the given balancing factor 0 ≤ β ≤ 1000. The
// order of elements stored in the tree is provided by the comparison function.
//
// If any keys are given, the tree is initialized to contain them, otherwise an
// empty tree is created.  For keys that are known in advance it is more
// efficient to allocate storage for them at construction time, versus adding
// them separately later.
//
// The balancing factor controls how unbalanced the tree is permitted to be,
// with 0 being strictest (as near as possible to 50% weight balance) and 1000
// being loosest (no rebalancing). Stricter balance incurs more overhead for
// rebalancing, but provides faster lookups. A good default is 250.
//
// New panics if β < 0 or β > 1000.
func New[T any](β int, lessThan func(a, b T) bool, keys ...T) *Tree[T] {
	if β < 0 || β > maxBalance {
		panic("β out of range")
	}
	tree := &Tree[T]{
		β:        β,
		lessThan: lessThan,
		limit:    limitFunc(β),
		size:     len(keys),
		max:      len(keys),
	}
	if len(keys) != 0 {
		nodes := make([]*node[T], len(keys))
		for i, key := range keys {
			nodes[i] = &node[T]{X: key}
		}
		sort.Slice(nodes, func(i, j int) bool {
			return lessThan(nodes[i].X, nodes[j].X)
		})
		tree.root = extract(nodes)
	}
	return tree
}

// A Tree is the root of a scapegoat tree. A *Tree is not safe for concurrent
// use without external synchronization.
type Tree[T any] struct {
	root *node[T]

	// β identifies a point on the interval [maxBalance,fracLimit], and we
	// compute the balance fraction as β/fracLimit. This permits breakpoint
	// computations to use only fixed-point integer arithmetic and only
	// requires one floating-point operation per insertion to recompute the
	// depth limit.

	β        int               // balancing factor
	lessThan func(a, b T) bool // key comparison
	limit    func(n int) int   // depth limit for size n
	size     int               // cache of root.size()
	max      int               // max of size since last rebuild of root
}

func toFraction(β int) float64 { return (float64(β) + maxBalance) / fracLimit }

// limitFunc returns a function that computes the depth limit for a tree of
// size n given the balance factor β.
func limitFunc(β int) func(int) int {
	inv := 1 / toFraction(β)
	if inv == 1 { // int(+Inf) ⇒ undefined
		return func(n int) int { return n + 1 }
	}
	base := math.Log(inv)
	return func(n int) int { return int(math.Log(float64(n)) / base) }
}

// Add inserts key into the tree. If key is already present, Add returns false
// without modifying the tree. Otherwise it adds the key and returns true.
func (t *Tree[T]) Add(key T) bool {
	// We don't yet know whether the insertion will add mass to the tree; we
	// conservatively assume it might for purposes of choosing a depth limit.
	ins, ok, _, _ := t.insert(key, false, t.root, t.limit(t.size+1))
	t.incSize(ok)
	t.root = ins
	return ok
}

// Replace inserts key into the tree. If key is already present, Replace
// updates the existing value and returns false. Otherwise it adds key and
// returns true.
func (t *Tree[T]) Replace(key T) bool {
	ins, ok, _, _ := t.insert(key, true, t.root, t.limit(t.size+1))
	t.incSize(ok)
	t.root = ins
	return ok
}

// incSize increments t.size and updates t.max if inserted is true.
func (t *Tree[T]) incSize(inserted bool) {
	if inserted {
		t.size++
		if t.size > t.max {
			t.max = t.size
		}
	}
}

// insert key in order under root, with the given depth limit.
//
// If replace is true and an existing node has an equivalent key, it is updated
// with the given key; otherwise, inserting an existing key is a no-op.
//
// Returns the modified tree, and reports whether a new node was added and the
// height of the returned node above the point of insertion.
// If the insertion did not exceed the depth limit, size == 0.
// Otherwise, size == ins.size() meaning a scapegoat is needed.
func (t *Tree[T]) insert(key T, replace bool, root *node[T], limit int) (ins *node[T], added bool, size, height int) {
	// Descending phase: Insert the key into the tree structure.
	var sib *node[T]
	if root == nil {
		if limit < 0 {
			size = 1
		}
		return &node[T]{X: key}, true, size, 0
	} else if t.lessThan(key, root.X) {
		ins, added, size, height = t.insert(key, replace, root.left, limit-1)
		root.left = ins
		sib = root.right
		height++
	} else if t.lessThan(root.X, key) {
		ins, added, size, height = t.insert(key, replace, root.right, limit-1)
		root.right = ins
		sib = root.left
		height++
	} else {
		// Replacing an existing node. This cannot introduce a violation, so we
		// can return immediately without triggering a goat search.
		if replace {
			root.X = key
		}
		return root, false, 0, 0
	}

	// Ascending phase, a.k.a., goat rodeo.
	// Uses the selection strategy from section 4.6 of Galperin & Rivest .

	// If size != 0, we exceeded the depth limit and are looking for a goat.
	// Note: size == ins.size() not root.size() at this point.
	if size > 0 {
		sibSize := sib.size()          // size of sibling subtree
		rootSize := sibSize + 1 + size // new size of root

		if bw := t.limit(rootSize); height <= bw {
			size = rootSize // not the goat you're looking for; move along
		} else {
			// root is the goat; rewrite it and signal the activations above us
			// to stop looking by setting size to 0.
			root = rewrite(root, rootSize)
			size = 0
		}
	}
	return root, added, size, height
}

// Remove key from the tree and report whether it was present.
func (t *Tree[T]) Remove(key T) bool {
	del, ok := t.root.remove(key, t.lessThan)
	t.root = del
	if ok {
		t.size--
		if bw := (t.max*t.β + maxBalance) / fracLimit; t.size < bw {
			t.root = rewrite(t.root, t.size)
			t.max = t.size
		}
	}
	return ok
}

// remove key from the subtree under n, returning the modified tree reporting
// whether the mass of the tree was decreased.
func (n *node[T]) remove(key T, lessThan func(a, b T) bool) (_ *node[T], ok bool) {
	if n == nil {
		return nil, false // nothing to do
	} else if lessThan(key, n.X) {
		n.left, ok = n.left.remove(key, lessThan)
		return n, ok
	} else if lessThan(n.X, key) {
		n.right, ok = n.right.remove(key, lessThan)
		return n, ok
	} else if n.left == nil {
		return n.right, true
	} else if n.right == nil {
		return n.left, true
	}

	// At this point we need to remove n, but it has two children.
	// Do the usual trick.
	goat := popMinRight(n)
	n.X = goat.X
	return n, true
}

// Len reports the number of elements stored in the tree. This is a
// constant-time query.
func (t *Tree[T]) Len() int { return t.size }

// IsEmpty reports whether t is empty.
func (t *Tree[T]) IsEmpty() bool { return t.size == 0 }

// Clear discards all the values in t, leaving it empty.
func (t *Tree[T]) Clear() { t.size = 0; t.max = 0; t.root = nil }

// Get reports whether key is present in the tree, and returns the matching key
// if so, or a zero value if the key is not present.
func (t *Tree[T]) Get(key T) (_ T, ok bool) {
	cur := t.root
	for cur != nil {
		if t.lessThan(key, cur.X) {
			cur = cur.left
		} else if t.lessThan(cur.X, key) {
			cur = cur.right
		} else {
			return cur.X, true
		}
	}
	return
}

// Inorder calls f for each key of t in order. If f returns false, Inorder
// stops and returns false; otherwise it returns true after visiting all
// elements of t.
func (t *Tree[T]) Inorder(f func(key T) bool) bool { return t.root.inorder(f) }

// InorderAfter calls f for each key greater than or equal to key, in order.
// if f returns false, InorderAfter stops and returns fales. Otherwise, it
// returns true after visiting all eligible elements of t.
func (t *Tree[T]) InorderAfter(key T, f func(key T) bool) bool {
	return t.root.inorderAfter(key, t.lessThan, f)
}

// Min returns the minimum key from t. If t is empty, Min returns a zero key.
func (t *Tree[T]) Min() T {
	if t.root == nil {
		var zero T
		return zero
	}
	cur := t.root
	for cur.left != nil {
		cur = cur.left
	}
	return cur.X
}

// Max returns the maximum key from t. If t is empty, Max returns a zero key.
func (t *Tree[T]) Max() T {
	if t.root == nil {
		var zero T
		return zero
	}
	cur := t.root
	for cur.right != nil {
		cur = cur.right
	}
	return cur.X
}
