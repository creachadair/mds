package mtest_test

import (
	"fmt"
	"testing"

	"github.com/creachadair/mds/mtest"
)

// testStub implements the mtest.TB interface as a capturing shim to verify
// that test failures are reported properly.
type testStub struct {
	failed bool
	text   string
}

func (t *testStub) Fatal(args ...any) {
	t.failed = true
	t.text = fmt.Sprint(args...)
}

func (t *testStub) Fatalf(msg string, args ...any) {
	t.failed = true
	t.text = fmt.Sprintf(msg, args...)
}

func (*testStub) Helper() {}

func TestMustPanic(t *testing.T) {
	t.Run("OK", func(t *testing.T) {
		v := mtest.MustPanic(t, func() { panic("pass") })
		t.Logf("Panic reported: %v", v)
	})

	t.Run("Fail", func(t *testing.T) {
		var s testStub
		v := mtest.MustPanic(&s, func() {})
		if !s.failed {
			t.Error("Test did not fail as expected")
		}
		if s.text == "" {
			t.Error("Failure did not log a message")
		}
		if v != nil {
			t.Errorf("Unexpected panic value: %v", v)
		}
	})
}

func TestMustPanicf(t *testing.T) {
	t.Run("OK", func(t *testing.T) {
		v := mtest.MustPanicf(t, func() { panic("pass") }, "bad things")
		t.Logf("Panic reported: %v", v)
	})

	t.Run("Fail", func(t *testing.T) {
		var s testStub
		v := mtest.MustPanicf(&s, func() {}, "bad: %d", 11)
		if !s.failed {
			t.Error("Test did not fail as expected")
		}
		if s.text != "bad: 11" {
			t.Errorf("Wrong message: got %q, want bad: 11", s.text)
		}
		if v != nil {
			t.Errorf("Unexpected panic value: %v", v)
		}
	})
}
