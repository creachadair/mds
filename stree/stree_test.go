package stree_test

import (
	"cmp"
	"sort"
	"strings"
	"testing"

	"github.com/creachadair/mds/mapset"
	"github.com/creachadair/mds/stree"
	gocmp "github.com/google/go-cmp/cmp"
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
	out := mapset.New(ws...).RemoveAll(drop).Slice()
	sort.Strings(out)
	return out
}

func TestNew(t *testing.T) {
	t.Run("Empty", func(t *testing.T) {
		tree := stree.New(100, cmp.Compare[string])
		if n := tree.Len(); n != 0 {
			t.Errorf("Len of empty tree: got %v, want 0", n)
		}
		if !tree.IsEmpty() {
			t.Error("IsEmpty should be true for an empty tree")
		}
	})
	t.Run("NonEmpty", func(t *testing.T) {
		tree := stree.New(200, cmp.Compare[string], "please", "fetch", "your", "slippers")
		got := allWords(tree)
		want := []string{"fetch", "please", "slippers", "your"}
		if diff := gocmp.Diff(got, want); diff != "" {
			t.Errorf("New<Tree produced unexpected output (-got, +want)\n%s", diff)
		}

		if n := tree.Len(); n != len(want) {
			t.Errorf("Len: got %d, want %d", n, len(want))
		}
	})
}

