package stree

import "slices"

// A Cursor is an anchor to a location within a Tree that can be used to
// navigate the structure of the tree. A cursor is Valid if it points to a
// non-empty subtree of its tree.
type Cursor[T any] struct {
	// The sequence of nodes from the root to the current item.
	// The pointers are shared with the underlying tree.
	// If this is empty, the cursor is invalid.
	path []*node[T]
}

// Valid reports whether c is a valid cursor, meaning it points to a non-empty
// subtree of its containing tree. A nil Cursor is treated as invalid.
func (c *Cursor[T]) Valid() bool { return c != nil && len(c.path) != 0 }

// Clone returns a clone of c that points to the same location, but which is
// unaffected by subsequent movement of c (and vice versa).
func (c *Cursor[T]) Clone() *Cursor[T] {
	if !c.Valid() {
		return c
	}
	return &Cursor[T]{path: slices.Clone(c.path)}
}

// Key returns the key at the current location of the cursor.
// An invalid Cursor returns a zero-valued key.
func (c *Cursor[T]) Key() T {
	if c.Valid() {
		return c.path[len(c.path)-1].X
	}
	var zero T
	return zero
}

// findNext reports the location of the successor of c.
// If no successor exists, it returns (nil, -1).
//
// If the successor is a descendant of c, it returns (right, -1) where right is
// the right child of c.
//
// Otherwise the successor is an ancestor of c, and it returns (nil, i) giving
// the offset i in the path where that ancestor is located.
//
// Precondition: c is valid.
func (c *Cursor[T]) findNext() (*node[T], int) {
	i := len(c.path) - 1

	// If the current node has a right subtree, its successor is there.
	if min := c.path[i].right; min != nil {
		return min, -1
	}

	// Otherwise, we have to walk back up the tree.  If the current node is the
	// left child of its parent, the parent is its successor. If not, we keep
	// going until we find an ancestor that IS the left child of its parent.  If
	// no such ancestor exists, there is no successor.
	j := i - 1 // j is parent, i is child
	for j >= 0 {
		// The current node is the left child of its parent.
		if c.path[i] == c.path[j].left {
			return nil, j
		}
		i = j
		j--
	}
	return nil, -1
}

// HasNext reports whether c has a successor.
// An invalid cursor has no successor.
func (c *Cursor[T]) HasNext() bool {
	if c.Valid() {
		n, i := c.findNext()
		return n != nil || i >= 0
	}
	return false
}

// Next advances c to its successor in the tree, and returns c.
// If c had no successor, it becomes invalid.
func (c *Cursor[T]) Next() *Cursor[T] {
	if c.Valid() {
		min, j := c.findNext()
		if min != nil {
			for ; min != nil; min = min.left {
				c.path = append(c.path, min)
			}
		} else if j >= 0 {
			c.path = c.path[:j+1]
		} else {
			c.path = nil
		}
	}
	return c
}

// findPrev reports the location of the predecessor of c.
// If no predecessor exists, it returns (nil, -1).
//
// If the predecessor is a descendant of c, it returns (left, -1) where left is
// the left child of c.
//
// Otherwise the predecessoris an ancestor of c, and it returns (nil, i) giving
// the offset i in the path where that ancestor is located.
//
// Precondition: c is valid.
func (c *Cursor[T]) findPrev() (*node[T], int) {
	i := len(c.path) - 1

	// If the current node has a left subtree, its predecessor is there.
	if max := c.path[i].left; max != nil {
		return max, -1
	}

	// Otherwise, we have to walk back up the tree.  If the current node is the
	// right child of its parent, the parent is its predecessor. If not, we keep
	// going until we find an ancestor that IS the right child of its parent.
	// If no such ancestor exists, there is no predecessor.
	j := i - 1 // j is parent, i is child
	for j >= 0 {
		// The current node is the right child of its parent.
		if c.path[i] == c.path[j].right {
			return nil, j
		}
		i = j
		j--
	}
	return nil, -1
}

// HasPrev reports whether c has a predecessor.
// An invalid cursor has no predecessor.
func (c *Cursor[T]) HasPrev() bool {
	if c.Valid() {
		n, i := c.findPrev()
		return n != nil || i >= 0
	}
	return false
}

// Prev advances c to its predecessor in the tree, and returns c.
// If c had no predecessor, it becomes invalid.
func (c *Cursor[T]) Prev() *Cursor[T] {
	if c.Valid() {
		max, j := c.findPrev()
		if max != nil {
			for ; max != nil; max = max.right {
				c.path = append(c.path, max)
			}
		} else if j >= 0 {
			c.path = c.path[:j+1]
		} else {
			c.path = nil
		}
	}
	return c
}

// HasLeft reports whether c has a non-empty left subtree.
// An invalid cursor has no left subtree.
func (c *Cursor[t]) HasLeft() bool { return c.Valid() && c.path[len(c.path)-1].left != nil }

// Left moves to the left subtree of c, and returns c.
// If c had no left subtree, it becomes invalid.
func (c *Cursor[T]) Left() *Cursor[T] {
	if c.Valid() {
		if left := c.path[len(c.path)-1].left; left != nil {
			c.path = append(c.path, left)
		} else {
			c.path = nil // invalidate
		}
	}
	return c
}

// HasRight reports whether c has a non-empty right subtree.
// An invalid cursor has no right subtree.
func (c *Cursor[t]) HasRight() bool { return c.Valid() && c.path[len(c.path)-1].right != nil }

// Right moves to the right subtree of c, and returns c.
// If c had no right subtree, it becomes invalid.
func (c *Cursor[T]) Right() *Cursor[T] {
	if c.Valid() {
		if right := c.path[len(c.path)-1].right; right != nil {
			c.path = append(c.path, right)
		} else {
			c.path = nil // invalidate
		}
	}
	return c
}

// HasParent reports whether c has a parent.
// An invalid cursor has no parent.
func (c *Cursor[T]) HasParent() bool { return c.Valid() && len(c.path) > 1 }

// Up moves to the parent of c, and returns c.
// If c had no parent, it becomes invalid..
func (c *Cursor[T]) Up() *Cursor[T] {
	if c.Valid() {
		// Note that this may result in c being invalid, if it was already
		// pointed at the root of the tree.
		c.path = c.path[:len(c.path)-1]
	}
	return c
}

// Min moves c to the minimum element of its subtree, and returns c.
func (c *Cursor[T]) Min() *Cursor[T] {
	if c.Valid() {
		min := c.path[len(c.path)-1]
		for min.left != nil {
			min = min.left
			c.path = append(c.path, min)
		}
	}
	return c
}

// Max moves c to the maximum element of its subtree, and returns c.
func (c *Cursor[T]) Max() *Cursor[T] {
	if c.Valid() {
		max := c.path[len(c.path)-1]
		for max.right != nil {
			max = max.right
			c.path = append(c.path, max)
		}
	}
	return c
}

// Inorder calls f for each key of the subtree rooted at c in order. If f
// returns false, Inorder stops and returns false; otherwise it returns true
// after visiting all elements of c.
func (c *Cursor[T]) Inorder(f func(key T) bool) bool {
	if c.Valid() {
		return c.path[len(c.path)-1].inorder(f)
	}
	return true
}
