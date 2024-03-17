package mdiff

import (
	"fmt"
	"io"
	"strconv"
	"time"

	"github.com/creachadair/mds/slice"
	"github.com/creachadair/mds/value"
)

// FormatFunc is a function that renders a Diff as text to an io.Writer.
//
// A FormatFunc should accept a nil options pointer.  Use the Get method to
// obtain the default values described under [FormatOptions].
type FormatFunc func(w io.Writer, d *Diff, o *FormatOptions) error

// TimeFormat is the default format string used to render timestamps in context
// and unified diff outputs. It is based on the RFC 2822 time format.
const TimeFormat = "2006-01-02 15:04:05.999999 -0700"

// FormatOptions are optional settings to format a diff.  A nil pointer is
// ready for use and provides defaults as described.
type FormatOptions struct {
	// OmitHeader, if true, instructs the formatter to skip the header section
	// giving filenames and timestamps.
	OmitHeader bool

	// Left is the filename to use for the left-hand input.
	// If omitted, it uses "a".
	Left string

	// Right is the filename to use for the right-hand argument.
	// If omitted, it uses "b".
	Right string

	// LeftTime is the timestamp to use for the left-hand input.
	// If zero it uses the current wall-clock time.
	LeftTime time.Time

	// RightTime is the timestamp to use for the right-hand input.
	// If zero it uses the current wall-clock time.
	RightTime time.Time

	// TimeFormat specifies the time format to use for timestamps.
	// Any format string accepted by time.Format is permitted.
	// If omitted, it uses the TimeFormat constant.
	TimeFormat string
}

// Get returns an options value in which any unspecified fields from o are
// populated with defaults. If o == nil, defaults are supplied for all fields.
func (o *FormatOptions) Get() FormatOptions {
	out := value.At(o)
	if out.Left == "" {
		out.Left = "a"
	}
	if out.Right == "" {
		out.Right = "b"
	}
	if out.TimeFormat == "" {
		out.TimeFormat = TimeFormat
	}
	if zl, zr := out.LeftTime.IsZero(), out.RightTime.IsZero(); zl || zr {
		now := time.Now()
		if zl {
			out.LeftTime = now
		}
		if zr {
			out.RightTime = now
		}
	}
	return out
}

// FormatUnified is a [FormatFunc] that renders d in the [unified diff] format
// introduced by GNU diff.
//
// [unified diff]: https://www.gnu.org/software/diffutils/manual/html_node/Unified-Format.html
func FormatUnified(w io.Writer, d *Diff, o *FormatOptions) error {
	if len(d.Chunks) == 0 {
		return nil
	}
	opts := o.Get()
	if !opts.OmitHeader {
		fmt.Fprintf(w, "--- %s\t%s\n", opts.Left, opts.LeftTime.Format(opts.TimeFormat))
		fmt.Fprintf(w, "+++ %s\t%s\n", opts.Right, opts.RightTime.Format(opts.TimeFormat))
	}
	for _, c := range d.Chunks {
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

// FormatContext is a [FormatFunc] that renders d in the [context diff] format
// introduced by BSD diff.
//
// [context diff]: https://www.gnu.org/software/diffutils/manual/html_node/Context-Format.html
func FormatContext(w io.Writer, d *Diff, o *FormatOptions) error {
	if len(d.Chunks) == 0 {
		return nil
	}
	opts := o.Get()
	if !opts.OmitHeader {
		fmt.Fprintf(w, "*** %s\t%s\n", opts.Left, opts.LeftTime.Format(opts.TimeFormat))
		fmt.Fprintf(w, "--- %s\t%s\n", opts.Right, opts.RightTime.Format(opts.TimeFormat))
	}
	for _, c := range d.Chunks {
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

// Format is a [FormatFunc] that renders d in the "normal" Unix diff format.
// The options are not used by this format, and context is ignored.
func Format(w io.Writer, d *Diff, _ *FormatOptions) error {
	lpos, rpos := 1, 1
	for _, e := range d.Edits {
		switch e.Op {
		case slice.OpDrop:
			// Diff considers deletions to happen AFTER the previous line rather
			// than on the current one.
			epos := rpos
			if rpos == 1 {
				epos--
			}
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
