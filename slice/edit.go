package slice

import (
	"fmt"
)

// LCS computes a longest common subsequence of as and bs.
func LCS[T comparable, Slice ~[]T](as, bs Slice) Slice {
	return lcsRec(as, bs, make(map[[2]int]Slice))
}

func lcsRec[T comparable, Slice ~[]T](as, bs Slice, m map[[2]int]Slice) Slice {
	if len(as) == 0 || len(bs) == 0 {
		return nil
	}
	na, nb := len(as), len(bs)
	if v, ok := m[[2]int{na, nb}]; ok {
		return v
	}
	if as[na-1] == bs[nb-1] {
		ans := append(lcsRec(as[:na-1], bs[:nb-1], m), as[na-1])
		m[[2]int{na, nb}] = ans
		return ans
	}

	lhs := lcsRec(as[:na-1], bs, m)
	rhs := lcsRec(as, bs[:nb-1], m)
	if len(lhs) >= len(rhs) {
		m[[2]int{na, nb}] = lhs
		return lhs
	} else {
		m[[2]int{na, nb}] = rhs
		return rhs
	}
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
// transform lhs into rhs when applied.
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
