package stree

import (
	"cmp"
	"math"
	"testing"
)

func TestVine(t *testing.T) {
	const numElem = 25

	// Construct a tree with consecutive integers.
	tree := New(100, cmp.Compare[int])
	for i := range numElem {
		tree.Add(i + 1)
	}

	// Flatten the tree node into a right-linked vine and verify that the result
	// contains the original elements.
	hd := treeToVine(tree.root)

	t.Run("Collapse", func(t *testing.T) {
		i := 0
		for cur := hd; cur != nil; cur = cur.right {
			i++
			if cur.X != i {
				t.Errorf("Node value: got %d, want %d", cur.X, i)
			}
			if cur.left != nil {
				t.Errorf("Node %d has a non-nil left pointer: %v", i, cur.left)
			}
		}

		if i != numElem {
			t.Errorf("Got %d nodes, want %d", i, numElem)
		}
	})

	// Reconstitute the tree and verify it is balanced and properly ordered.
	t.Run("Rebuild", func(t *testing.T) {
		rec := vineToTree(hd, numElem)
		want := int(math.Ceil(math.Log2(numElem)))
		if got := rec.height(); got > want {
			t.Errorf("Got height %d, want %d", got, want)
		}

		i := 0
		rec.inorder(func(z int) bool {
			i++
			if z != i {
				t.Errorf("Node value: got %d, want %d", z, i)
			}
			return true
		})

		if i != numElem {
			t.Errorf("Got %d nodes, want %d", i, numElem)
		}
	})
}

func (n *node[T]) height() int {
	if n == nil {
		return 0
	}
	return max(n.left.height(), n.right.height()) + 1
}
