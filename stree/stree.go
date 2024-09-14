// Package stree implements the Scapegoat Tree, an approximately-balanced
// binary search tree.
//
// A scapegoat tree supports worst-case O(lg n) lookup and amortized O(lg n)
// insertion and deletion (the worst-case cost of a single insert or delete
// operation is O(n)).
//
// Scapegoat trees are relatively memory-efficient, as interior nodes do not
// require any ancillary metadata for balancing purposes, and the tree itself
// costs only a few words of bookkeeping overhead beyond the nodes. Rebalancing
// uses the Day-Stout-Warren (DSW) in-place algorithm, which does not require
// any additional heap allocations.
//
// The scapegoat tree algorithm is described by the paper:
//
//	I. Galperin, R. Rivest: "Scapegoat Trees"
//	https://people.csail.mit.edu/rivest/pubs/GR93.pdf
package stree

import (
	"fmt"
	"iter"
	"math"
	"slices"
)

const (
	maxBalance = 1000
	fracLimit  = 2 * maxBalance
)

// New returns a new tree with the given balancing factor 0 ≤ β ≤ 1000. The
// order of elements stored in the tree is provided by the comparison function,
// where compare(a, b) must be <0 if a < b, =0 if a == b, and >0 if a > b.
//
// If any keys are given, the tree is initialized to contain them, otherwise an
// empty tree is created.  When the initial set of keys is known in advance it
// is more efficient to add them during tree construction, versus versus adding
// them separately later.
//
// The balancing factor controls how unbalanced the tree is permitted to be,
// with 0 being strictest (as near as possible to 50% weight balance) and 1000
// being loosest (no rebalancing). Stricter balance incurs more overhead for
// rebalancing, but provides faster lookups. A good default is 250.
//
// New panics if β < 0 or β > 1000.
func New[T any](β int, compare func(a, b T) int, keys ...T) *Tree[T] {
	if β < 0 || β > maxBalance {
		panic("β out of range")
	}
	tree := &Tree[T]{
		β:       β,
		compare: compare,
		limit:   limitFunc(β),
	}
	if len(keys) != 0 {
		nodes := make([]*node[T], len(keys))
		for i, key := range keys {
			nodes[i] = &node[T]{X: key}
		}
		slices.SortFunc(nodes, func(a, b *node[T]) int {
			return compare(a.X, b.X)
		})
		nodes = slices.CompactFunc(nodes, func(a, b *node[T]) bool {
			return compare(a.X, b.X) == 0
		})
		tree.size = len(nodes)
		tree.max = len(nodes)
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

	β       int              // balancing factor
	compare func(a, b T) int // key comparison
	limit   func(n int) int  // depth limit for size n
	size    int              // cache of root.size()
	max     int              // max of size since last rebuild of root
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

// Clone returns a deep copy of t with identical settings. Operations on the
// clone do not affect t and vice versa.
func (t *Tree[T]) Clone() *Tree[T] {
	cp := *t                 // shallow copy of the top-level structures
	cp.root = t.root.clone() // deep copy of the contents
	return &cp
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
		// This is the base case where we add mass to the tree.  If we exceeded
		// the depth limit reaching this point, flag it to the caller by
		// returning a non-zero size. The 1 accounts for the node we just
		// inserted; the caller will update it as we unwind the insertion.
		if limit < 0 {
			size = 1
		}
		return &node[T]{X: key}, true, size, 0
	}
	cmp := t.compare(key, root.X)
	if cmp < 0 {
		ins, added, size, height = t.insert(key, replace, root.left, limit-1)
		root.left = ins
		sib = root.right
		height++
	} else if cmp > 0 {
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
	// Uses the selection strategy from section 4.6 of Galperin & Rivest.

	// If size != 0, we exceeded the depth limit and are looking for a goat.
	// Note: size == ins.size() not root.size() at this point.
	if size > 0 {
		sibSize := sib.size()          // size of sibling subtree
		rootSize := sibSize + 1 + size // new size of root

		if bw := t.limit(rootSize); height <= bw {
			// Update the size to include the rest of the current root, but this
			// is not the scapegoat yet. Keep unwinding.
			size = rootSize
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
	del, ok := t.root.remove(key, t.compare)
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
func (n *node[T]) remove(key T, compare func(a, b T) int) (_ *node[T], ok bool) {
	if n == nil {
		return nil, false // nothing to do
	}
	cmp := compare(key, n.X)
	if cmp < 0 {
		n.left, ok = n.left.remove(key, compare)
		return n, ok
	} else if cmp > 0 {
		n.right, ok = n.right.remove(key, compare)
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

func (t *Tree[T]) String() string {
	return fmt.Sprintf("stree.Tree(β=%d:size=%d)", t.β, t.size)
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
		cmp := t.compare(key, cur.X)
		if cmp < 0 {
			cur = cur.left
		} else if cmp > 0 {
			cur = cur.right
		} else {
			return cur.X, true
		}
	}
	return
}

// Inorder is a range function that visits each key of t in order.
func (t *Tree[T]) Inorder(yield func(key T) bool) { t.root.inorder(yield) }

// InorderAfter returns a range function for each key greater than or equal to
// key, in order.
func (t *Tree[T]) InorderAfter(key T) iter.Seq[T] {
	return func(yield func(T) bool) {
		t.root.inorderAfter(key, t.compare, yield)
	}
}

// Cursor constructs a cursor to the specified key, or nil if key is not
// present in the tree.
func (t *Tree[T]) Cursor(key T) *Cursor[T] {
	path := t.root.pathTo(key, t.compare)
	if len(path) == 0 || t.compare(path[len(path)-1].X, key) != 0 {
		return nil
	}
	return &Cursor[T]{path: path}
}

// Root returns a Cursor to the root of t, or nil if t is empty.
func (t *Tree[T]) Root() *Cursor[T] {
	if t.root == nil {
		return nil
	}
	return &Cursor[T]{path: []*node[T]{t.root}}
}

// Min returns the minimum key in t. If t is empty, a zero key is returned.
func (t *Tree[T]) Min() T {
	cur := t.root
	if cur == nil {
		var zero T
		return zero
	}
	for cur.left != nil {
		cur = cur.left
	}
	return cur.X
}

// Max returns the maximum key in t. If t is empty, a zero key is returned.
func (t *Tree[T]) Max() T {
	cur := t.root
	if cur == nil {
		var zero T
		return zero
	}
	for cur.right != nil {
		cur = cur.right
	}
	return cur.X
}

// KV is a convenience type for storing key-value pairs in a Tree, where the
// key type T is used for comparison while the value type U is ignored.  Use
// the Compare method to adapt a comparison for T to a KV on T.
//
// For convenience of notation, you can create a type alias for an
// instantiation of this type:
//
//	type metrics = stree.KV[string, float64]
//	compare := metrics{}.Compare(cmp.Compare)
type KV[T, U any] struct {
	Key   T
	Value U
}

// Compare converts a function comparing values of type T into a function that
// compares the Key field of a KV[T, U].
func (KV[T, U]) Compare(compare func(a, b T) int) func(ka, kb KV[T, U]) int {
	return func(ka, kb KV[T, U]) int { return compare(ka.Key, kb.Key) }
}
