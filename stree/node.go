package stree

type node[T any] struct {
	X           T
	left, right *node[T]
}

// clone returns a deep copy of n.
func (n *node[T]) clone() *node[T] {
	if n == nil {
		return nil
	}
	return &node[T]{X: n.X, left: n.left.clone(), right: n.right.clone()}
}

// size reports the number of nodes contained in the tree rooted at n.
// If n == nil, this is defined as 0.
func (n *node[T]) size() int {
	if n == nil {
		return 0
	}
	return 1 + n.left.size() + n.right.size()
}

// treeToVine rewrites the tree rooted at n into an inorder linked list, and
// returns the first element of the list. The nodes are modified in-place and
// linked via their right pointers; the left pointers of all the nodes are set
// to nil.
//
// This conversion uses the in-place iterative approach of Stout & Warren,
// where the tree is denormalized in-place via rightward rotations.
func treeToVine[T any](n *node[T]) *node[T] {
	stub := &node[T]{right: n} // sentinel
	cur := stub
	for cur.right != nil {
		C := cur.right
		if C.left == nil {
			cur = C
			continue
		}

		// Right rotate:
		//   cur     into     cur
		//     \                \
		//      C                L
		//     / \              / \
		//    L   z            x   C
		//   / \                  / \
		//  x   y                y   z
		L := C.left
		C.left = L.right // y ← C
		L.right = C      // L → C
		cur.right = L    // n → L
	}
	return stub.right
}

// rotateLeft rewrites the chain of nodes starting from n and linked by right
// pointers, by left-rotating the specified number of times. This function will
// panic if count exceeds the length of the chain.
//
// A single left-rotation transforms:
//
//	n      into       n
//	 \                 \
//	  C                 R
//	 / \               / \
//	x   R             C   z
//	   / \           / \
//	  y   z         x   y
//
// Note that any of x, y, and z may be nil.
func rotateLeft[T any](n *node[T], count int) {
	next := n
	for range count {
		C := next.right
		R := C.right

		C.right = R.left // C → y
		R.left = C       // C ← R
		next.right = R   // n → R

		next = R // advance
	}
}

// vineToTree rewrites the chain of count nodes starting from n and linked by
// right pointers, into a balanced tree rooted at n. It returns the root of the
// resulting new tree. It will panic if count exceeds the chain length.
//
// This uses Stout & Warren's extension of Day's algorithm that produces a tree
// that is "full" (as much as possible), with leaves filled left-to-right.
func vineToTree[T any](n *node[T], count int) *node[T] {
	// Compute the largest power of 2 no greater than count, less 1.
	// That is the size of the largest full tree not exceeding count nodes.
	step := 1
	for step <= count {
		step = (2 * step) + 1 // == 2*k - 1
	}
	step /= 2

	stub := &node[T]{right: n}

	// Step 1: Pack the "loose" elements left over.
	// For example, if count == 21 then step == 15 and 6 are left over.
	// After (count - step) == 6 rotations we have 15 (a full tree).
	// This is done first to ensure the leaves fill left-to-right.
	rotateLeft(stub, count-step)

	// Step 2: Pack the full tree and its subtrees.
	left := step
	for left > 1 {
		left /= 2
		rotateLeft(stub, left)
	}
	return stub.right
}

// extract constructs a balanced tree from the given nodes and returns the root
// of the tree. The child pointers of the resulting nodes are updated in place.
// This function does not allocate on the heap. The nodes must be
// sorted and free of duplicates.
func extract[T any](nodes []*node[T]) *node[T] {
	if len(nodes) == 0 {
		return nil
	}
	mid := (len(nodes) - 1) / 2
	root := nodes[mid]
	root.left = extract(nodes[:mid])
	root.right = extract(nodes[mid+1:])
	return root
}

// rewrite composes flatten and extract, returning the rewritten root.
// Costs a single size-element array allocation, plus O(lg size) stack space,
// but does no other allocation.
func rewrite[T any](root *node[T], size int) *node[T] {
	return vineToTree(treeToVine(root), size)
}

// popMinRight removes the smallest node from the right subtree of root,
// modifying the tree in-place and returning the node removed.
// This function panics if root == nil or root.right == nil.
func popMinRight[T any](root *node[T]) *node[T] {
	par, goat := root, root.right
	for goat.left != nil {
		par, goat = goat, goat.left
	}
	if par == root {
		root.right = goat.right
	} else {
		par.left = goat.right
	}
	goat.left = nil
	goat.right = nil
	return goat
}

// inorder visits the subtree under n inorder, calling f until f returns false.
func (n *node[T]) inorder(f func(T) bool) bool {
	for n != nil {
		if ok := n.left.inorder(f); !ok {
			return false
		} else if ok := f(n.X); !ok {
			return false
		}
		n = n.right
	}
	return true
}

// pathTo returns the sequence of nodes beginning at n leading to key, if key
// is present. If key was found, its node is the last element of the path.
func (n *node[T]) pathTo(key T, compare func(a, b T) int) []*node[T] {
	var path []*node[T]
	cur := n
	for cur != nil {
		path = append(path, cur)
		cmp := compare(key, cur.X)
		if cmp < 0 {
			cur = cur.left
		} else if cmp > 0 {
			cur = cur.right
		} else {
			break
		}
	}
	return path
}

// inorderAfter visits the elements of the subtree under n not less than key
// inorder, calling f for each until f returns false.
func (n *node[T]) inorderAfter(key T, compare func(a, b T) int, f func(T) bool) bool {
	// Find the path from the root to key. Any nodes greater than or equal to
	// key must be on or to the right of this path.
	path := n.pathTo(key, compare)
	for i := len(path) - 1; i >= 0; i-- {
		cur := path[i]
		if compare(cur.X, key) < 0 {
			continue
		} else if ok := f(cur.X); !ok {
			return false
		} else if ok := cur.right.inorder(f); !ok {
			return false
		}
	}
	return true
}
