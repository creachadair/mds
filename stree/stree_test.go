package stree_test

import (
	"sort"
	"strings"
	"testing"

	"github.com/creachadair/mds/mapset"
	"github.com/creachadair/mds/stree"
	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
)

// Export all the words in tree in their stored order.
func allWords(tree *stree.Tree[string]) []string {
	var got []string
	tree.Inorder(func(key string) bool {
		got = append(got, key)
		return true
	})
	return got
}

func sortedUnique(ws []string, drop mapset.Set[string]) []string {
	out := mapset.New[string](ws...).RemoveAll(drop).Slice()
	sort.Strings(out)
	return out
}

func TestNew(t *testing.T) {
	t.Run("Empty", func(t *testing.T) {
		tree := stree.New(100, stringLess)
		if n := tree.Len(); n != 0 {
			t.Errorf("Len of empty tree: got %v, want 0", n)
		}
		if !tree.IsEmpty() {
			t.Error("IsEmpty should be true for an empty tree")
		}
	})
	t.Run("NonEmpty", func(t *testing.T) {
		tree := stree.New(200, stringLess, "please", "fetch", "your", "slippers")
		got := allWords(tree)
		want := []string{"fetch", "please", "slippers", "your"}
		if diff := cmp.Diff(got, want); diff != "" {
			t.Errorf("New<Tree produced unexpected output (-got, +want)\n%s", diff)
		}

		if n := tree.Len(); n != len(want) {
			t.Errorf("Len: got %d, want %d", n, len(want))
		}
	})
}

func TestRemoval(t *testing.T) {
	words := strings.Fields(`a foolish consistency is the hobgoblin of little minds`)
	tree := stree.New[string](0, stringLess, words...)

	got := allWords(tree)
	want := sortedUnique(words, nil)
	if diff := cmp.Diff(want, got); diff != "" {
		t.Errorf("Original input differs from expected (-want, +got)\n%s", diff)
	}
	drop := mapset.New("a", "is", "of", "the")
	for w := range drop {
		if !tree.Remove(w) {
			t.Errorf("Remove(%q) returned false, wanted true", w)
		}
	}

	got = allWords(tree)
	want = sortedUnique(words, drop)
	if diff := cmp.Diff(want, got); diff != "" {
		t.Errorf("Tree after removal is incorrect (-want, +got)\n%s", diff)
	}
}

func TestInsertion(t *testing.T) {
	type kv struct {
		key string
		val int
	}

	tree := stree.New[kv](300, func(a, b kv) bool {
		return a.key < b.key
	})
	checkInsert := func(f func(kv) bool, key string, val int, ok bool) {
		t.Helper()
		got := f(kv{key, val})
		if got != ok {
			t.Errorf("Add(%q, %v): got %v, want %v", key, val, got, ok)
		}
	}
	checkValue := func(key string, want int) {
		got, ok := tree.Get(kv{key: key})
		if !ok {
			t.Errorf("Key %q not found", key)
		} else if got.val != want {
			t.Errorf("Key %q: got %v, want %v", key, got.val, want)
		}
	}

	checkInsert(tree.Add, "x", 2, true)
	checkValue("x", 2)
	checkInsert(tree.Add, "x", 3, false)
	checkValue("x", 2)
	checkInsert(tree.Replace, "x", 5, false)
	checkValue("x", 5)
	checkInsert(tree.Replace, "y", 7, true)
	checkValue("y", 7)
	checkInsert(tree.Add, "y", 0, false)
	checkValue("y", 7)
}

func TestInorderAfter(t *testing.T) {
	keys := []string{"8", "6", "7", "5", "3", "0", "9"}
	tree := stree.New(0, stringLess, keys...)
	tests := []struct {
		key  string
		want string
	}{
		{"A", ""},
		{"9", "9"},
		{"8", "8 9"},
		{"7", "7 8 9"},
		{"6", "6 7 8 9"},
		{"5", "5 6 7 8 9"},
		{"4", "5 6 7 8 9"},
		{"3", "3 5 6 7 8 9"},
		{"2", "3 5 6 7 8 9"},
		{"1", "3 5 6 7 8 9"},
		{"0", "0 3 5 6 7 8 9"},
		{"", "0 3 5 6 7 8 9"},
	}
	for _, test := range tests {
		want := strings.Fields(test.want)
		var got []string
		tree.InorderAfter(test.key, func(key string) bool {
			got = append(got, key)
			return true
		})
		if diff := cmp.Diff(want, got, cmpopts.EquateEmpty()); diff != "" {
			t.Errorf("InorderAfter(%v) result differed from expected\n%s", test.key, diff)
		}
	}
}

