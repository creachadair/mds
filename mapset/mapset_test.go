package mapset_test

import (
	"testing"

	"github.com/creachadair/mds/mapset"
	"github.com/google/go-cmp/cmp"
)

func check[T comparable](t *testing.T, s mapset.Set[T], want ...T) mapset.Set[T] {
	t.Helper()
	m := make(mapset.Set[T], len(want))
	for _, w := range want {
		m[w] = struct{}{}
	}
	if n := s.Len(); n != len(m) {
		t.Errorf("Wrong length: got %d, want %d", n, len(m))
	}
	if diff := cmp.Diff(m, s); diff != "" {
		t.Errorf("Wrong contents (-want, +got):\n%s", diff)
	}
	return s
}

func TestIntersect(t *testing.T) {
	e1 := check(t, mapset.New[int]())
	nat := check(t, mapset.New(1, 2, 3, 4, 5, 6), 1, 2, 3, 4, 5, 6)
	odd := check(t, mapset.New(1, 3, 5, 7, 9, 11), 1, 3, 5, 7, 9, 11)
	evn := check(t, mapset.New(2, 4, 6, 8, 10), 2, 4, 6, 8, 10)
	prm := check(t, mapset.New(2, 3, 5, 7, 11, 13), 2, 3, 5, 7, 11, 13)

	tests := []struct {
		lhs, rhs mapset.Set[int]
		want     []int
	}{
		// Various empty cases.
		{nil, nil, nil},
		{e1, e1, nil},
		{e1, nil, nil},
		{nil, e1, nil},

		// Intersection with empty is empty.
		{nat, nil, nil},
		{nil, nat, nil},

		// Intersection with disjoint is empty.
		{odd, evn, nil},
		{evn, odd, nil},

		// Intersection should be commutative.
		{nat, odd, []int{1, 3, 5}},
		{odd, nat, []int{1, 3, 5}},
		{nat, prm, []int{2, 3, 5}},
		{odd, prm, []int{3, 5, 7, 11}},

		// Self-intersection should be self.
		{nat, nat, nat.Slice()},
		{odd, odd, odd.Slice()},
	}
	for _, tc := range tests {
		got := mapset.Intersect(tc.lhs, tc.rhs)
		check(t, got, tc.want...)
	}
}

func TestBasic(t *testing.T) {
	t.Run("Empty", func(t *testing.T) {
		check(t, mapset.New[string]())
		check(t, mapset.Set[int]{})
		check(t, mapset.NewSize[struct{ A string }](100))
	})

	t.Run("New", func(t *testing.T) {
		check(t, mapset.New("a", "b"), "a", "b")
		check(t, mapset.New(1, 2, 3), 1, 2, 3)
	})

	t.Run("Clone", func(t *testing.T) {
		inputs := []string{"apple", "pear", "plum", "cherry"}
		s1 := mapset.New(inputs...)
		s2 := s1.Clone()
		check(t, s1, inputs...)
		check(t, s2, inputs...)

		var s3 mapset.Set[int]
		if c := s3.Clone(); c == nil {
			t.Error("Clone of nil should not be nil")
		}

		s4 := mapset.New[float64]()
		if s4 == nil {
			t.Error("New should not return nil")
		}
		if c := s4.Clone(); c == nil {
			t.Error("Clone of empty should not be nil")
		}
	})

	t.Run("Clear", func(t *testing.T) {
		s := check(t, mapset.New("a", "man", "a", "plan", "a", "canal"), "a", "man", "plan", "canal")
		check(t, s.Clear())
		if len(s) != 0 {
			t.Error("Length after clear should be zero")
		}
	})

	t.Run("Slice", func(t *testing.T) {
		tests := [][]int{
			nil,
			{},
			{1, 2, 3},
			{4, 1, 8, 9},
			{2, 2, 3, 2, 2},
		}
		for _, in := range tests {
			s := check(t, mapset.New(in...), in...)
			check(t, s, s.Slice()...)
		}
	})
}

