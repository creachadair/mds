package value_test

import (
	"testing"

	"github.com/creachadair/mds/value"
)

func TestPtr(t *testing.T) {
	p1 := value.Ptr("foo")
	p2 := value.Ptr("foo")
	if p1 == p2 {
		t.Errorf("Values should have distinct pointers (%p == %p)", p1, p1)
	}
	if *p1 != "foo" || *p2 != "foo" {
		t.Errorf("Got p1=%q, p2=%q; wanted both foo", *p1, *p2)
	}
}
