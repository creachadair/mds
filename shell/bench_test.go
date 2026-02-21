// Copyright (c) 2015, Michael J. Fromberger

package shell_test

import (
	"bytes"
	"fmt"
	"math"
	"math/rand"
	"slices"
	"strings"
	"testing"

	"github.com/creachadair/mds/shell"
)

var input string

// Generate a long random string with balanced quotations for perf testing.
func init() {
	var buf bytes.Buffer

	src := rand.NewSource(12345)
	r := rand.New(src)

	const alphabet = "abcdefghijklmnopqrstuvwxyz0123456789   \t\n\n\n"
	pick := func(f float64) byte {
		pos := math.Ceil(f*float64(len(alphabet))) - 1
		return alphabet[int(pos)]
	}

	const inputLen = 100000
	var quote struct {
		q byte
		n int
	}
	for range inputLen {
		if quote.n == 0 {
			q := r.Float64()
			if q < .1 {
				quote.q = '"'
				quote.n = r.Intn(256)
				buf.WriteByte('"')
				continue
			} else if q < .15 {
				quote.q = '\''
				quote.n = r.Intn(256)
				buf.WriteByte('\'')
				continue
			}
		}
		buf.WriteByte(pick(r.Float64()))
		if quote.n > 0 {
			quote.n--
			if quote.n == 0 {
				buf.WriteByte(quote.q)
			}
		}
	}
	input = buf.String()
}

func BenchmarkSplit(b *testing.B) {
	var lens []int
	for i := 1; i < len(input); i *= 4 {
		lens = append(lens, i)
	}
	lens = append(lens, len(input))

	s := shell.NewScanner(nil)
	b.ResetTimer()
	for _, n := range lens {
		b.Run(fmt.Sprintf("len_%d", n), func(b *testing.B) {
			for b.Loop() {
				s.Reset(strings.NewReader(input[:n]))
				s.Each(ignore)
			}
		})
	}
}

func BenchmarkQuote(b *testing.B) {
	const alphabet = "abcdefghijklmnopqrstuvwxyz0123456789   \t\n\n\n"
	src := rand.NewSource(67890)
	r := rand.New(src)

	var buf bytes.Buffer
	for range 100000 {
		switch v := r.Float64(); {
		case v < 0.5:
			buf.WriteByte('\'')
		case v < 0.1:
			buf.WriteByte('"')
		case v < 0.15:
			buf.WriteByte('\\')
		default:
			pos := math.Ceil(r.Float64()*float64(len(alphabet))) - 1
			buf.WriteByte(alphabet[int(pos)])
		}
	}

	input := buf.String()
	parts, _ := shell.Split(input)
	b.Logf("Input length: %d bytes, %d tokens", len(input), len(parts))

	b.Run("Quote", func(b *testing.B) {
		for b.Loop() {
			shell.Quote(input)
		}
	})
	b.Run("Join", func(b *testing.B) {
		for b.Loop() {
			shell.Join(parts)
		}
	})
}

func ignore(string) bool { return true }

func FuzzSplitJoin(f *testing.F) {
	// Check the invariant promised in the documentation, viz., that joining
	// together a set of strings and then splitting them apart results in the
	// same strings.
	//
	// Since we cannot pass a []string to the fuzz engine, encode the input
	// slice as a string using "|" separators, which are split apart inside the
	// test into multiple strings.

	for _, s := range []string{"", "''", "foo", "|foo|", "|foo |'bar baz'|", "foo|bar", "foo bar|baz"} {
		f.Add(s)
	}
	f.Fuzz(func(t *testing.T, in string) {
		parts := strings.Split(in, "|")
		out, _ := shell.Split(shell.Join(parts))
		if !slices.Equal(parts, out) {
			t.Fatalf("Got %+q, want %+q", out, parts)
		}
	})
}
