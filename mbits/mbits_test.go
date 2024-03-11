package mbits_test

import (
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