func TestItems(t *testing.T) {
	t.Run("Add", func(t *testing.T) {
		s := check(t, mapset.New(2, 3, 5, 7), 2, 3, 5, 7)
		check(t, s.Add(2, 5, 11, 13), 2, 3, 5, 7, 11, 13)
		check(t, s, 2, 3, 5, 7, 11, 13)
	})

	t.Run("Remove", func(t *testing.T) {
		s := check(t, mapset.New(2, 3, 5, 7, 11, 13, 17), 2, 3, 5, 7, 11, 13, 17)
		check(t, s.Remove(13, 17, 2, 8, 4, 1), 3, 5, 7, 11)
		check(t, s, 3, 5, 7, 11)
	})

	t.Run("AddAll", func(t *testing.T) {
		s1 := check(t, mapset.New(1, 3, 5, 7), 1, 3, 5, 7)
		s2 := check(t, mapset.New(2, 4, 6), 2, 4, 6)

		check(t, s1.AddAll(s2), 1, 2, 3, 4, 5, 6, 7)
		check(t, s1, 1, 2, 3, 4, 5, 6, 7)
		check(t, s2, 2, 4, 6)
	})

	t.Run("RemoveAll", func(t *testing.T) {
		s1 := check(t, mapset.New(1, 2, 3, 4, 5), 1, 2, 3, 4, 5)
		s2 := check(t, mapset.New(2, 4, 6, 8, 10), 2, 4, 6, 8, 10)

		check(t, s1.RemoveAll(s2), 1, 3, 5)
		check(t, s1, 1, 3, 5)
		check(t, s2, 2, 4, 6, 8, 10)
	})

	t.Run("Pop", func(t *testing.T) {
		s := check(t, mapset.New(1, 2), 1, 2)
		if got := s.Pop(); got != 1 && got != 2 {
			t.Errorf("Pop: got %d, want 1 or 2", got)
		}
		if s.Len() != 1 {
			t.Errorf("Length after Pop: got %d, want 1", s.Len())
		}
	})
}

