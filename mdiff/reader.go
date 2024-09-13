package mdiff

import (
	"bufio"
	"errors"
	"fmt"
	"io"
	"strconv"
	"strings"
	"time"

	"github.com/creachadair/mds/slice"
)

// A Patch is the parsed representation of a diff read from text format.
type Patch struct {
	FileInfo *FileInfo // nil if no file header was present
	Chunks   []*Chunk
}

// Format renders a patch in textual format using the specified format function.
func (p *Patch) Format(w io.Writer, f FormatFunc) error { return f(w, p.Chunks, p.FileInfo) }

// ReadGitPatch reads a sequence of unified diff [patches] in the format
// produced by "git diff -p" with default settings. The commit metadata and
// header lines are ignored.
//
// [patches]: https://git-scm.com/docs/diff-format#generate_patch_text_with_p
func ReadGitPatch(r io.Reader) ([]*Patch, error) {
	var out []*Patch

	rd := &diffReader{br: bufio.NewReader(r)}
	for {
		// Look for the "diff --git ..." line.
		if err := scanToPrefix(rd, "diff "); err == io.EOF {
			if len(out) == 0 {
				return nil, errors.New("no patches found")
			}
			return out, nil
		}

		// Skip headers until the "--- " patch header.
		if err := scanToPrefix(rd, "--- "); err == io.EOF {
			return nil, fmt.Errorf("line %d: missing patch header", rd.ln)
		} else if err != nil {
			return nil, fmt.Errorf("line %d: %w", rd.ln, err)
		}

		if err := readUnifiedHeader(rd); err != nil {
			return nil, fmt.Errorf("line %d: read patch header: %w", rd.ln, err)
		} else if rd.fileInfo == nil {
			return nil, fmt.Errorf("line %d: incomplete patch header", rd.ln)
		}

		for {
			err := readUnifiedChunk(rd)
			if err == io.EOF || errors.Is(err, errUnexpectedPrefix) {
				out = append(out, &Patch{Chunks: rd.chunks, FileInfo: rd.fileInfo})
				rd.chunks = nil
				break
			} else if err != nil {
				return nil, err
			}
			// get more
		}
		// An unexpected prefix we will handle on the next iteration.
	}
}

// ReadUnified reads a unified diff patch from r.
func ReadUnified(r io.Reader) (*Patch, error) {
	rd := &diffReader{br: bufio.NewReader(r)}
	if err := readUnified(rd); err != nil {
		return nil, err
	}
	return &Patch{FileInfo: rd.fileInfo, Chunks: rd.chunks}, nil
}

// Read reads an old-style "normal" Unix diff patch from r.
func Read(r io.Reader) (*Patch, error) {
	rd := &diffReader{br: bufio.NewReader(r)}
	if err := readNormal(rd); err != nil {
		return nil, err
	}
	return &Patch{Chunks: rd.chunks}, nil
}

// A diffReader provides common plumbing for reading a text diff.  It keeps
// track of line numbers and one line of lookahead, and accumulates information
// about a file header, if one is present.
type diffReader struct {
	br    *bufio.Reader
	ln    int
	saved *string

	fileInfo *FileInfo
	chunks   []*Chunk
}

// readline reads the next available line from the input, or returns the pushed
// back lookahead line if one is available.
func (r *diffReader) readline() (string, error) {
	if r.saved != nil {
		out := *r.saved
		r.saved = nil
		return out, nil
	}
	line, err := r.br.ReadString('\n')
	if err == io.EOF {
		if line == "" {
			return "", err
		}
		r.ln++
		return line, nil
	} else if err != nil {
		return "", err
	}
	r.ln++
	return strings.TrimSuffix(line, "\n"), nil
}

// unread pushes s on the front of the line buffer. Only one line of pushback
// is supported.
func (r *diffReader) unread(s string) { r.saved = &s }

func parseFileLine(s string, timeFormat ...string) (string, time.Time) {
	name, rest, ok := strings.Cut(s, "\t")
	if ok {
		for _, tf := range timeFormat {
			if ts, err := time.Parse(tf, rest); err == nil {
				return name, ts
			}
		}
	}
	return name, time.Time{}
}

