package mdiff

import (
	"cmp"
	"fmt"
	"io"
	"strconv"
	"time"

	"github.com/creachadair/mds/slice"
)

// FormatFunc is a function that renders diff chunks as text to an io.Writer.
//
// A FormatFunc should accept a nil info pointer, and should skip or supply
// default values for missing fields.
type FormatFunc func(w io.Writer, ch []*Chunk, fi *FileInfo) error

// TimeFormat is the default format string used to render timestamps in context
// and unified diff outputs. It is based on the RFC 2822 time format.
const TimeFormat = "2006-01-02 15:04:05.999999 -0700"

// FileInfo specifies file metadata to use when formatting a diff.
type FileInfo struct {
	// Left is the filename to use for the left-hand input.
	Left string

	// Right is the filename to use for the right-hand argument.
	Right string

	// LeftTime is the timestamp to use for the left-hand input.
	LeftTime time.Time

	// RightTime is the timestamp to use for the right-hand input.
	RightTime time.Time

	// TimeFormat specifies the time format to use for timestamps.
	// Any format string accepted by time.Format is permitted.
	// If omitted, it uses the TimeFormat constant.
	TimeFormat string
}

// Unified is a [FormatFunc] that renders ch in the [unified diff] format
// introduced by GNU diff. If fi == nil, the file header is omitted.
//
// [unified diff]: https://www.gnu.org/software/diffutils/manual/html_node/Unified-Format.html
func Unified(w io.Writer, ch []*Chunk, fi *FileInfo) error {
	if len(ch) == 0 {
		return nil
	}
	if fi != nil {
		fmtFileHeader(w, "--- ", cmp.Or(fi.Left, "a"), fi.LeftTime, cmp.Or(fi.TimeFormat, TimeFormat))
		fmtFileHeader(w, "+++ ", cmp.Or(fi.Right, "b"), fi.RightTime, cmp.Or(fi.TimeFormat, TimeFormat))
	}
	for _, c := range ch {
		fmt.Fprintln(w, "@@", uspan("-", c.LStart, c.LEnd), uspan("+", c.RStart, c.REnd), "@@")
		for _, e := range c.Edits {
			switch e.Op {
			case slice.OpDrop:
				writeLines(w, "-", e.X)
			case slice.OpEmit:
				writeLines(w, " ", e.X)
			case slice.OpCopy:
				writeLines(w, "+", e.Y)
			case slice.OpReplace:
				writeLines(w, "-", e.X)
				writeLines(w, "+", e.Y)
			}
		}
	}
	return nil
}

func fmtFileHeader(w io.Writer, prefix, name string, ts time.Time, tfmt string) {
	fmt.Fprint(w, prefix, name)
	if !ts.IsZero() {
		fmt.Fprint(w, "\t", ts.Format(tfmt))
	}
	fmt.Fprintln(w)
}

// Context is a [FormatFunc] that renders ch in the [context diff] format
// introduced by BSD diff. If fi == nil, the file header is omitted.
//
// [context diff]: https://www.gnu.org/software/diffutils/manual/html_node/Context-Format.html
func Context(w io.Writer, ch []*Chunk, fi *FileInfo) error {
	if len(ch) == 0 {
		return nil
	}
	if fi != nil {
		fmtFileHeader(w, "*** ", cmp.Or(fi.Left, "a"), fi.LeftTime, cmp.Or(fi.TimeFormat, TimeFormat))
		fmtFileHeader(w, "--- ", cmp.Or(fi.Right, "b"), fi.RightTime, cmp.Or(fi.TimeFormat, TimeFormat))
	}
	for _, c := range ch {
		// Why 15 stars? I can't say. Berkeley just liked it better that way.
		fmt.Fprintln(w, "***************")
		fmt.Fprintf(w, "*** %s ****\n", dspan(c.LStart, c.LEnd))
		if hasRelevantEdits(c.Edits, slice.OpDrop) {
			for _, e := range c.Edits {
				switch e.Op {
				case slice.OpDrop:
					writeLines(w, "- ", e.X)
				case slice.OpEmit:
					writeLines(w, "  ", e.X)
				case slice.OpReplace:
					writeLines(w, "! ", e.X)
				}
			}
		}
		fmt.Fprintf(w, "--- %s ----\n", dspan(c.RStart, c.REnd))
		if hasRelevantEdits(c.Edits, slice.OpCopy) {
			for _, e := range c.Edits {
				switch e.Op {
				case slice.OpCopy:
					writeLines(w, "+ ", e.Y)
				case slice.OpEmit:
					writeLines(w, "  ", e.X)
				case slice.OpReplace:
					writeLines(w, "! ", e.Y)
				}
			}
		}
	}
	return nil
}

// Normal is a [FormatFunc] that renders ch in the "normal" [Unix diff] format.
// This format does not include a file header, so the FileInfo is ignored.
//
// [Unix diff]: https://www.gnu.org/software/diffutils/manual/html_node/Detailed-Normal.html
func Normal(w io.Writer, ch []*Chunk, _ *FileInfo) error {
	for _, c := range ch {
		lpos, rpos := c.LStart, c.RStart
		for _, e := range c.Edits {
			switch e.Op {
			case slice.OpDrop:
				// Diff considers deletions to happen AFTER the previous line rather
				// than on the current one.
				fmt.Fprintf(w, "%sd%d\n", dspan(lpos, lpos+len(e.X)), rpos-1)
				writeLines(w, "< ", e.X)
				lpos += len(e.X)

			case slice.OpEmit:
				lpos += len(e.X)
				rpos += len(e.X)

			case slice.OpCopy:
				// Diff considers insertions to happen AFTER the previons line rather
				// than on the current one.
				fmt.Fprintf(w, "%da%s\n", lpos-1, dspan(rpos, rpos+len(e.Y)))
				writeLines(w, "> ", e.Y)
				rpos += len(e.Y)

			case slice.OpReplace:
				fmt.Fprintf(w, "%sc%s\n", dspan(lpos, lpos+len(e.X)), dspan(rpos, rpos+len(e.Y)))
				writeLines(w, "< ", e.X)
				fmt.Fprintln(w, "---")
				writeLines(w, "> ", e.Y)
				lpos += len(e.X)
				rpos += len(e.Y)
			}
		}
	}
	return nil
}

// dspan formats the range start, end as a diff span.
func dspan(start, end int) string {
	if end-start == 1 {
		return strconv.Itoa(start)
	}
	return fmt.Sprintf("%d,%d", start, end-1)
}

// uspan formats the range start, end as a unified diff span.
func uspan(side string, start, end int) string {
	if end-start == 1 {
		return side + strconv.Itoa(start)
	}
	return fmt.Sprintf("%s%d,%d", side, start, end-start)
}

func writeLines(w io.Writer, pfx string, lines []string) {
	for _, line := range lines {
		fmt.Fprint(w, pfx, line, "\n")
	}
}

// hasRelevantEdits reports whether es contains at least one edit with either
// the specified opcode or slice.OpReplace.
func hasRelevantEdits(es []Edit, op slice.EditOp) bool {
	for _, e := range es {
		if e.Op == op || e.Op == slice.OpReplace {
			return true
		}
	}
	return false
}
