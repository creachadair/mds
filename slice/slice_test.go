package slice_test

import (
	"sort"
	"strings"
	"testing"

	"github.com/creachadair/mds/mtest"
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
		mtest.MustPanic(t, func() { slice.Split([]string(nil), 1) })
		mtest.MustPanic(t, func() { slice.Split([]string(nil), -1) })
		mtest.MustPanic(t, func() { slice.Split([]int{1, 2, 3}, 4) })
		mtest.MustPanic(t, func() { slice.Split([]int{1, 2, 3}, -4) })
	})
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

func TestRotate(t *testing.T) {
	tests := []struct {
		input string
		k     int
		want  string
	}{
		{"x", 0, "x"},
		{"x", -1, "x"},

		{"a b c", 0, "a b c"},
		{"a b c", 1, "c a b"},
		{"a b c", 2, "b c a"},

		{"a b c", -1, "b c a"},
		{"a b c", -2, "c a b"},
		{"a b c", -3, "a b c"},

		{"30 987 405 309", 0, "30 987 405 309"},
		{"30 987 405 309", 1, "309 30 987 405"},
		{"30 987 405 309", 2, "405 309 30 987"},
		{"30 987 405 309", 3, "987 405 309 30"},
		{"30 987 405 309", 4, "30 987 405 309"},

		{"30 987 405 309", -3, "309 30 987 405"},
		{"30 987 405 309", -2, "405 309 30 987"},
		{"30 987 405 309", -1, "987 405 309 30"},

		{"a b c d e f g h", 0, "a b c d e f g h"},
		{"a b c d e f g h", 1, "h a b c d e f g"},
		{"a b c d e f g h", -1, "b c d e f g h a"},
		{"a b c d e f g h", 2, "g h a b c d e f"},
		{"a b c d e f g h", -2, "c d e f g h a b"},
		{"a b c d e f g h", 3, "f g h a b c d e"},
		{"a b c d e f g h", -3, "d e f g h a b c"},
		{"a b c d e f g h", 4, "e f g h a b c d"},
		{"a b c d e f g h", -4, "e f g h a b c d"},
		{"a b c d e f g h", 5, "d e f g h a b c"},
		{"a b c d e f g h", -5, "f g h a b c d e"},
		{"a b c d e f g h", 6, "c d e f g h a b"},
		{"a b c d e f g h", -6, "g h a b c d e f"},
	}
	for _, tc := range tests {
		got := strings.Fields(tc.input)
		slice.Rotate(got, tc.k)
		if diff := cmp.Diff(got, strings.Fields(tc.want)); diff != "" {
			t.Errorf("Rotate %s %d (-got, +want):\n%s", tc.input, tc.k, diff)
		}
	}

	t.Run("Bounds", func(t *testing.T) {
		mtest.MustPanic(t, func() { slice.At([]string(nil), 0) })
		mtest.MustPanic(t, func() { slice.At([]string(nil), -1) })
		mtest.MustPanic(t, func() { slice.At([]string{}, 0) })
		mtest.MustPanic(t, func() { slice.At([]string{}, -1) })
		mtest.MustPanic(t, func() { slice.At([]string{"a", "b", "c"}, 10) })
		mtest.MustPanic(t, func() { slice.At([]string{"a", "b", "c"}, -5) })
	})
}

func TestAt(t *testing.T) {
	tests := []struct {
		input string
		k     int
		want  string
	}{
		{"", 0, ""}, // empty input
		{"X", 0, "X"},
		{"X", 1, "X"},
		{"X", -1, "X"},
		{"A B", 0, "A B"},
		{"A B", 1, "B A"},
		{"A B", -1, "B A"},
		{"A B C D E", 0, "A B C D E"},
		{"A B C D E", 1, "E A B C D"},
		{"A B C D E", 2, "D E A B C"},
		{"A B C D E", 3, "C D E A B"},
		{"A B C D E", 4, "B C D E A"},
		{"A B C D E", 5, "A B C D E"},
		{"A B C D E", -1, "B C D E A"},
		{"A B C D E", -2, "C D E A B"},
		{"A B C D E", -3, "D E A B C"},
		{"A B C D E", -4, "E A B C D"},
		{"A B C D E", -5, "A B C D E"},
	}
	for _, tc := range tests {
		got := strings.Fields(tc.input)
		slice.Rotate(got, tc.k)
		if diff := cmp.Diff(got, strings.Fields(tc.want)); diff != "" {
			t.Errorf("Rotate %q %d (-got, +want):\n%s", tc.input, tc.k, diff)
		}
	}
}

