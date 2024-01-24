package mtest_test

import (
	"testing"

	"github.com/creachadair/mds/mtest"
)

func TestMustPanic(t *testing.T) {
	v := mtest.MustPanic(t, func() { panic("pass") })
	t.Logf("Panic reported: %v", v)
}

func TestMustPanicf(t *testing.T) {
	v := mtest.MustPanicf(t, func() { panic("pass") }, "bad things: %d", 5)
	t.Logf("Panic reported: %v", v)
}
