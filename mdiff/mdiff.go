// Package mdiff supports creating textual diffs.
//
// To create a diff between two slices of strings, call:
//
//	diff := mdiff.New(lhs, rhs)
//
// The diff.Chunks field contains the disjoint chunks of the input where edits
// have been applied. The complete edit sequence is in diff.Edits.
//
// By default, a diff does not include any context lines. To add up to n lines
// of context, call:
//
//	diff.AddContext(n)
//
// This adds additional edits to the head and tail of each chunk containing the
// context lines, if any, that were found in the input.  Adding context may
// cause chunks to overlap. To remove the overlap, call:
//
//	diff.Unify()
//
// This modifies the diff in-place to merge adjacent and overlapping chunks, so
// that their contexts are not repeated.
//
// These operations can be chained to produce a (unified) diff with context:
//
//	diff := mdiff.New(lhs, rhs).AddContext(3).Unify()
//
// # Output
//
// To write a diff in textual format, use one of the formatting functions.  For
// example, use [Format] to write an old-style Unix diff output to stdout:
//
//	mdiff.Format(os.Stdout, diff, nil)
//
// The [FormatContext] and [FormatUnified] functions allow rendering a diff in
// those formats instead. Use [FileInfo] to tell the formatter the names and
// timestamps to use for their file headers:
//
//	mdiff.FormatUnified(os.Stdout, diff, &mdiff.FileInfo{
//	   Left:  "dir/original.go",
//	   Right: "dir/patched.go",
//	})
//
// If the options are omitted, default placeholders are used instead. You can
// also implement your own function using the same signature; the options and
// defaults are exported and usable from another package.
package mdiff

import (
	"github.com/creachadair/mds/slice"
)

// A Diff represents the difference between two slices of strings.
type Diff struct {
	// The left and right inputs. These fields alias the slices passed to New.
	Left, Right []string

	// The diff chunks, in order. If the inputs are identical, Chunks is empty.
	Chunks []*Chunk

	// The sequence of edits, in order, applied to transform Left into Right.
	Edits []Edit
}

// New constructs a Diff from the specified string slices.
// A diff constructed by New has 0 lines of context.
func New(lhs, rhs []string) *Diff {
	es := slice.EditScript(lhs, rhs)

	out := []*Chunk{{LStart: 1, RStart: 1, LEnd: 1, REnd: 1}}
	cur := out[0]

	lcur, rcur := 1, 1
	addl := func(n int) { lcur += n; cur.LEnd += n }
	addr := func(n int) { rcur += n; cur.REnd += n }
	for _, e := range es {
		// If there is a gap after the previous chunk, start a new one, unless
		// the previous chunk is empty in which case take it over.
		if lcur > cur.LEnd || rcur > cur.REnd {
			if cur.LEnd != cur.LStart || cur.REnd != cur.RStart {
				cur = new(Chunk)
				out = append(out, cur)
			}
			cur.LStart, cur.LEnd = lcur, lcur
			cur.RStart, cur.REnd = rcur, rcur
		}

		switch e.Op {
		case slice.OpDrop:
			addl(len(e.X))

		case slice.OpCopy:
			addr(len(e.Y))

		case slice.OpReplace:
			addl(len(e.X))
			addr(len(e.Y))

		case slice.OpEmit:
			// Don't count emitted lines against the chunk size,
			// and don't append emits to the edit list.
			lcur += len(e.X)
			rcur += len(e.X)
			continue
		}

		cur.Edits = append(cur.Edits, e)
	}

	// If the last chunk empty, remove it entirely.
	if cur.LEnd == cur.LStart && cur.REnd == cur.RStart {
		out = out[:len(out)-1]
	}

	return &Diff{Left: lhs, Right: rhs, Chunks: out, Edits: es}
}

// AddContext updates d so that each chunk has up to n lines of context before
// and after, to the extent possible. Context lines are those that are equal on
// both sides of the diff. AddContext returns d.
//
// Adding context may result in overlapping chunks. Call Unify to merge
// overlapping chunks.
func (d *Diff) AddContext(n int) *Diff {
	if n <= 0 || len(d.Chunks) == 0 {
		return d
	}

	// Gather lines of context, add Emit operations to each chunk corresponding
	// to those lines, and update the line ranges.
	for _, c := range d.Chunks {
		pre, post := d.findContext(c, n)
		if len(pre) != 0 {
			c.Edits = append([]Edit{{Op: slice.OpEmit, X: pre}}, c.Edits...)
			c.LStart -= len(pre)
			c.RStart -= len(pre)
		}
		if len(post) != 0 {
			c.Edits = append(c.Edits, Edit{Op: slice.OpEmit, X: post})
			c.LEnd += len(post)
			c.REnd += len(post)
		}
	}
	return d
}