func TestCursor(t *testing.T) {
	t.Run("EmptyTree", func(t *testing.T) {
		tree := stree.New(250, stringLess)

		// An empty tree reports nil cursors.
		if got := tree.Cursor("whatever"); got.Valid() {
			t.Errorf("Cursor on empty tree: got %v, want nil", got)
		} else if key := got.Key(); key != "" {
			t.Errorf("Nil cursor key: got %q, want empty", key)
		}
		if got := tree.Root(); got.Valid() {
			t.Errorf("Root on empty tree: got %v, want invalid", got)
		}
		if got := tree.Root().Min(); got.Valid() {
			t.Errorf("Min on empty tree: got %v, want invalid", got)
		}
		if got := tree.Root().Max(); got.Valid() {
			t.Errorf("Max on empty tree: got %v, want invalid", got)
		}
	})

	tree := stree.New(250, stringLess, "a", "b", "c", "d", "e", "f", "g")
	t.Run("Iterate", func(t *testing.T) {
		type tcursor = *stree.Cursor[string]

		var cur tcursor
		prev := (tcursor).Prev
		next := (tcursor).Next
		jump := func(c tcursor) func(tcursor) tcursor {
			return func(tcursor) tcursor { cur = c; return c }
		}
		pgm := []struct {
			step  func(tcursor) tcursor
			valid bool
			key   string
		}{
			{jump(tree.Root().Min()), true, "a"},
			{next, true, "b"},
			{prev, true, "a"},
			{prev, false, ""},

			{jump(tree.Root().Min()), true, "a"},
			{next, true, "b"},
			{prev, true, "a"},
			{next, true, "b"},
			{next, true, "c"},
			{next, true, "d"},
			{prev, true, "c"},

			{jump(tree.Root().Max()), true, "g"},
			{next, false, ""},

			{jump(tree.Root().Max()), true, "g"},
			{prev, true, "f"},
			{prev, true, "e"},
			{prev, true, "d"},
			{prev, true, "c"},

			{jump(tree.Cursor("d")), true, "d"},
			{next, true, "e"},
			{prev, true, "d"},
			{prev, true, "c"},

			{jump(tree.Root().Min()), true, "a"},
			{prev, false, ""},
		}
		for i, in := range pgm {
			if got := in.step(cur); got.Valid() != in.valid {
				t.Errorf("Step %d: got %v (%v), want %v", i+1, got, got.Valid(), in.valid)
			}
			if key := cur.Key(); key != in.key {
				t.Errorf("Step %d: got key %q, want %q", i+1, key, in.key)
			}
		}
	})

	t.Run("Navigate", func(t *testing.T) {
		root := tree.Root()
		t.Logf("Root: %q", root.Key())

		t.Run("Right", func(t *testing.T) {
			var got []string
			for r := root.Clone().Right().Min(); r.Valid(); r.Next() {
				got = append(got, r.Key())
			}
			if diff := cmp.Diff(got, []string{"e", "f", "g"}); diff != "" {
				t.Errorf("Right tree (-got, +want):\n%s", diff)
			}
		})
		t.Run("Left", func(t *testing.T) {
			var got []string
			for l := root.Clone().Left().Max(); l.Valid(); l.Prev() {
				got = append(got, l.Key())
			}
			if diff := cmp.Diff(got, []string{"c", "b", "a"}); diff != "" {
				t.Errorf("Left tree (-got, +want):\n%s", diff)
			}
		})
	})

	t.Run("Traverse", func(t *testing.T) {
		var got []string
		tree.Cursor("f").Inorder(func(s string) bool {
			got = append(got, s)
			return true
		})
		if diff := cmp.Diff(got, []string{"e", "f", "g"}); diff != "" {
			t.Errorf("Right tree (-got, +want):\n%s", diff)
		}
	})
}