// parseSpan parses a string in the format "xM,N" where x is an arbitrary
// string prefix and M and N are positive integer values.
// If the string has the format "xM' only, parseSpan returns M, 0.
func parseSpan(tag, s string) (lo, hi int, err error) {
	rest, ok := strings.CutPrefix(s, tag)
	if !ok {
		return 0, 0, fmt.Errorf("missing %q prefix", tag)
	}
	lohi := strings.SplitN(rest, ",", 2)
	lo, err = strconv.Atoi(lohi[0])
	if err != nil {
		return 0, 0, err
	}
	if len(lohi) == 1 {
		return lo, 0, nil
	}
	hi, err = strconv.Atoi(lohi[1])
	if err != nil {
		return 0, 0, err
	}
	return lo, hi, nil
}

// readUnified reads a unified diff from r, with an optional header.
func readUnified(r *diffReader) error {
	if err := readUnifiedHeader(r); err != nil {
		return fmt.Errorf("diff header: %w", err)
	}
	for {
		err := readUnifiedChunk(r)
		if err == io.EOF {
			return nil
		} else if err != nil {
			return err
		}
	}
}

// readUnifiedHeader reads a unified diff header from r.
func readUnifiedHeader(r *diffReader) error {
	lline, err := r.readline()
	if err != nil {
		return err
	}
	lhs, ok := strings.CutPrefix(lline, "--- ")
	if !ok {
		r.unread(lline)
		return nil
	}
	var fi FileInfo
	fi.Left, fi.LeftTime = parseFileLine(lhs, TimeFormat)

	rline, err := r.readline()
	if err != nil {
		return err
	}
	rhs, ok := strings.CutPrefix(rline, "+++ ")
	if !ok {
		return errors.New("missing right header")
	}
	fi.Right, fi.RightTime = parseFileLine(rhs, TimeFormat)
	r.fileInfo = &fi
	return nil
}

// readUnifiedChunk reads a single unified diff chunk from r.
func readUnifiedChunk(r *diffReader) error {
	line, err := r.readline()
	if err != nil {
		return err
	}

	// Unified diff headers are "@@ -lspan +rspan @@".
	// But git diff adds additional stuff after the second "@@" to give the
	// reader context. To support that, we relax the format check slightly.
	parts := strings.Fields(line)
	if len(parts) < 4 || parts[0] != "@@" || parts[3] != "@@" {
		return fmt.Errorf("line %d: invalid chunk header %q", r.ln, line)
	}
	llo, lhi, err := parseSpan("-", parts[1])
	if err != nil {
		return fmt.Errorf("line %d: left span: %w", r.ln, err)
	}
	rlo, rhi, err := parseSpan("+", parts[2])
	if err != nil {
		return fmt.Errorf("line %d: right span: %w", r.ln, err)
	}

	ch := &Chunk{LStart: llo, LEnd: llo + lhi, RStart: rlo, REnd: rlo + rhi}
	add := func(op slice.EditOp, text string) {
		if len(ch.Edits) == 0 || ch.Edits[len(ch.Edits)-1].Op != op {
			ch.Edits = append(ch.Edits, Edit{Op: op})
		}
		e := slice.PtrAt(ch.Edits, -1)
		switch op {
		case slice.OpDrop, slice.OpEmit:
			e.X = append(e.X, text)
		case slice.OpCopy:
			e.Y = append(e.Y, text)
		default:
			panic("unexpected operator " + string(op))
		}
	}

nextLine:
	for {
		line, err := r.readline()
		if err == io.EOF {
			break // end of input, end of chunk
		} else if err != nil {
			return err
		} else if line == "" {
			return fmt.Errorf("line %d: unexpected blank line", r.ln)
		}
		switch line[0] {
		case ' ': // context
			add(slice.OpEmit, line[1:])
		case '-': // deletion from lhs
			add(slice.OpDrop, line[1:])
		case '+': // addition from rhs
			add(slice.OpCopy, line[1:])
		case '@': // another diff chunk
			r.unread(line)
			break nextLine
		default:
			// Something else, maybe the start of a new patch or something.
			// Report an error, but save the line and the chunk in case the caller
			// knows what to do about it in context.
			r.unread(line)
			r.chunks = append(r.chunks, ch)
			return fmt.Errorf("line %d: %w %c", r.ln, errUnexpectedPrefix, line[0])
		}
	}
	r.chunks = append(r.chunks, ch)
	return nil
}

