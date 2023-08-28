package slice_test

import (
	"sort"
	"strings"
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

	type testCase[K comparable, V any, T ~map[K]V] struct {
		input T
		want  []K
	}
	type TM map[string]int

	tests := []testCase[string, int, map[string]int]{
		{map[string]int(nil), nil},
		{map[string]int{}, nil},
		{map[string]int{"a": 1, "b": 2, "c": 3}, []string{"a", "b", "c"}},
		{TM(nil), nil},
		{TM{}, nil},
		{TM{"a": 1, "b": 2, "c": 3}, []string{"a", "b", "c"}},
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

func TestSplit(t *testing.T) {
	tests := []struct {
		input    string
		i        int
		lhs, rhs []string
	}{
		{"", 0, nil, nil},
		{"a", 0, nil, []string{"a"}},
		{"a", 1, []string{"a"}, nil},
		{"a b c", 0, nil, []string{"a", "b", "c"}},
		{"a b c", 1, []string{"a"}, []string{"b", "c"}},
		{"a b c", 2, []string{"a", "b"}, []string{"c"}},
		{"a b c", 3, []string{"a", "b", "c"}, nil},
		{"a b c d", -1, []string{"a", "b", "c"}, []string{"d"}},
		{"a b c d", -2, []string{"a", "b"}, []string{"c", "d"}},
		{"a b c d", -3, []string{"a"}, []string{"b", "c", "d"}},
		{"a b c d", -4, nil, []string{"a", "b", "c", "d"}},
	}
	for _, tc := range tests {
		input := strings.Fields(tc.input)
		lhs, rhs := slice.Split(input, tc.i)
		if diff := cmp.Diff(tc.lhs, lhs, cmpopts.EquateEmpty()); diff != "" {
			t.Errorf("Split %q %d lhs (-want, +got):\n%s", input, tc.i, diff)
		}
		if diff := cmp.Diff(tc.rhs, rhs, cmpopts.EquateEmpty()); diff != "" {
			t.Errorf("Split %q %d rhs (-want, +got):\n%s", input, tc.i, diff)
		}
	}
	t.Run("Range", func(t *testing.T) {
		mustPanic(t, func() { slice.Split([]string(nil), 1) })
		mustPanic(t, func() { slice.Split([]string(nil), -1) })
		mustPanic(t, func() { slice.Split([]int{1, 2, 3}, 4) })
		mustPanic(t, func() { slice.Split([]int{1, 2, 3}, -4) })
	})
}

func TestSplitLast(t *testing.T) {
	tests := []struct {
		input string
		rest  []string
		want  string
	}{
		{"", nil, ""},
		{"foo", nil, "foo"},
		{"foo bar", []string{"foo"}, "bar"},
		{"foo bar baz", []string{"foo", "bar"}, "baz"},
	}
	for _, tc := range tests {
		input := strings.Fields(tc.input)
		rest, got := slice.SplitLast(input)
		if diff := cmp.Diff(tc.rest, rest, cmpopts.EquateEmpty()); diff != "" {
			t.Errorf("SplitLast %q rest (-want, +got):\n%s", input, diff)
		}
		if got != tc.want {
			t.Errorf("SplitLast %q last: got %q, want %q", input, got, tc.want)
		}
	}
}

func TestMatchingKeys(t *testing.T) {
	even := func(z int) bool { return z%2 == 0 }
	big := func(z int) bool { return z > 10 }
	type M = map[string]int
	tests := []struct {
		m    M
		f    func(int) bool
		want []string
	}{
		{nil, even, nil},
		{M{}, even, nil},
		{M{"x": 1, "y": 2, "z": 3}, even, []string{"y"}},
		{M{"x": 1, "y": 3}, even, nil},
		{M{"x": 2, "y": 4}, even, []string{"x", "y"}},
		{M{"x": 2, "y": 3, "z": 5, "a": 11, "b": 13, "c": 19, "d": 0}, big,
			[]string{"a", "b", "c"}},
	}
	for _, tc := range tests {
		got := slice.MatchingKeys(tc.m, tc.f)
		sort.Strings(got)
		sort.Strings(tc.want)
		if diff := cmp.Diff(tc.want, got); diff != "" {
			t.Errorf("MatchingKeys (-want, +got):\n%s", diff)
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

func mustPanic(t *testing.T, f func()) {
	defer func() {
		if x := recover(); x != nil {
			t.Logf("Panic recovered (OK): %v", x)
		}
	}()
	f()
	t.Fatal("Expected panic did not occur")
}
