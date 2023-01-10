package stree

import (
	"flag"
	"fmt"
	"io"
	"log"
	"os"
	"sort"
	"strings"
	"testing"

	"github.com/creachadair/mds/mapset"
	"github.com/google/go-cmp/cmp"
)

var (
	strictness = flag.Int("balance", 100, "Balancing factor")
	textFile   = flag.String("text", "cask.txt", "Text file to read for basic testing")
	dotFile    = flag.String("dot", "", "Emit DOT output to this file")
	sortWords  = flag.Bool("sort", false, "Sort input words before insertion")
)

func lessString(a, b string) bool { return a < b }

func sortedUnique(ws []string) []string {
	out := mapset.New[string](ws...).Slice()
	sort.Strings(out)
	return out
}

// Construct a tree with the words from input, returning the finished tree and
// the original words as split by strings.Fields.
func makeTree(β int, input string) (*Tree[string], []string) {
	tree := New(β, lessString)
	words := strings.Fields(input)
	if *sortWords {
		sort.Strings(words)
	}
	for _, w := range words {
		tree.Add(w)
	}
	return tree, words
}

// Export all the words in tree in their stored order.
func allWords(tree *Tree[string]) []string {
	var got []string
	tree.Inorder(func(key string) bool {
		got = append(got, key)
		return true
	})
	return got
}

func (n *node[T]) height() int {
	if n == nil {
		return 0
	}
	h := n.left.height()
	if r := n.right.height(); r > h {
		h = r
	}
	return h + 1
}

func TestBasicProperties(t *testing.T) {
	// http://www.gutenberg.org/files/1063/1063-h/1063-h.htm
	text, err := os.ReadFile(*textFile)
	if err != nil {
		t.Fatalf("Reading text: %v", err)
	}
	tree, words := makeTree(*strictness, string(text))
	t.Logf("%v has height %d", tree, tree.root.height())
	dumpTree(tree)

	got := allWords(tree)
	want := sortedUnique(words)
	if diff := cmp.Diff(want, got); diff != "" {
		t.Errorf("Inorder produced unexpected output (-want, +got)\n%s", diff)
	}

	// Verify that the values are of the correct type.
	for _, word := range words {
		if _, ok := tree.Get(word); !ok {
			t.Errorf("Word %q not found", word)
		}
	}

	// Verify that clearing a tree leaves it empty.
	if n := tree.Len(); n != len(want) {
		t.Errorf("Len: got %d, want %d", n, len(want))
	}
	tree.Clear()
	if !tree.IsEmpty() {
		t.Error("IsEmpty should report true after Clear")
	}
	if n := tree.Len(); n != 0 {
		t.Errorf("Len: got %d, want 0", n)
	}

	tree.Add("FINIS")
	if n := tree.Len(); n != 1 {
		t.Errorf("Len: got %d, want 1", n)
	}
	if tree.IsEmpty() {
		t.Error("IsEmpty should report false after Add")
	}
}

// If an output file is specified, dump a DOT graph of tree.
func dumpTree(tree *Tree[string]) {
	if *dotFile == "" {
		return
	}
	f, err := os.Create(*dotFile)
	if err != nil {
		log.Fatalf("Unable to create DOT output: %v", err)
	}
	dotTree(f, tree.root)
	if err := f.Close(); err != nil {
		log.Fatalf("Unable to close output: %v", err)
	}
}

// Render tree to a GraphViz graph.
func dotTree(w io.Writer, root *node[string]) {
	fmt.Fprintln(w, "digraph Tree {")

	i := 0
	next := func() int {
		i++
		return i
	}

	var ptree func(*node[string]) int
	ptree = func(root *node[string]) int {
		if root == nil {
			return 0
		}
		id := next()
		fmt.Fprintf(w, "\tN%04d [label=\"%s\"]\n", id, root.X)
		if lc := ptree(root.left); lc != 0 {
			fmt.Fprintf(w, "\tN%04d -> N%04d\n", id, lc)
		}
		if rc := ptree(root.right); rc != 0 {
			fmt.Fprintf(w, "\tN%04d -> N%04d\n", id, rc)
		}
		return id
	}
	ptree(root)
	fmt.Fprintln(w, "}")
}
