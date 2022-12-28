package stree_test

import (
	"sort"
	"strings"
	"testing"

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

func sortedUnique(ws []string, drop func(string) bool) []string {
	m := make(map[string]struct{})
	for _, w := range ws {
		if drop == nil || !drop(w) {
			m[w] = struct{}{}
		}
	}
	out := make([]string, 0, len(m))
	for key := range m {
		out = append(out, key)
	}
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
	drop := map[string]bool{"a": true, "is": true, "of": true, "the": true}
	for w := range drop {
		if !tree.Remove(w) {
			t.Errorf("Remove(%q) returned false, wanted true", w)
		}
	}

	got = allWords(tree)
	want = sortedUnique(words, func(w string) bool {
		return drop[w]
	})
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
