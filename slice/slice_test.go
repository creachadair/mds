package slice_test

import (
	"slices"
	"sort"
	"strconv"
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

func TestZero(t *testing.T) {
	zs := []int{1, 2, 3, 4, 5}
	slice.Zero(zs[3:])
	if diff := cmp.Diff(zs, []int{1, 2, 3, 0, 0}); diff != "" {
		t.Errorf("Zero (-got, +want):\n%s", diff)
	}
	zs = []int{1, 2, 3, 4, 5}
	slice.Zero(zs[:3])
	if diff := cmp.Diff(zs, []int{0, 0, 0, 4, 5}); diff != "" {
		t.Errorf("Zero (-got, +want):\n%s", diff)
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
		var got []string
		for key := range slice.MatchingKeys(tc.m, tc.f) {
			got = append(got, key)
		}
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
		mtest.MustPanic(t, func() { slice.At([]string{"a", "b", "c"}, 3) })
		mtest.MustPanic(t, func() { slice.At([]string{"a", "b", "c"}, 10) })
		mtest.MustPanic(t, func() { slice.At([]string{"a", "b", "c"}, -5) })
	})
}

func TestAt(t *testing.T) {
	mtest.MustPanic(t, func() { slice.At([]int(nil), 0) })
	mtest.MustPanic(t, func() { slice.At([]int{1, 2}, 5) })
	mtest.MustPanic(t, func() { slice.At([]int{1, 2}, -3) })

	tests := []struct {
		input string
		k     int
		want  string
	}{
		{"X", 0, "X"},
		{"X", -1, "X"},
		{"A B", 0, "A"},
		{"A B", 1, "B"},
		{"A B", -1, "B"},
		{"A B", -2, "A"},
		{"A B C D E", 0, "A"},
		{"A B C D E", 1, "B"},
		{"A B C D E", 2, "C"},
		{"A B C D E", 3, "D"},
		{"A B C D E", 4, "E"},
		{"A B C D E", -1, "E"},
		{"A B C D E", -2, "D"},
		{"A B C D E", -3, "C"},
		{"A B C D E", -4, "B"},
		{"A B C D E", -5, "A"},
	}
	for _, tc := range tests {
		input := strings.Fields(tc.input)
		if got := slice.At(input, tc.k); got != tc.want {
			t.Errorf("At %q %d: got %q, want %q", input, tc.k, got, tc.want)
		}
	}
}

func TestPtrAt(t *testing.T) {
	tests := []struct {
		input string
		k     int
		want  string
	}{
		{"X", 0, "X"},
		{"X", -1, "X"},
		{"A B", 0, "A"},
		{"A B", 1, "B"},
		{"A B", -1, "B"},
		{"A B", -2, "A"},
		{"A B C D E", 0, "A"},
		{"A B C D E", 1, "B"},
		{"A B C D E", 2, "C"},
		{"A B C D E", 3, "D"},
		{"A B C D E", 4, "E"},
		{"A B C D E", -1, "E"},
		{"A B C D E", -2, "D"},
		{"A B C D E", -3, "C"},
		{"A B C D E", -4, "B"},
		{"A B C D E", -5, "A"},
	}
	for _, tc := range tests {
		input := strings.Fields(tc.input)

		got := slice.PtrAt(input, tc.k)
		idx := tc.k
		if idx < 0 {
			idx += len(input)
		}
		if want := &input[idx]; got != want {
			t.Errorf("PtrAt %q %d: got ptr %p, want %p", input, tc.k, got, want)
		}
		if *got != tc.want {
			t.Errorf("PtrAt %q %d: got value %q, want %q", input, tc.k, *got, tc.want)
		}
	}
	if got := slice.PtrAt([]string{"a"}, 10); got != nil {
		t.Errorf("PtrAt(*, 10): got %v, want nil", got)
	}
}

