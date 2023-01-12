package slice_test

import (
	"testing"

	"github.com/creachadair/mds/slice"
	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
)

type testCase[T any] struct {
	desc    string
	input   []T
	want    []T // desired output, including order
	wantPos int // desired index
	keep    func(T) bool
}

func TestPartition(t *testing.T) {
	for _, test := range []testCase[string]{
		{"Empty input keep everything", nil, nil, 0,
			func(string) bool { return true }},
		{"Empty input keep nothing", nil, nil, 0,
			func(string) bool { return false }},
		{"Keep everything",
			[]string{"a", "b", "c"}, []string{"a", "b", "c"}, 3,
			func(string) bool { return true }},
		{"Drop everything",
			[]string{"a", "b", "c"}, nil, 0,
			func(string) bool { return false }},
		{"Keep vowels",
			[]string{"a", "b", "c", "d", "e", "f", "g"}, []string{"a", "e"}, 2,
			func(s string) bool { return s == "a" || s == "e" }},
		{"Drop vowels",
			[]string{"a", "b", "c", "d", "e", "f", "g"}, []string{"b", "c", "d", "f", "g"}, 5,
			func(s string) bool { return s != "a" && s != "e" }},
		{"Even-length strings",
			[]string{"laugh", "while", "you", "can", "monkey", "boy"}, []string{"monkey"}, 1,
			func(s string) bool { return len(s)%2 == 0 }},
		{"Longer strings",
			[]string{"join", "us", "now", "and", "share", "the", "software"},
			[]string{"share", "software"}, 2,
			func(s string) bool { return len(s) > 4 }},
	} {
		t.Run(test.desc, test.run)
	}
	for _, test := range []testCase[int]{
		{"Less than 5",
			[]int{8, 0, 2, 7, 5, 3, 4}, []int{0, 2, 3, 4}, 4,
			func(z int) bool { return z < 5 }},
		{"Keep runs",
			[]int{2, 2, 4, 1, 1, 3, 6, 6, 6, 5, 8}, []int{2, 2, 4, 6, 6, 6, 8}, 7,
			func(z int) bool { return z%2 == 0 }},
	} {
		t.Run(test.desc, test.run)
	}
}

func (tc *testCase[T]) run(t *testing.T) {
	t.Helper()

	cp := copyOf(tc.input)
	t.Logf("Input: %+v", cp)

	gotPos := slice.Partition(cp, tc.keep)
	if gotPos != tc.wantPos {
		t.Errorf("Partition index: got %d, want %d", gotPos, tc.wantPos)
	}

	t.Logf("After partition: %+v ~ %+v", cp[:gotPos], cp[gotPos:])
	diff := cmp.Diff(tc.want, cp[:gotPos], cmpopts.EquateEmpty())
	if diff != "" {
		if diff != "" {
			t.Errorf("Partition result (-want, +got)\n%s", diff)
		}
	}
}

func copyOf[T any](vs []T) []T {
	out := make([]T, len(vs))
	copy(out, vs)
	return out
}