func TestCoalesce(t *testing.T) {
	tests := []struct {
		input []string
		want  string
	}{
		{nil, ""},
		{[]string{}, ""},
		{[]string{""}, ""},
		{[]string{"", ""}, ""},

		{[]string{"a"}, "a"},
		{[]string{"a", "b"}, "a"},
		{[]string{"a", "", "b"}, "a"},
		{[]string{"", "a", "b"}, "a"},
		{[]string{"", "", "b"}, "b"},
		{[]string{"", "", "", "q", ""}, "q"},
	}
	for _, tc := range tests {
		got := slice.Coalesce(tc.input...)
		if got != tc.want {
			t.Errorf("Coalesce %q: got %q, want %q", tc.input, got, tc.want)
		}
	}
}

func TestChunks(t *testing.T) {
	tests := []struct {
		input string
		n     int
		want  [][]string
	}{
		// An empty slice has only one covering.
		{"", 1, [][]string{{}}},
		{"", 5, [][]string{{}}},

		{"x", 1, [][]string{{"x"}}},
		{"x", 2, [][]string{{"x"}}},

		{"a b c d e", 1, [][]string{{"a"}, {"b"}, {"c"}, {"d"}, {"e"}}},
		{"a b c d e", 2, [][]string{{"a", "b"}, {"c", "d"}, {"e"}}},
		{"a b c d e", 3, [][]string{{"a", "b", "c"}, {"d", "e"}}},
		{"a b c d e", 4, [][]string{{"a", "b", "c", "d"}, {"e"}}},
		{"a b c d e", 5, [][]string{{"a", "b", "c", "d", "e"}}},
		{"a b c d e", 6, [][]string{{"a", "b", "c", "d", "e"}}}, // n > len(input)
	}
	for _, tc := range tests {
		got := slice.Chunks(strings.Fields(tc.input), tc.n)
		for i := 0; i+1 < len(got); i++ {
			if len(got[i]) != tc.n {
				t.Errorf("Chunk %d has length %d, want %d", i+1, len(got[i]), tc.n)
			}
		}
		if diff := cmp.Diff(got, tc.want); diff != "" {
			t.Errorf("Chunks(%q, %d): (-got, +want)\n%s", tc.input, tc.n, diff)
		}
	}

	t.Logf("OK n=0: %v", mtest.MustPanic(t, func() { slice.Chunks([]string{"a"}, 0) }))
	t.Logf("OK n<0: %v", mtest.MustPanic(t, func() { slice.Chunks([]string{"a"}, -1) }))
}

func TestBatches(t *testing.T) {
	input := []int{1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13}
	tests := []struct {
		n    int
		want [][]int
	}{
		{1, [][]int{input}},
		{2, [][]int{{1, 2, 3, 4, 5, 6, 7}, {8, 9, 10, 11, 12, 13}}},
		{3, [][]int{{1, 2, 3, 4, 5}, {6, 7, 8, 9}, {10, 11, 12, 13}}},
		{4, [][]int{{1, 2, 3, 4}, {5, 6, 7}, {8, 9, 10}, {11, 12, 13}}},
		{5, [][]int{{1, 2, 3}, {4, 5, 6}, {7, 8, 9}, {10, 11}, {12, 13}}},
		{6, [][]int{{1, 2, 3}, {4, 5}, {6, 7}, {8, 9}, {10, 11}, {12, 13}}},
		{7, [][]int{{1, 2}, {3, 4}, {5, 6}, {7, 8}, {9, 10}, {11, 12}, {13}}},
		{8, [][]int{{1, 2}, {3, 4}, {5, 6}, {7, 8}, {9, 10}, {11}, {12}, {13}}},
		{9, [][]int{{1, 2}, {3, 4}, {5, 6}, {7, 8}, {9}, {10}, {11}, {12}, {13}}},
		{10, [][]int{{1, 2}, {3, 4}, {5, 6}, {7}, {8}, {9}, {10}, {11}, {12}, {13}}},
		{11, [][]int{{1, 2}, {3, 4}, {5}, {6}, {7}, {8}, {9}, {10}, {11}, {12}, {13}}},
		{12, [][]int{{1, 2}, {3}, {4}, {5}, {6}, {7}, {8}, {9}, {10}, {11}, {12}, {13}}},
		{13, [][]int{{1}, {2}, {3}, {4}, {5}, {6}, {7}, {8}, {9}, {10}, {11}, {12}, {13}}},
	}
	for _, tc := range tests {
		got := slice.Batches(input, tc.n)
		if diff := cmp.Diff(got, tc.want); diff != "" {
			t.Errorf("Batches(%v, %d): (-got, +want)\n%s", input, tc.n, diff)
		}
	}

	t.Logf("OK n=0: %v", mtest.MustPanic(t, func() { slice.Batches(input, 0) }))
	t.Logf("OK n<0: %v", mtest.MustPanic(t, func() { slice.Batches(input, -1) }))
	t.Logf("OK n>len: %v", mtest.MustPanic(t, func() { slice.Batches(input, len(input)+1) }))
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