func TestRemoval(t *testing.T) {
	words := strings.Fields(`a foolish consistency is the hobgoblin of little minds`)
	tree := stree.New(0, cmp.Compare, words...)

	got := allWords(tree)
	want := sortedUnique(words, nil)
	if diff := gocmp.Diff(want, got); diff != "" {
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
	if diff := gocmp.Diff(want, got); diff != "" {
		t.Errorf("Tree after removal is incorrect (-want, +got)\n%s", diff)
	}
}

func TestInsertion(t *testing.T) {
	type kv = stree.KV[string, int]

	tree := stree.New(300, kv{}.Compare(cmp.Compare))
	checkInsert := func(f func(kv) bool, key string, val int, ok bool) {
		t.Helper()
		got := f(kv{key, val})
		if got != ok {
			t.Errorf("Add(%q, %v): got %v, want %v", key, val, got, ok)
		}
	}
	checkValue := func(key string, want int) {
		got, ok := tree.Get(kv{Key: key})
		if !ok {
			t.Errorf("Key %q not found", key)
		} else if got.Value != want {
			t.Errorf("Key %q: got %v, want %v", key, got.Value, want)
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
	tree := stree.New(0, cmp.Compare[string], keys...)
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
		if diff := gocmp.Diff(want, got, cmpopts.EquateEmpty()); diff != "" {
			t.Errorf("InorderAfter(%v) result differed from expected\n%s", test.key, diff)
		}
	}
}

func TestCursor(t *testing.T) {
	t.Run("EmptyTree", func(t *testing.T) {
		tree := stree.New(250, strings.Compare)

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

	tree := stree.New(250, strings.Compare, "a", "b", "c", "d", "e", "f", "g")
	t.Run("Iterate", func(t *testing.T) {
		type tcursor = *stree.Cursor[string]

		var cur tcursor
		prev := (tcursor).Prev
		left := (tcursor).Left
		next := (tcursor).Next
		right := (tcursor).Right
		up := (tcursor).Up

		// A pseudo-operation that ignores its input, and updates cur to a new
		// cursor and returns that cursor.
		jump := func(c tcursor) func(tcursor) tcursor {
			return func(tcursor) tcursor { cur = c; return c }
		}

		// Each step of this test is an instruction that applies a step to cur
		// and specifies whether the result cursor should be valid, and what the
		// key reported by its Key method should be.
		pgm := []struct {
			step  func(tcursor) tcursor
			valid bool
			key   string
		}{
			// Initially, cur should be invalid.
			{jump(cur), false, ""},

			// Navigation from an invalid cursor should not panic, but should
			// leave the cursor invalid.
			{up, false, ""},
			{next, false, ""},
			{prev, false, ""},
			{left, false, ""},
			{right, false, ""},

			// Go to the root and navigate around it.
			{jump(tree.Root()), true, "d"},
			{next, true, "e"},
			{prev, true, "d"},
			{prev, true, "c"},

			// Check left and right children.
			{jump(tree.Root()), true, "d"},
			{left, true, "b"},
			{right, true, "c"},
			{left, false, ""},

			// Navigate up toward the root.
			{jump(tree.Root().Min()), true, "a"},
			{up, true, "b"},
			{up, true, "d"},
			{up, false, ""},

			// Walk off the start to invalid
			{jump(tree.Root().Min()), true, "a"},
			{prev, false, ""},

			// Navigate around the min.
			{jump(tree.Root().Min()), true, "a"},
			{next, true, "b"},
			{prev, true, "a"},
			{next, true, "b"},
			{next, true, "c"},
			{next, true, "d"},
			{next, true, "e"}, // cross the root successfully
			{prev, true, "d"},

			// Walk off the end to invalid.
			{jump(tree.Root().Max()), true, "g"},
			{next, false, ""},

			// Navigate around the max.
			{jump(tree.Root().Max()), true, "g"},
			{prev, true, "f"},
			{prev, true, "e"},
			{prev, true, "d"},
			{prev, true, "c"},

			// Find a non-existing element.
			{jump(tree.Cursor("nonesuch")), false, ""},

			// Find an existing element.
			{jump(tree.Cursor("d")), true, "d"},
			{next, true, "e"},
			{prev, true, "d"},
			{prev, true, "c"},
			{up, true, "b"},
			{left, true, "a"},
			{up, true, "b"},
			{right, true, "c"},
			{right, false, ""},

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

		t.Run("Forward", func(t *testing.T) {
			var got []string
			for r := tree.Cursor("f").Min(); r.Valid(); r.Next() {
				got = append(got, r.Key())
			}
			if diff := gocmp.Diff(got, []string{"e", "f", "g"}); diff != "" {
				t.Errorf("Forward walk (-got, +want):\n%s", diff)
			}
		})
		t.Run("Reverse", func(t *testing.T) {
			var got []string
			for l := tree.Cursor("b").Max(); l.Valid(); l.Prev() {
				got = append(got, l.Key())
			}
			if diff := gocmp.Diff(got, []string{"c", "b", "a"}); diff != "" {
				t.Errorf("Reverse walk (-got, +want):\n%s", diff)
			}
		})
	})

	t.Run("Traverse", func(t *testing.T) {
		var got []string
		tree.Cursor("f").Inorder(func(s string) bool {
			got = append(got, s)
			return true
		})
		if diff := gocmp.Diff(got, []string{"e", "f", "g"}); diff != "" {
			t.Errorf("Right tree (-got, +want):\n%s", diff)
		}
	})
}

func TestKV(t *testing.T) {
	type kv = stree.KV[string, int]
	compare := kv{}.Compare(cmp.Compare)

	st := stree.New(250, compare, []kv{
		{"hello", 1},
		{"is", 2},
		{"there", 3},
		{"anybody", 4},
		{"in", 5},
		{"here", 6},
	}...)

	var gotk []string
	var gotv []int
	st.Inorder(func(kv kv) bool {
		gotk = append(gotk, kv.Key)
		gotv = append(gotv, kv.Value)
		return true
	})

	if diff := gocmp.Diff(gotk, []string{"anybody", "hello", "here", "in", "is", "there"}); diff != "" {
		t.Errorf("Keys (-got, +want):\n%s", diff)
	}
	if diff := gocmp.Diff(gotv, []int{4, 1, 6, 5, 2, 3}); diff != "" {
		t.Errorf("Values (-got, +want):\n%s", diff)
	}
}
