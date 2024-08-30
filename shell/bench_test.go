// Copyright (c) 2015, Michael J. Fromberger

package shell_test

import (
	"bytes"
	"fmt"
	"math"
	"math/rand"
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

	b.ResetTimer()
	for _, n := range lens {
		b.Run(fmt.Sprintf("len_%d", n), func(b *testing.B) {
			for range b.N {
				shell.NewScanner(strings.NewReader(input[:n])).Each(ignore)
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
	b.ResetTimer()

	for range b.N {
		shell.Quote(input)
	}
}

func ignore(string) bool { return true }
