package slice_test

import (
	"testing"

	"github.com/creachadair/mds/slice"
	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
)

type testCase[T comparable] struct {
	desc  string
	input []T
	want  []T // desired output, including order
	keep  func(T) bool
}

func TestPartition(t *testing.T) {
	for _, test := range []testCase[string]{
		{"Nil keep all", nil, nil,
			func(string) bool { return true }},
		{"Nil drop all", nil, nil,
			func(string) bool { return false }},
		{"One keep", []string{"x"}, []string{"x"},
			func(string) bool { return true }},
		{"One drop", []string{"x"}, nil,
			func(string) bool { return false }},
		{"Keep all",
			[]string{"a", "b", "c"}, []string{"a", "b", "c"},
			func(string) bool { return true }},
		{"Drop all",
			[]string{"a", "b", "c"}, nil,
			func(string) bool { return false }},
		{"Keep vowels",
			[]string{"a", "b", "c", "d", "e", "f", "g"}, []string{"a", "e"},
			func(s string) bool { return s == "a" || s == "e" }},
		{"Drop vowels",
			[]string{"a", "b", "c", "d", "e", "f", "g"}, []string{"b", "c", "d", "f", "g"},
			func(s string) bool { return s != "a" && s != "e" }},
		{"Even-length strings",
			[]string{"laugh", "while", "you", "can", "monkey", "boy"}, []string{"monkey"},
			func(s string) bool { return len(s)%2 == 0 }},
		{"Longer strings",
			[]string{"join", "us", "now", "and", "share", "the", "software"},
			[]string{"share", "software"},
			func(s string) bool { return len(s) > 4 }},
	} {
		t.Run(test.desc, test.partition)
	}
	for _, test := range []testCase[int]{
		{"Less than 5",
			[]int{8, 0, 2, 7, 5, 3, 4}, []int{0, 2, 3, 4},
			func(z int) bool { return z < 5 }},
		{"Keep runs",
			[]int{2, 2, 4, 1, 1, 3, 6, 6, 6, 5, 8}, []int{2, 2, 4, 6, 6, 6, 8},
			func(z int) bool { return z%2 == 0 }},
	} {
		t.Run(test.desc, test.partition)
	}
}

func TestDedup(t *testing.T) {
	for _, test := range []testCase[int]{
		{"Nil", nil, nil, nil},
		{"Empty", []int{}, nil, nil},
		{"One", []int{100}, []int{100}, nil},
		{"NoRuns", []int{1, 3, 2, 4}, []int{1, 3, 2, 4}, nil},
		{"Single", []int{5, 5, 5, 5, 5}, []int{5}, nil},
		{"Two", []int{2, 2, 2, 3, 3, 3}, []int{2, 3}, nil},
		{"Repeat", []int{1, 3, 3, 1}, []int{1, 3, 1}, nil},
		{"NoRunsAsc", []int{0, 1, 2, 3, 4}, []int{0, 1, 2, 3, 4}, nil},
		{"NoRunsDesc", []int{10, 9, 8, 7}, []int{10, 9, 8, 7}, nil},
		{"RunsAsc", []int{0, 1, 1, 1, 2, 2, 3}, []int{0, 1, 2, 3}, nil},
		{"RunsDesc", []int{5, 5, 5, 3, 3, 1, 1, 0}, []int{5, 3, 1, 0}, nil},

		// Runs:           a  b---  c  b---  d---  e---------
		{"Unsorted", []int{1, 0, 0, 9, 0, 0, 3, 3, 2, 2, 2, 2}, []int{1, 0, 9, 0, 3, 2}, nil},
	} {
		t.Run(test.desc, test.dedup)
	}
}

func TestRemove(t *testing.T) {
	tests := []struct {
		name  string
		input []int
		pos   int
		want  []int
	}{
		{"Nil", nil, 0, nil},
		{"Empty", []int{}, 0, []int{}},
		{"End", []int{1, 2, 3}, 3, []int{1, 2, 3}},
		{"Last", []int{1, 2, 3}, 2, []int{1, 2}},
		{"Middle", []int{1, 2, 3}, 1, []int{1, 3}},
		{"First", []int{1, 2, 3}, 0, []int{2, 3}},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			got := slice.Remove(test.input, test.pos)
			if diff := cmp.Diff(test.want, got); diff != "" {
				t.Errorf("Remove(%v, %d) result (-want, +got)\n%s", test.input, test.pos, diff)
			}
		})
	}
}

func TestInsert(t *testing.T) {
	tests := []struct {
		name  string
		input []int
		value int
		pos   int
		want  []int
	}{
		{"Nil", nil, 10, 0, []int{10}},
		{"Empty", []int{}, 20, 0, []int{20}},
		{"End", []int{1, 2}, 3, 2, []int{1, 2, 3}},
		{"Middle", []int{1, 3}, 2, 1, []int{1, 2, 3}},
		{"First", []int{2, 3}, 1, 0, []int{1, 2, 3}},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			got := slice.Insert(test.input, test.pos, test.value)
			if diff := cmp.Diff(test.want, got); diff != "" {
				t.Errorf("Insert(%v, %v, %d) result (-want, +got)\n%s",
					test.input, test.value, test.pos, diff)
			}
		})
	}
}

func TestReverse(t *testing.T) {
	tests := []struct {
		name  string
		input []int
		want  []int
	}{
		{"Nil", nil, nil},
		{"Empty", []int{}, nil},
		{"Single", []int{11}, []int{11}},
		{"Multiple", []int{1, 2, 3, 4, 5}, []int{5, 4, 3, 2, 1}},
		{"Palindrome", []int{1, 2, 3, 2, 1}, []int{1, 2, 3, 2, 1}},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			cp := append([]int(nil), tc.input...)
			slice.Reverse(cp)
			if diff := cmp.Diff(tc.want, cp); diff != "" {
				t.Errorf("Reverse(%v) result (-want, +got)\n%s", tc.input, diff)
			}
		})
	}
}

func TestMapKeys(t *testing.T) {
	cmpStrings := func(a, b string) bool { return a < b }

	tests := []struct {
		input map[string]int
		want  []string
	}{
		{nil, nil},
		{map[string]int{}, nil},
		{map[string]int{"a": 1, "b": 2, "c": 3}, []string{"a", "b", "c"}},
	}
	for _, tc := range tests {
		got := slice.MapKeys(tc.input)
		diff := cmp.Diff(tc.want, got, cmpopts.SortSlices(cmpStrings))
		if diff != "" {
			t.Logf("Input: %v", tc.input)
			t.Errorf("MapKeys (-want, +got):\n%s", diff)
		}
	}
}

func (tc *testCase[T]) partition(t *testing.T) {
	t.Helper()

	cp := copyOf(tc.input)
	t.Logf("Input: %+v", cp)

	got := slice.Partition(cp, tc.keep)
	t.Logf("After partition: %+v ~ %+v", got, cp[len(got):])
	diff := cmp.Diff(tc.want, got, cmpopts.EquateEmpty())
	if diff != "" {
		t.Errorf("Partition result (-want, +got)\n%s", diff)
	}
}

func (tc *testCase[T]) dedup(t *testing.T) {
	t.Helper()

	cp := copyOf(tc.input)
	t.Logf("Input: %+v", cp)

	got := slice.Dedup(cp)
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