func TestChunks(t *testing.T) {
	tests := []struct {
		input string
		n     int
		want  [][]string
	}{
		// An empty slice has only one covering.
		{"", 0, nil},
		{"", 1, nil},
		{"", 5, nil},

		{"x", 0, nil},
		{"x", 1, [][]string{{"x"}}},
		{"x", 2, [][]string{{"x"}}},

		{"a b c d e", 0, nil},
		{"a b c d e", 1, [][]string{{"a"}, {"b"}, {"c"}, {"d"}, {"e"}}},
		{"a b c d e", 2, [][]string{{"a", "b"}, {"c", "d"}, {"e"}}},
		{"a b c d e", 3, [][]string{{"a", "b", "c"}, {"d", "e"}}},
		{"a b c d e", 4, [][]string{{"a", "b", "c", "d"}, {"e"}}},
		{"a b c d e", 5, [][]string{{"a", "b", "c", "d", "e"}}},
		{"a b c d e", 6, [][]string{{"a", "b", "c", "d", "e"}}}, // n > len(input)
	}
	for _, tc := range tests {
		got := slices.Collect(slice.Chunks(strings.Fields(tc.input), tc.n))
		for i := range len(got) - 1 {
			if len(got[i]) != tc.n {
				t.Errorf("Chunk %d has length %d, want %d", i+1, len(got[i]), tc.n)
			}
		}
		if diff := cmp.Diff(got, tc.want); diff != "" {
			t.Errorf("Chunks(%q, %d): (-got, +want)\n%s", tc.input, tc.n, diff)
		}
	}

	t.Logf("OK n<0: %v", mtest.MustPanic(t, func() { slice.Chunks([]string{"a"}, -1) }))
}

func TestBatches(t *testing.T) {
	input := []int{1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13}
	tests := []struct {
		n    int
		want [][]int
	}{
		{0, nil},
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
		{100, [][]int{{1}, {2}, {3}, {4}, {5}, {6}, {7}, {8}, {9}, {10}, {11}, {12}, {13}}},
	}
	for _, tc := range tests {
		got := slices.Collect(slice.Batches(input, tc.n))
		if diff := cmp.Diff(got, tc.want); diff != "" {
			t.Errorf("Batches(%v, %d): (-got, +want)\n%s", input, tc.n, diff)
		}
	}

	t.Logf("OK n<0: %v", mtest.MustPanic(t, func() { slice.Batches(input, -1) }))
}

func TestStripe(t *testing.T) {
	split := func(s string) []string {
		out := strings.Fields(s)
		for i, w := range out {
			if w == "@" {
				out[i] = ""
			}
		}
		return out
	}
	makeInput := func(s string) [][]string {
		var out [][]string
		for _, line := range strings.Split(s, "|") {
			out = append(out, split(line))
		}
		return out
	}
	tests := []struct {
		input string
		i     int
		want  string
	}{
		{"", 0, ""},

		{"a", 0, "a"},
		{"a", 1, ""},

		{"a b c", 0, "a"},
		{"a b c", 1, "b"},
		{"a b c", 2, "c"},
		{"a b c", 3, ""},

		{"a b c|d e f", 0, "a d"},
		{"a b c|d e f", 1, "b e"},
		{"a b c|d e f", 2, "c f"},

		{"a b c|d e|f g h i", 0, "a d f"},
		{"a b c|d e|f g h i", 1, "b e g"},
		{"a b c|d e|f g h i", 2, "c h"},
		{"a b c|d e|f g h i", 3, "i"},
		{"a b c|d e|f g h i", 4, ""},

		{"a @ c|@ e f|g h i", 0, "a @ g"},
		{"a @ c|@ e f|g h i", 1, "@ e h"},
		{"a @ c|@ e f|g h i", 2, "c f i"},
		{"a @ c|@ e f|g h i", 3, ""},

		{"a b c d|e f @ g h|i j @", 0, "a e i"},
		{"a b c d|e f @ g h|i j @", 1, "b f j"},
		{"a b c d|e f @ g h|i j @", 2, "c @ @"},
		{"a b c d|e f @ g h|i j @", 3, "d g"},
		{"a b c d|e f @ g h|i j @", 4, "h"},
		{"a b c d|e f @ g h|i j @", 5, ""},
	}
	for _, tc := range tests {
		got := slice.Stripe(makeInput(tc.input), tc.i)
		if diff := cmp.Diff(got, split(tc.want), cmpopts.EquateEmpty()); diff != "" {
			t.Errorf("Strip %d (-got, want):\ninput:\n%s\n%s",
				tc.i, strings.ReplaceAll(tc.input, "|", "\n"), diff)
		}
	}
}