func TestCompare(t *testing.T) {
	t.Run("Intersects", func(t *testing.T) {
		s1 := check(t, mapset.New(1, 2, 3, 4), 1, 2, 3, 4)

		if s2 := check(t, mapset.New[int]()); s2.Intersects(s1) || s1.Intersects(s2) {
			t.Error("Empty set should not intersect")
		}

		if s2 := check(t, mapset.New(3, 5, 7), 3, 5, 7); !s2.Intersects(s1) || !s1.Intersects(s2) {
			t.Error("Sets should intersect")
		}

		if s2 := check(t, mapset.New(6, 8, 10), 6, 8, 10); s2.Intersects(s1) || s1.Intersects(s2) {
			t.Error("Sets should not intersect")
		}
	})

	t.Run("HasAll", func(t *testing.T) {
		s1 := check(t, mapset.New(1, 2, 3, 4), 1, 2, 3, 4)

		if s0 := check(t, mapset.New[int]()); !s0.HasAll() {
			t.Error("Empty set should have all of no elements")
		} else if s0.HasAll(1) {
			t.Error("Empty set should have none of no elements")
		}

		if !s1.HasAll() {
			t.Error("Empty elements should report true")
		}
		if !s1.HasAll(1, 3) {
			t.Errorf("Set %v should contain 1 and 3", s1)
		}
		if s1.HasAll(1, 4, 5) {
			t.Errorf("Set %v should not contain 5", s1)
		}
	})

	t.Run("HasAny", func(t *testing.T) {
		s1 := check(t, mapset.New(1, 2, 3, 4), 1, 2, 3, 4)
		var s2 mapset.Set[int]

		if check(t, mapset.New[int]()).HasAny(1) {
			t.Error("Empty set should not have any elements")
		}
		if s2.HasAny(1) {
			t.Error("Nil set should not have any elements")
		}
		if s1.HasAny() {
			t.Error("Empty elements should report false")
		}
		if !s1.HasAny(9, 2, 8, 4) {
			t.Errorf("Set %v should contain 2 and 4", s1)
		}
		if s1.HasAny(7, 5, 6) {
			t.Errorf("Set %v should not contain 7, 5, or 6", s1)
		}
	})

	t.Run("IsSubset", func(t *testing.T) {
		s1 := check(t, mapset.New(1, 2, 3, 4, 5, 6, 7), 1, 2, 3, 4, 5, 6, 7)
		s2 := check(t, mapset.New(2, 3, 5, 7), 2, 3, 5, 7)
		s3 := check(t, mapset.New(1, 3, 9), 1, 3, 9)
		s4 := check(t, mapset.New[int]())

		tests := []struct {
			a, b mapset.Set[int]
			want bool
		}{
			{s1, s1, true},
			{s1, s2, false},
			{s2, s1, true},
			{s3, s1, false},
			{s3, s2, false},
			{s2, s3, false},
			{s4, s1, true},
			{s1, s4, false},
			{s4, s4, true},
		}

		for _, test := range tests {
			if eq := test.a.IsSubset(test.b); eq != test.want {
				t.Errorf("IsSubset: got %v, want %v\na = %+v\nb = %+v", eq, test.want, test.a, test.b)
			}
		}
	})

	t.Run("Equals", func(t *testing.T) {
		s1 := check(t, mapset.New(1, 2, 3), 1, 2, 3)
		t.Logf("Test needle: %+v", s1)

		s2 := s1.Clone()
		s3 := mapset.New(1, 2, 3, 4)
		s4 := mapset.New(1, 2, 3)
		s5 := mapset.New(2, 3)
		s6 := mapset.New(1, 2, 4)

		tests := []struct {
			a, b mapset.Set[int]
			want bool
		}{
			{s1, s2, true},
			{s1, s3, false},
			{s1, s4, true},
			{s1, s5, false},
			{s1, s6, false},
			{s4, s6, false},

			// Various permutations of empty.
			{mapset.Set[int](nil), mapset.Set[int](nil), true}, // both nil
			{mapset.New[int](), mapset.New[int](), true},       // both non-nil
			{mapset.Set[int](nil), mapset.New[int](), true},    // one nil, one not
		}
		for _, test := range tests {
			if eq := test.a.Equals(test.b); eq != test.want {
				t.Errorf("Equals: got %v, want %v\na = %+v\nb = %+v", eq, test.want, test.a, test.b)
			}
			if eq := test.b.Equals(test.a); eq != test.want {
				t.Errorf("Equals: got %v, want %v\na = %+v\nb = %+v", eq, test.want, test.a, test.b)
			}
		}
	})
}

func TestKeys(t *testing.T) {
	t.Run("Nil", func(t *testing.T) {
		check(t, mapset.Keys(map[string]int(nil)))
	})
	t.Run("Empty", func(t *testing.T) {
		check(t, mapset.Keys(map[string]int{}))
	})
	t.Run("String", func(t *testing.T) {
		check(t, mapset.Keys(map[string]bool{
			"a": true, "b": false, "c": true,
		}), "a", "b", "c")
	})
	t.Run("Int", func(t *testing.T) {
		check(t, mapset.Keys(map[int]string{
			1: "red", 2: "orange", 3: "yellow",
		}), 1, 2, 3)
	})
}

func TestValues(t *testing.T) {
	t.Run("Nil", func(t *testing.T) {
		check(t, mapset.Values(map[string]int(nil)))
	})
	t.Run("Empty", func(t *testing.T) {
		check(t, mapset.Values(map[string]int{}))
	})
	t.Run("String", func(t *testing.T) {
		check(t, mapset.Values(map[int]string{
			1: "foo", 2: "bar", 3: "baz", 4: "foo", 5: "quux",
		}), "foo", "bar", "baz", "quux")
	})
	t.Run("Int", func(t *testing.T) {
		check(t, mapset.Values(map[string]int{
			"red": 1, "green": 1, "blue": 2, "white": 2,
		}), 1, 2)
	})
}
