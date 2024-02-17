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
	type rec struct {
		N    int // length of longest subsequence at this junction
		Path []T // longest subsequence on this path
	}
	p := make([]rec, len(as)+1)
	c := make([]rec, len(as)+1)

	// Fill the rows top to bottom, left to right, since the optimization
	// recurrence needs the previous element in the same row, and the same and
	// previous elements in the previous row.
	for j := 1; j <= len(bs); j++ {
		p, c = c, p // swap the double buffer

		// Fill the current row.
		for i := 1; i <= len(as); i++ {
			if as[i-1] == bs[j-1] {
				c[i] = rec{p[i-1].N + 1, append(p[i-1].Path, as[i-1])}
			} else if c[i-1].N >= p[i].N {
				c[i] = rec{c[i-1].N, c[i-1].Path}
			} else {
				c[i] = rec{p[i].N, p[i].Path}
			}
		}
	}

	return c[len(as)].Path
}

// EditOp is the opcode of an edit sequence instruction.
type EditOp byte

const (
	OpDelete  EditOp = '-' // Delete items from lhs
	OpInsert  EditOp = '+' // Insert items from rhs
	OpCopy    EditOp = '=' // Copy elements from lhs
	OpReplace EditOp = 'x' // Replace with items from rhs
)

// Edit is an edit operation transforming specified as part of a diff.
// Each edit refers to a specific span of one of the inputs.
type Edit struct {
	Op EditOp // the diff operation to apply at the current offset

	// N specifies the number of inputs affected by the operation.
	N int

	// X specifies an additionl argument affected by the operation:
	//
	// For OpDelete and OpCopy, X is not used and will be 0.
	// For OpInsert and OpReplace, X specifies a starting offset in rhs from
	// which values are to be copied.
	X int
}

func (e Edit) String() string {
	if e.Op == OpInsert || e.Op == OpReplace {
		return fmt.Sprintf("%c%d:%d", e.Op, e.N, e.X)
	}
	return fmt.Sprintf("%c%d", e.Op, e.N)
}

// EditScript computes a minimal-length sequence of Edit operations that will
// transform lhs into rhs. The result is empty if lhs == rhs.
//
// This implementation takes O(mn) time and O(min(m, n)) space to compute a
// longest common subsequence, plus overhead of O(m+n) to construct the edit
// sequence from the LCS.
//
// An edit sequence is processed in order starting at offset 0 of lhs. Items
// are sent to the output according to the following rules.
//
// For each element e of the edit script, if e.Op is:
//
//   - OpDelete: advance the offset by e.N (no output)
//   - OpInsert: output e.N elements from rhs at position e.X
//   - OpCopy: output e.N elements from lhs at the current offset, and advance
//     the offset by e.N positions
//   - OpReplace: output e.N elements from rhs at position e.X, and advance the
//     offset by e.N positions
//
// After all edits are processed, output any remaining elements of lhs.  This
// completes the processing of the script.
func EditScript[T comparable, Slice ~[]T](lhs, rhs Slice) []Edit {
	lcs := LCS(lhs, rhs)

	// To construct the edit sequence, i scans forward through lcs.
	// For each i, we find the unclaimed elements of lhs and rhs prior to the
	// occurrence of lcs[i].
	//
	// Elements of lhs before lcs[i] must be deleted from the result.
	// Elements of rhs before lcs[i] must be inserted into the result.
	// Elements equal to lcs members are copied.
	//
	// However, whenever we have deletes followed immediately by inserts, the
	// net effect is to "replace" some or all of the deleted items with the
	// inserted ones. We represent this case explicitly with a replace edit.
	lpos, rpos, i := 0, 0, 0

	var out []Edit
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

		// Add exchanges for overlapping delete + insert pairs.
		if x := min(lend-lpos, rend-rpos); x > 0 {
			out = append(out, Edit{Op: OpReplace, N: x, X: rpos})
			lpos += x
			rpos += x
		}

		// Record any remaining unpaired deletions and insertions.
		// Note deletions need to go first.
		if lend > lpos {
			out = append(out, Edit{Op: OpDelete, N: lend - lpos})
		}
		if rend > rpos {
			out = append(out, Edit{Op: OpInsert, N: rend - rpos, X: rpos})
		}

		lpos, rpos = lend, rend

		// Reaching here, lhs[lpos] == rhs[rpos] == lcs[i].
		// Count how many elements are equal and copy them.
		m := 1
		for i+m < len(lcs) && lhs[lpos+m] == rhs[rpos+m] {
			m++
		}
		i += m
		lpos += m
		rpos += m
		out = append(out, Edit{Op: OpCopy, N: m})
	}

	// Add exchanges for overlapping delete + insert pairs.
	if x := min(len(lhs)-lpos, len(rhs)-rpos); x > 0 {
		out = append(out, Edit{Op: OpReplace, N: x, X: rpos})
		lpos += x
		rpos += x
	}
	// Delete any leftover elements of lhs.
	if n := len(lhs) - lpos; n > 0 {
		out = append(out, Edit{Op: OpDelete, N: n})
	}
	// Insert any leftover elements of rhs.
	if n := len(rhs) - rpos; n > 0 {
		out = append(out, Edit{Op: OpInsert, N: n, X: rpos})
	}
	if n := len(out); n > 0 && out[n-1].Op == OpCopy {
		return out[:n-1]
	}
	return out
}