// errUnexpectedPrefix is a sentinel error reported by readUnifiedChunk to
// report a line that is not part of a chunk.
var errUnexpectedPrefix = errors.New("unexpected prefix")

// readNormal reads a "normal" Unix diff patch from r.
func readNormal(r *diffReader) error {
	for {
		line, err := r.readline()
		if err == io.EOF {
			return nil
		} else if err != nil {
			return err
		} else if line == "" {
			return fmt.Errorf("line %d: unexpected blank line", r.ln)
		}
		var lspec, cmd, rspec string
		if x, y, ok := strings.Cut(line, "a"); ok { // add lines from rhs
			lspec, cmd, rspec = x, "a", y
		} else if x, y, ok := strings.Cut(line, "c"); ok { // replace lines
			lspec, cmd, rspec = x, "c", y
		} else if x, y, ok := strings.Cut(line, "d"); ok { // delete lines from lhs
			lspec, cmd, rspec = x, "d", y
		} else {
			return fmt.Errorf("line %d: invalid change command %q", r.ln, line)
		}

		llo, lhi, err := parseSpan("", lspec)
		if err != nil {
			return fmt.Errorf("line %d: invalid line range %q: %w", r.ln, lspec, err)
		} else if lhi == 0 {
			lhi = llo // m, 0 → m, m
		}
		lhi++

		rlo, rhi, err := parseSpan("", rspec)
		if err != nil {
			return fmt.Errorf("line %d: invalid line range %q: %w", r.ln, rspec, err)
		} else if rhi == 0 {
			rhi = rlo // n, 0 → n, n
		}
		rhi++

		sln := r.ln
		e, err := readNormalEdit(r)
		if err != nil {
			return err
		}
		switch cmd {
		case "a":
			e.Op = slice.OpCopy
			llo++ // Adds happen after the marked line.
		case "c":
			e.Op = slice.OpReplace
		case "d":
			e.Op = slice.OpDrop
			rlo++ // Deletes happen after the marked line.
		}

		// Cross-check the number of lines reported in the change spec with the
		// number we actually read out of the chunk data.
		if n := rhi - rlo; len(e.Y) != n && (cmd == "a" || cmd == "c") {
			return fmt.Errorf("line %d: add got %d lines, want %d", sln, len(e.Y), n)
		}
		if n := lhi - llo; len(e.X) != n && (cmd == "c" || cmd == "d") {
			return fmt.Errorf("line %d: delete got %d lines, want %d", sln, len(e.X), n)
		}
		r.chunks = append(r.chunks, &Chunk{
			Edits:  []Edit{e},
			LStart: llo, LEnd: lhi,
			RStart: rlo, REnd: rhi,
		})
	}
}

func readNormalEdit(r *diffReader) (Edit, error) {
	var e Edit
	var below bool // whether we have seen a "---" separator
	for {
		line, err := r.readline()
		if err == io.EOF {
			break
		} else if err != nil {
			return Edit{}, err
		}
		if rst, ok := strings.CutPrefix(line, "< "); ok {
			if below || len(e.Y) != 0 {
				return Edit{}, fmt.Errorf("line %d: unexpected delete line %q", r.ln, line)
			}
			e.X = append(e.X, rst)
		} else if rst, ok := strings.CutPrefix(line, "> "); ok {
			if len(e.X) != 0 && !below {
				return Edit{}, fmt.Errorf("line %d: unexpected insert line %q", r.ln, line)
			}
			e.Y = append(e.Y, rst)
		} else if line == "---" {
			if below {
				return Edit{}, fmt.Errorf("line %d: unexpected --- separator", r.ln)
			}
			below = true
		} else {
			r.unread(line)
			break
		}
	}
	return e, nil
}

// scanToPrefix reads forward to a line starting with pfx, and returns nil.
// The matching line is unread so the caller can reuse it.
func scanToPrefix(r *diffReader, prefix string) error {
	for {
		line, err := r.readline()
		if err != nil {
			return err // may be io.EOF, caller will check
		}
		if strings.HasPrefix(line, prefix) {
			r.unread(line)
			return nil
		}
	}
}
