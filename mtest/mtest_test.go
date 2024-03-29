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

func (*testStub) Helper()        {}
func (*testStub) Cleanup(func()) {}

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

func TestSwap(t *testing.T) {
	testValue := "original"

	t.Run("Swapped", func(t *testing.T) {
		old := mtest.Swap(t, &testValue, "replacement")

		if old != "original" {
			t.Errorf("Old value is %q, want original", old)
		}
		if testValue != "replacement" {
			t.Errorf("Test value is %q, want replacement", testValue)
		}
	})

	t.Run("NoSwap", func(t *testing.T) {
		if testValue != "original" {
			t.Errorf("Test value is %q, want original", testValue)
		}
	})

	if testValue != "original" {
		t.Errorf("Test value after is %q, want original", testValue)
	}
}