// Unify updates d in-place to merge chunks that adjoin or overlap.  For a Diff
// returned by New, this is a no-op; however AddContext may cause chunks to
// abut or to overlap. Unify returns d.
//
// Unify updates the edits of any merged chunks, but does not modify the
// original edit sequence in d.Edits.
func (d *Diff) Unify() *Diff {
	if len(d.Chunks) == 0 {
		return d
	}

	merged := []*Chunk{d.Chunks[0]}

	for _, c := range d.Chunks[1:] {
		last := slice.At(merged, -1)
		// If c does not abut or overlap last, there is nothing to do.
		if c.LStart > last.LEnd {
			merged = append(merged, c)
			continue
		}

		lap := last.LEnd - c.LStart
		end, start := slice.PtrAt(last.Edits, -1), slice.PtrAt(c.Edits, 0)

		// If the chunks strictly overlap, it means one at least chunk has a
		// context edit that runs into the other's span (possibly both).
		//
		// Cut off the overlapping lines from the context edit, and if doing so
		// results in an empty context, remove that edit from the chunk.
		//
		// Note that it is safe to modify the context edits here, as they were
		// constructed explicitly by AddContext and do not share state with the
		// original script edits.
		if lap > 0 {
			if end.Op == slice.OpEmit { // last has post-context
				if lap == len(end.X) { // remove the whole edit
					last.Edits = last.Edits[:len(last.Edits)-1]
					end = slice.PtrAt(last.Edits, -1)
				} else {
					end.X = end.X[:len(end.X)-lap] // drop the overlap
				}
				// Fix up the range.
				last.LEnd -= lap
				last.REnd -= lap
			} else if start.Op == slice.OpEmit { // start has pre-context
				if lap == len(start.X) { // remove the whole edit
					c.Edits = c.Edits[1:]
					start = slice.PtrAt(c.Edits, 0)
				} else {
					start.X = start.X[lap:] // drop the overlap
				}
				// Fix up the range.
				c.LStart += lap
				c.RStart += lap
			}

			// Reaching here, the two must now abut properly.
			if c.LStart < last.LEnd {
				panic("diff: context merge did not work correctly")
			}
		}

		// If both chunks have context edits at the boundary, combine them into a
		// single edit at the end of last. Any overlap has already been fixed.
		if end.Op == slice.OpEmit && start.Op == slice.OpEmit {
			// Move the edited lines from the head of c, and adjust the ends.
			end.X = append(end.X, start.X...)
			last.LEnd += len(start.X)
			last.REnd += len(start.X)

			// Discard the edit from the head of c, and adjust the starts.
			c.Edits = c.Edits[1:]
			c.LStart += len(start.X)
			c.RStart += len(start.X)
		}

		// Merge.
		last.LEnd = c.LEnd
		last.REnd = c.REnd
		last.Edits = append(last.Edits, c.Edits...)
	}
	d.Chunks = merged
	return d
}

// findContext returns slices of up to n strings before and after the specified
// chunk that are equal on the left and right sides of the diff.  Either or
// both slices may be empty if there are no such lines.
func (d *Diff) findContext(c *Chunk, n int) (pre, post []string) {
	lcur, rcur := c.LStart-1, c.RStart-1
	lend, rend := c.LEnd-1, c.REnd-1

	for i := 1; i <= n; i++ {
		p, q := lcur-i, rcur-i
		if p < 0 || q < 0 || d.Left[p] != d.Right[q] {
			break
		}
		pre = append(pre, d.Left[p]) // they are equal, so pick one
	}
	slice.Reverse(pre) // we walked backward from the start

	for i := 0; i < n; i++ {
		p, q := lend+i, rend+i
		if p >= len(d.Left) || q >= len(d.Right) || d.Left[p] != d.Right[q] {
			break
		}
		post = append(post, d.Left[p])
	}
	return
}

// A Chunk is a contiguous region within a diff covered by one or more
// consecutive edit operations.
type Chunk struct {
	// The edits applied within this chunk, in order.
	Edits []Edit

	// The starting and ending lines of this chunk in the left input.
	// Lines are 1-based, and the range includes start but excludes end.
	LStart, LEnd int

	// The starting and ending lines of this chunk in the right input.
	// Lines are 1-based and the range includes start but excludes end.
	RStart, REnd int
}

// Edit is an edit operation on strings.  It is exported here so the caller
// does not need to import slice directly.
type Edit = slice.Edit[string]
