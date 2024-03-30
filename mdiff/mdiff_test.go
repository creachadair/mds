package mdiff_test

import (
	"bytes"
	"fmt"
	"strings"
	"testing"
	"time"

	"github.com/creachadair/mds/mdiff"
	"github.com/creachadair/mds/mstr"
	"github.com/creachadair/mds/slice"
	gocmp "github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"

	_ "embed"
)

// Type satisfaction checks.
var (
	_ mdiff.FormatFunc = mdiff.Format
	_ mdiff.FormatFunc = mdiff.FormatContext
	_ mdiff.FormatFunc = mdiff.FormatUnified
)

var (
	// These example input files were copied from
	//
	// https://www.gnu.org/software/diffutils/manual/html_node/Sample-diff-Input.html

	//go:embed testdata/lhs.txt
	lhs string

	//go:embed testdata/rhs.txt
	rhs string

	// The comparison files were generated by running the diff command on the
	// above input files. The timestamps and paths recorded in the output are
	// relative to the package directory.

	// Old-style diff output: diff testdata/lhs.txt testdata/rhs.txt
	//go:embed testdata/odiff.txt
	odiff string

	// Unified diff output (3 lines): diff -u testdata/lhs.txt testdata/rhs.txt
	//go:embed testdata/udiff.txt
	udiff string

	// Context diff output (3 lines): diff -c testdata/lhs.txt testdata/rhs.txt
	//go:embed testdata/cdiff.txt
	cdiff string

	lhsLines = mstr.Lines(lhs)
	rhsLines = mstr.Lines(rhs)
)

func TestDiff(t *testing.T) {
	t.Run("Empty", func(t *testing.T) {
		d := mdiff.New(nil, nil)
		if diff := gocmp.Diff(d, &mdiff.Diff{}, cmpopts.EquateEmpty()); diff != "" {
			t.Errorf("Diff of empty (-got, +want):\n%s", diff)
		}
	})

	t.Run("Equal", func(t *testing.T) {
		input := strings.Fields("no king rules forever my son")
		if diff := gocmp.Diff(mdiff.New(input, input), &mdiff.Diff{
			Left:  input,
			Right: input,
		}, cmpopts.EquateEmpty()); diff != "" {
			t.Errorf("Diff of equal (-got, +want):\n%s", diff)
		}
	})
}

func TestNoAlias(t *testing.T) {
	// The documentation promises that adding context and unifying does not
	// disturb the original edit sequence.
	d := mdiff.New(lhsLines, rhsLines)
	logDiff(t, d)

	before := fmt.Sprint(d.Edits)
	d.AddContext(6).Unify()
	if fmt.Sprint(d.Edits) != before {
		t.Errorf("Edits were altered:\n got %v,\nwant %s", d.Edits, before)
	}
}

func TestFormat(t *testing.T) {

	t.Run("Normal", func(t *testing.T) {
		d := mdiff.New(lhsLines, rhsLines)

		var buf bytes.Buffer
		mdiff.Format(&buf, d, nil)
		if got := buf.String(); got != odiff {
			t.Errorf("Normal diff disagrees with testdata.\nGot:\n%s\n\nWant:\n%s", got, odiff)
		}
	})

	t.Run("Context", func(t *testing.T) {
		d := mdiff.New(lhsLines, rhsLines).AddContext(3).Unify()
		logDiff(t, d)

		// This is the timestamp recorded for the testdata file. If you edit the
		// file, you may need to update this.
		when := time.Date(2024, 3, 16, 18, 53, 15, 123450000, time.UTC)

		var buf bytes.Buffer
		mdiff.FormatContext(&buf, d, &mdiff.FileInfo{
			Left:       "testdata/lhs.txt",
			LeftTime:   when,
			Right:      "testdata/rhs.txt",
			RightTime:  when,
			TimeFormat: time.ANSIC,
		})
		if got := buf.String(); got != cdiff {
			t.Errorf("Context diff disagrees with testdata.\nGot:\n%s\n\nWant:\n%s", got, cdiff)
		}
	})

	t.Run("Unified", func(t *testing.T) {
		d := mdiff.New(lhsLines, rhsLines).AddContext(3).Unify()

		// This is the timestamp recorded for the testdata file.  If you edit the
		// file, you may need to update this.
		when := time.Date(2024, 3, 16, 17, 47, 40, 123450000, time.UTC)

		var buf bytes.Buffer
		mdiff.FormatUnified(&buf, d, &mdiff.FileInfo{
			Left:      "testdata/lhs.txt",
			LeftTime:  when,
			Right:     "testdata/rhs.txt",
			RightTime: when,
		})
		if got := buf.String(); got != udiff {
			t.Errorf("Unified diff disagrees with testdata.\nGot:\n%s\n\nWant:\n%s", got, udiff)
		}
	})

	t.Run("Unified/NoTime", func(t *testing.T) {
		d := mdiff.New(lhsLines, rhsLines).AddContext(3).Unify()

		var buf bytes.Buffer
		mdiff.FormatUnified(&buf, d, &mdiff.FileInfo{Left: "a/fuzzy", Right: "b/wuzzy"})
		lines := mstr.Lines(buf.String())
		if diff := gocmp.Diff(slice.Head(lines, 2), []string{
			"--- a/fuzzy",
			"+++ b/wuzzy",
		}); diff != "" {
			t.Errorf("Header (-got, +want):\n%s", diff)
		}
		t.Logf("Diff:\n%s\n...", strings.Join(slice.Head(lines, 5), "\n"))
	})

	t.Run("Empty/Normal", func(t *testing.T) {
		empty := mdiff.New(lhsLines, lhsLines)
		var buf bytes.Buffer
		mdiff.Format(&buf, empty, nil)
		if got := buf.String(); got != "" {
			t.Errorf("Format: got:\n%s\nwant empty", got)
		}
	})

	t.Run("Empty/Context", func(t *testing.T) {
		empty := mdiff.New(lhsLines, lhsLines).AddContext(3).Unify()
		var buf bytes.Buffer
		mdiff.FormatContext(&buf, empty, nil)
		if got := buf.String(); got != "" {
			t.Errorf("Format: got:\n%s\nwant empty", got)
		}
	})

	t.Run("Empty/Unified", func(t *testing.T) {
		empty := mdiff.New(lhsLines, lhsLines).AddContext(3).Unify()
		var buf bytes.Buffer
		mdiff.FormatUnified(&buf, empty, nil)
		if got := buf.String(); got != "" {
			t.Errorf("Format: got:\n%s\nwant empty", got)
		}
	})
}

func logDiff(t *testing.T, d *mdiff.Diff) {
	t.Helper()
	t.Logf("Input left: %d lines, right: %d lines; diff has %d edits",
		len(d.Left), len(d.Right), len(d.Edits))

	t.Log("Original edits:")
	for i, e := range d.Edits {
		t.Logf("%d: %v", i+1, e)
	}

	t.Log("Chunks:")
	for i, c := range d.Chunks {
		t.Logf("%d: %d edits -%d,%d +%d,%d", i+1, len(c.Edits),
			c.LStart, c.LEnd-c.LStart, c.RStart, c.REnd-c.RStart)
		for j, e := range c.Edits {
			t.Logf(" E %d: %+v\n", j+1, e)
		}
	}
}
