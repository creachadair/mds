package mbits_test

import (
	"fmt"
	"strings"
	"testing"

	"github.com/creachadair/mds/mbits"
)

func isZero(data []byte) bool {
	for _, b := range data {
		if b != 0 {
			return false
		}
	}
	return true
}

func TestZero(t *testing.T) {
	for _, s := range []string{
		"",
		"\x00",
		"\x00\x00\x00\x00\x00\x00\x00",
		"abcd\x00\x00efghij\x00jklmnopqrstuvwxyz",
		"abcdefgh",
		"abcdefgh1",
		"abcdefgh12",
		"abcdefgh123",
		"abcdefgh1234",
		"abcdefgh12345",
		"abcdefgh123456",
		"abcdefgh1234567",
		"abcdefgh12345678",
		"abcdefgh123456789",
		strings.Repeat("\x00", 1000),
		strings.Repeat("\xff", 1000),
		strings.Repeat("\x00\xff\x01", 1003),
	} {
		in := []byte(s)
		mbits.Zero(in)
		if !isZero(in) {
			t.Errorf("Zero %q did not work", s)
		}
	}
}

func TestLeadingZeroes(t *testing.T) {
	for _, nb := range []int{5, 16, 43, 100, 128} {
		t.Run(fmt.Sprintf("Buf%d", nb), func(t *testing.T) {
			buf := make([]byte, nb)
			if got := mbits.LeadingZeroes(buf); got != nb {
				t.Errorf("Got %d leading zeroes, want %d", got, nb)
			}

			// Test every possible offset.
			for i := range len(buf) {
				buf[i] = 1
				if got := mbits.LeadingZeroes(buf); got != i {
					t.Errorf("Got %d leading zeroes, want %d", got, i)
				}
				buf[i] = 0
			}
		})
	}
}

func TestTrailingZeroes(t *testing.T) {
	for _, nb := range []int{5, 16, 43, 100, 128} {
		t.Run(fmt.Sprintf("Buf%d", nb), func(t *testing.T) {
			buf := make([]byte, nb)
			if got := mbits.TrailingZeroes(buf); got != nb {
				t.Errorf("Got %d trailing zeroes, want %d", got, nb)
			}

			// Test every possible offset.
			for i := range len(buf) {
				pos := len(buf) - i - 1
				buf[pos] = 1
				if got := mbits.TrailingZeroes(buf); got != i {
					t.Errorf("Got %d trailing zeroes, want %d", got, i)
				}
				buf[pos] = 0
			}
		})
	}
}
