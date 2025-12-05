package mtest_test

import (
	"fmt"
	"io"
	"net/http"
	"testing"

	"github.com/creachadair/mds/mtest"
)

// testStub implements the mtest.TB interface as a capturing shim to verify
// that test failures are reported properly.
type testStub struct {
	failed bool
	text   string
}

func (testStub) Name() string { return "Test" }

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

func TestDiffLines(t *testing.T) {
	const s1 = "a\nbc\ndef\ng"
	const s2 = "a\nbc\nqq\ndef\n"
	if s1 == s2 {
		t.Fatalf("Probe strings are equal: %q", s1)
	}

	t.Run("Equal", func(t *testing.T) {
		if diff := mtest.DiffLines(s1, s1); diff != "" {
			t.Errorf("Diff %q, %q: got %q, want empty", s1, s1, diff)
		}
		if diff := mtest.DiffLines(s2, s2); diff != "" {
			t.Errorf("Diff %q, %q: got %q, want empty", s2, s2, diff)
		}
	})

	t.Run("Unequal", func(t *testing.T) {
		const want = `@@ -1,4 +1,5 @@
 a
 bc
+qq
 def
-g
+
`
		diff := mtest.DiffLines(s1, s2)
		t.Logf("Diff from %q to %q is\n%s", s1, s2, diff)
		if diff != want {
			t.Errorf("Wrong diff:\ngot  %q\nwant %q", diff, want)
		}
	})
}

func TestNewHTTPServer(t *testing.T) {
	m := http.NewServeMux()
	m.HandleFunc("GET /test", func(w http.ResponseWriter, r *http.Request) {
		t.Logf("Reqquest handler: caller is %q", r.RemoteAddr)
		http.Error(w, "ok", http.StatusOK)
	})
	srv, cli := mtest.NewHTTPServer(t, m)
	if got, want := srv.URL, "http://server:12345"; got != want {
		t.Errorf("Server URL: got %q, want %q", got, want)
	}

	rsp, err := cli.Get(srv.URL + "/test")
	if err != nil {
		t.Fatalf("Get failed: %v", err)
	}
	body, err := io.ReadAll(rsp.Body)
	rsp.Body.Close()
	if err != nil {
		t.Errorf("Read body: %v", err)
	}
	if rsp.StatusCode != http.StatusOK {
		t.Errorf("Status code: got %d, want %d", rsp.StatusCode, http.StatusOK)
	}
	if got, want := string(body), "ok\n"; got != want {
		t.Errorf("Response body: got %q, want %q", got, want)
	}
}