func TestHead(t *testing.T) {
	tests := []struct {
		input string
		n     int
		want  string
	}{
		{"", 0, ""},
		{"", 1, ""},
		{"a", 1, "a"},

		{"a b c", 0, ""},
		{"a b c", 1, "a"},
		{"a b c", 2, "a b"},
		{"a b c", 3, "a b c"},
		{"a b c", 4, "a b c"},
	}
	for _, tc := range tests {
		got := slice.Head(strings.Fields(tc.input), tc.n)
		if diff := cmp.Diff(got, strings.Fields(tc.want)); diff != "" {
			t.Errorf("Head %d (-got, +want):\n%s", tc.n, diff)
		}
	}
}

func TestTail(t *testing.T) {
	tests := []struct {
		input string
		n     int
		want  string
	}{
		{"", 0, ""},
		{"", 1, ""},
		{"a", 1, "a"},

		{"a b c", 0, ""},
		{"a b c", 1, "c"},
		{"a b c", 2, "b c"},
		{"a b c", 3, "a b c"},
		{"a b c", 4, "a b c"},
	}
	for _, tc := range tests {
		got := slice.Tail(strings.Fields(tc.input), tc.n)
		if diff := cmp.Diff(got, strings.Fields(tc.want)); diff != "" {
			t.Errorf("Tail %d (-got, +want):\n%s", tc.n, diff)
		}
	}
}

func TestSelect(t *testing.T) {
	tests := []struct {
		input, want []int
	}{
		{nil, nil},
		{[]int{}, nil},
		{[]int{1}, nil},
		{[]int{4}, []int{4}},
		{[]int{1, 4}, []int{4}},
		{[]int{4, 1}, []int{4}},
		{[]int{1, 2, 3}, []int{2}},
		{[]int{1, 2, 3, 2}, []int{2, 2}},
		{[]int{8, 6, 7, 5, 3, 0, 9}, []int{8, 6, 0}},
		{[]int{8, 8, 1, 1, 2, 3, 2}, []int{8, 8, 2, 2}},
	}

	isEven := func(z int) bool { return z%2 == 0 }
	for _, tc := range tests {
		got := slices.Collect(slice.Select(tc.input, isEven))
		if diff := cmp.Diff(got, tc.want); diff != "" {
			t.Errorf("Select %v (-got, +want):\n%s", tc.input, diff)
		}
	}
}

func TestMap(t *testing.T) {
	strlen := func(s string) int { return len(s) }
	atoi := func(s string) int { v, _ := strconv.Atoi(s); return v }
	tests := []struct {
		input []string
		f     func(string) int
		want  []int
	}{
		{nil, strlen, nil},
		{[]string{}, strlen, []int{}},
		{[]string{"a", "be", "sea"}, strlen, []int{1, 2, 3}},
		{[]string{"23", "37", "-59"}, atoi, []int{23, 37, -59}},
		{[]string{"xx", "15", "yy", "30"}, atoi, []int{0, 15, 0, 30}},
	}
	for _, tc := range tests {
		got := slice.Map(tc.input, tc.f)
		if diff := cmp.Diff(got, tc.want); diff != "" {
			t.Errorf("Map %q (-got, +want):\n%s", tc.input, diff)
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

	// Verify that the output is clipped to its length.
	cp2 := copyOf(cp)
	var zero T
	_ = append(got, zero)
	if diff := cmp.Diff(cp, cp2); diff != "" {
		t.Errorf("After append to result (-got, +want):\n%s", diff)
	}
}

func copyOf[T any](vs []T) []T {
	out := make([]T, len(vs))
	copy(out, vs)
	return out
}
