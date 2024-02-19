package slice

import (
	"fmt"
)

// LCS computes a longest common subsequence of as and bs.
//
// This implementation takes O(mn) time and O(P·min(m, n)) space for inputs of
// length m = len(as) and n = len(bs) and longest subsequence length P.
func LCS[T comparable, Slice ~[]T](as, bs Slice) Slice {
	if len(as) == 0 || len(bs) == 0 {
		return nil
	}

	// We maintain two rows of the optimization matrix, (p)revious and
	// (c)urrent. Rows are positions in bs and columns are positions in as.
	// The rows are extended by one position to get rid of the special case at
	// the beginning of the sequence (we use 1..n instead of 0..n-1).

	// Size the buffers based on the smaller input, since order does not matter.
	// This lets us use less storage with no time penalty.
	if len(bs) < len(as) {
		as, bs = bs, as
	}
	p := make([]Slice, len(as)+1)
	c := make([]Slice, len(as)+1)

	// Fill the rows top to bottom, left to right, since the optimization
	// recurrence needs the previous element in the same row, and the same and
	// previous elements in the previous row.
	for j := 1; j <= len(bs); j++ {
		p, c = c, p // swap the double buffer

		// Fill the current row.
		for i := 1; i <= len(as); i++ {
			if as[i-1] == bs[j-1] {
				c[i] = append(p[i-1], as[i-1])
			} else if len(c[i-1]) >= len(p[i]) {
				c[i] = c[i-1]
			} else {
				c[i] = p[i]
			}
		}
	}

	return c[len(as)]
}

// EditOp is the opcode of an edit sequence instruction.
type EditOp byte

const (
	OpDrop    EditOp = '-' // Drop items from lhs
	OpEmit    EditOp = '=' // Emit elements from lhs
	OpCopy    EditOp = '+' // Copy items from rhs
	OpReplace EditOp = '!' // Replace with items from rhs (== Drop+Copy)
)

// Edit is an edit operation transforming specified as part of a diff.
// Each edit refers to a specific span of one of the inputs.
type Edit[T comparable] struct {
	Op EditOp // the diff operation to apply at the current offset

	// X specifies the elements of lhs affected by the edit.
	// For OpDrop and OpReplace it is the elements to be dropped.
	// For OpEmit its the elements of to be emitted.
	// For OpCopy it is empty.
	X []T

	// Y specifies the elements of rhs affected by the edit.
	// For OpDrop and OpEmit it is empty.
	// For OpCopy and OpReplace it is the elements to be copied.
	Y []T
}

func (e Edit[T]) String() string {
	switch e.Op {
	case OpCopy:
		return fmt.Sprintf("%c%v", e.Op, e.Y)
	case OpReplace:
		x, y := fmt.Sprint(e.X), fmt.Sprint(e.Y)
		return fmt.Sprintf("%c[%s:%s]", e.Op, x[1:len(x)-1], y[1:len(y)-1])
	case OpDrop:
		return fmt.Sprintf("%c%d", e.Op, len(e.X))
	case OpEmit:
		return fmt.Sprintf("%c%v", e.Op, e.X)
	}
	return fmt.Sprintf("!%c[INVALID]", e.Op)
}

// EditScript computes a minimal-length sequence of Edit operations that will
// transform lhs into rhs. The result is empty if lhs == rhs. The slices stored
// in returned edit operations share storage with the inputs lhs and rhs.
//
// This implementation takes O(mn) time and O(P·min(m, n)) space to compute a
// longest common subsequence, plus overhead of O(m+n) time and space to
// construct the edit sequence from the LCS.
//
// An edit sequence is processed in order starting at offset 0 of lhs. Items
// are sent to the output according to the following rules.
//
// For each element e of the edit script, if e.Op is:
//
//   - OpDrop: advance the offset by e.N without emitting any output.
//
//   - OpEmit: output e.N elements from the current offset in lhs and advance
//     the offset by e.N positions.
//
//   - OpCopy: output e.N elements from position e.X of rhs.
//
//   - OpReplace: output e.N elements from position e.X of rhs, and advance
//     offset by e.N positions (a combination of Drop and Copy).
//
// After all edits are processed, output any remaining elements of lhs.  This
// completes the processing of the script.
func EditScript[T comparable, Slice ~[]T](lhs, rhs Slice) []Edit[T] {
	lcs := LCS(lhs, rhs)

	// To construct the edit sequence, i scans forward through lcs.
	// For each i, we find the unclaimed elements of lhs and rhs prior to the
	// occurrence of lcs[i].
	//
	// Elements of lhs before lcs[i] must be removed from the result.
	// Elements of rhs before lcs[i] must be added to the result.
	// Elements equal to lcs members are preserved as-written.
	//
	// However, whenever we have deletes followed immediately by inserts, the
	// net effect is to "replace" some or all of the deleted items with the
	// inserted ones. We represent this case explicitly with a replace edit.
	lpos, rpos, i := 0, 0, 0

	var out []Edit[T]
	for i < len(lcs) {
		// Count the numbers of elements of lhs and rhs prior to the first match.
		lend := lpos
		for lhs[lend] != lcs[i] {
			lend++
		}
		rend := rpos
		for rhs[rend] != lcs[i] {
			rend++
		}

		// If we have at least as many insertions as discards, combine them into
		// a single replace instruction.
		if n := lend - lpos; n > 0 && n <= rend-rpos {
			out = append(out, Edit[T]{Op: OpReplace, X: lhs[lpos:lend], Y: rhs[rpos : rpos+n]})
			rpos += n
		} else if lend > lpos {
			// Record drops (there may be none).
			out = append(out, Edit[T]{Op: OpDrop, X: lhs[lpos:lend]})
		}
		// Record copies (there may be none).
		if rend > rpos {
			out = append(out, Edit[T]{Op: OpCopy, Y: rhs[rpos:rend]})
		}

		lpos, rpos = lend, rend

		// Reaching here, lhs[lpos] == rhs[rpos] == lcs[i].
		// Count how many elements are equal and copy them.
		m := 1
		for i+m < len(lcs) && lhs[lpos+m] == rhs[rpos+m] {
			m++
		}
		out = append(out, Edit[T]{Op: OpEmit, X: lhs[lpos : lpos+m]})
		i += m
		lpos += m
		rpos += m
	}

	if n := len(lhs) - lpos; n > 0 && n <= len(rhs)-rpos {
		out = append(out, Edit[T]{Op: OpReplace, X: lhs[lpos:], Y: rhs[rpos : rpos+n]})
		rpos += n
	} else if n > 0 {
		// Drop any leftover elements of lhs.
		out = append(out, Edit[T]{Op: OpDrop, X: lhs[lpos:]})
	}
	// Copy any leftover elements of rhs.
	if len(rhs)-rpos > 0 {
		out = append(out, Edit[T]{Op: OpCopy, Y: rhs[rpos:]})
	}

	// As a special case, if the whole edit is a single emit, drop it so that
	// equal elements have an empty script.
	if len(out) == 1 && out[0].Op == OpEmit {
		return nil
	}
	return out
}
