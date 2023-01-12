package slice_test

import (
	"testing"

	"github.com/creachadair/mds/slice"
	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
)

type testCase[T comparable] struct {
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
		t.Run(test.desc, test.partition)
	}
	for _, test := range []testCase[int]{
		{"Less than 5",
			[]int{8, 0, 2, 7, 5, 3, 4}, []int{0, 2, 3, 4}, 4,
			func(z int) bool { return z < 5 }},
		{"Keep runs",
			[]int{2, 2, 4, 1, 1, 3, 6, 6, 6, 5, 8}, []int{2, 2, 4, 6, 6, 6, 8}, 7,
			func(z int) bool { return z%2 == 0 }},
	} {
		t.Run(test.desc, test.partition)
	}
}

func TestDedup(t *testing.T) {
	for _, test := range []testCase[int]{
		{"Nil", nil, nil, 0, nil},
		{"Empty", []int{}, nil, 0, nil},
		{"One", []int{100}, []int{100}, 1, nil},
		{"NoRuns", []int{1, 3, 2, 4}, []int{1, 3, 2, 4}, 4, nil},
		{"Single", []int{5, 5, 5, 5, 5}, []int{5}, 1, nil},
		{"Two", []int{2, 2, 2, 3, 3, 3}, []int{2, 3}, 2, nil},
		{"Repeat", []int{1, 3, 3, 1}, []int{1, 3, 1}, 3, nil},
		{"NoRunsAsc", []int{0, 1, 2, 3, 4}, []int{0, 1, 2, 3, 4}, 5, nil},
		{"NoRunsDesc", []int{10, 9, 8, 7}, []int{10, 9, 8, 7}, 4, nil},
		{"RunsAsc", []int{0, 1, 1, 1, 2, 2, 3}, []int{0, 1, 2, 3}, 4, nil},
		{"RunsDesc", []int{5, 5, 5, 3, 3, 1, 1, 0}, []int{5, 3, 1, 0}, 4, nil},

		// Runs:           a  b---  c  b---  d---  e---------
		{"Unsorted", []int{1, 0, 0, 9, 0, 0, 3, 3, 2, 2, 2, 2}, []int{1, 0, 9, 0, 3, 2}, 6, nil},
	} {
		t.Run(test.desc, test.dedup)
	}
}

func (tc *testCase[T]) partition(t *testing.T) {
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
		t.Errorf("Partition result (-want, +got)\n%s", diff)
	}
}

func (tc *testCase[T]) dedup(t *testing.T) {
	t.Helper()

	cp := copyOf(tc.input)
	t.Logf("Input: %+v", cp)

	got := slice.Dedup(cp)
	if len(got) != tc.wantPos {
		t.Errorf("Dedup index: got %d, want %d", len(got), tc.wantPos)
	}

	t.Logf("After dedup: %+v ~ %+v", got, cp[len(got):])
	diff := cmp.Diff(tc.want, got, cmpopts.EquateEmpty())
	if diff != "" {
		t.Errorf("Dedup result (-want, +got)\n%s", diff)
	}
}

func copyOf[T any](vs []T) []T {
	out := make([]T, len(vs))
	copy(out, vs)
	return out
}
