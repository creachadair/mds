// Copyright (c) 2015, Michael J. Fromberger

package shell_test

import (
	"fmt"
	"io"
	"log"
	"strings"
	"testing"

	"github.com/creachadair/mds/shell"
	"github.com/google/go-cmp/cmp"
)

func TestSplit(t *testing.T) {
	tests := []struct {
		in   string
		want []string
		ok   bool
	}{
		// Variations of empty input yield an empty split.
		{"", nil, true},
		{"   ", nil, true},
		{"\t", nil, true},
		{"\n ", nil, true},

		// Various escape sequences work properly.
		{`\ `, []string{" "}, true},
		{`a\ `, []string{"a "}, true},
		{`\\a`, []string{`\a`}, true},
		{`"a\"b"`, []string{`a"b`}, true},
		{`'\'`, []string{"\\"}, true},

		// Leading and trailing whitespace are discarded correctly.
		{"a", []string{"a"}, true},
		{" a", []string{"a"}, true},
		{"a\n", []string{"a"}, true},

		// Escaped newlines are magic in the correct ways.
		{"a\\\nb", []string{"ab"}, true},
		{"a \\\n  b\tc", []string{"a", "b", "c"}, true},

		// Various splits with and without quotes.  Quoted whitespace is
		// preserved.
		{"a b c", []string{"a", "b", "c"}, true},
		{`a 'b c'`, []string{"a", "b c"}, true},
		{"\"a\nb\"cd e'f'", []string{"a\nbcd", "ef"}, true},
		{"'\n \t '", []string{"\n \t "}, true},

		// Quoted empty strings are preserved in various places.
		{"''", []string{""}, true},
		{"a ''", []string{"a", ""}, true},
		{" a \"\" b ", []string{"a", "", "b"}, true},
		{"'' a", []string{"", "a"}, true},

		// Unbalanced quotation marks and escapes are detected.
		{"\\", []string{""}, false},          // escape without a target
		{"'", []string{""}, false},           // unclosed single
		{`"`, []string{""}, false},           // unclosed double
		{`'\''`, []string{`\`}, false},       // unclosed connected double
		{`"\\" '`, []string{`\`, ``}, false}, // unclosed separate single
		{"a 'b c", []string{"a", "b c"}, false},
		{`a "b c`, []string{"a", "b c"}, false},
		{`a "b \"`, []string{"a", `b "`}, false},
	}
	for _, test := range tests {
		got, ok := shell.Split(test.in)
		if ok != test.ok {
			t.Errorf("Split %#q: got valid=%v, want %v", test.in, ok, test.ok)
		}
		if diff := cmp.Diff(test.want, got); diff != "" {
			t.Errorf("Split %#q: (-want, +got)\n%s", test.in, diff)
		}
	}
}

func TestScannerSplit(t *testing.T) {
	tests := []struct {
		in         string
		want, rest []string
	}{
		{"", nil, nil},
		{" ", nil, nil},
		{"--", nil, nil},
		{"a -- b", []string{"a"}, []string{"b"}},
		{"a b c -- d -- e ", []string{"a", "b", "c"}, []string{"d", "--", "e"}},
		{`"a b c --" -- "d "`, []string{"a b c --"}, []string{"d "}},
		{` -- "foo`, nil, []string{"foo"}}, // unterminated
		{"cmd -flag -- arg1 arg2", []string{"cmd", "-flag"}, []string{"arg1", "arg2"}},
	}
	for _, test := range tests {
		t.Logf("Scanner split input: %q", test.in)

		s := shell.NewScanner(strings.NewReader(test.in))
		var got, rest []string
		for s.Next() {
			if s.Text() == "--" {
				rest = s.Split()
				break
			}
			got = append(got, s.Text())
		}

		if s.Err() != io.EOF {
			t.Errorf("Unexpected scan error: %v", s.Err())
		}

		if diff := cmp.Diff(test.want, got); diff != "" {
			t.Errorf("Scanner split prefix: (-want, +got)\n%s", diff)
		}
		if diff := cmp.Diff(test.rest, rest); diff != "" {
			t.Errorf("Scanner split suffix: (-want, +got)\n%s", diff)
		}
	}
}

func TestRoundTrip(t *testing.T) {
	tests := [][]string{
		nil,
		{"a"},
		{"a "},
		{"a", "b", "c"},
		{"a", "b c"},
		{"--flag=value"},
		{"m='$USER'", "nop+", "$$"},
		{`"a" b `, "c"},
		{"odd's", "bodkins", "x'", "x''", "x\"\"", "$x':y"},
		{"a=b", "--foo", "${bar}", `\$`},
		{"cat", "a${b}.txt", "|", "tee", "capture", "2>", "/dev/null"},
	}
	for _, test := range tests {
		s := shell.Join(test)
		t.Logf("Join %#q = %v", test, s)
		got, ok := shell.Split(s)
		if !ok {
			t.Errorf("Split %+q: should be valid, but is not", s)
		}
		if diff := cmp.Diff(test, got); diff != "" {
			t.Errorf("Split %+q: (-want, +got)\n%s", s, diff)
		}
	}
}

func ExampleScanner() {
	const input = `a "free range" exploration of soi\ disant novelties`
	s := shell.NewScanner(strings.NewReader(input))
	sum, count := 0, 0
	for tok := range s.Each {
		count++
		sum += len(tok)
	}
	fmt.Println(len(input), count, sum, s.Complete(), s.Err())
	// Output: 51 6 43 true EOF
}

func ExampleScanner_Rest() {
	const input = `things 'and stuff' %end% all the remaining stuff`
	s := shell.NewScanner(strings.NewReader(input))
	for tok := range s.Each {
		if tok == "%end%" {
			fmt.Print("found marker; ")
			break
		}
	}
	rest, err := io.ReadAll(s.Rest())
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println(string(rest))
	// Output: found marker; all the remaining stuff
}

func ExampleScanner_Each() {
	const input = `a\ b 'c d' "e f's g" stop "go directly to jail"`
	s := shell.NewScanner(strings.NewReader(input))
	for tok := range s.Each {
		fmt.Println(tok)
		if tok == "stop" {
			break
		}
	}
	if err := s.Err(); err != nil {
		log.Fatal(err)
	}
	// Output:
	// a b
	// c d
	// e f's g
	// stop
}

func ExampleScanner_Split() {
	const input = `cmd -flag=t -- foo bar baz`

	s := shell.NewScanner(strings.NewReader(input))
	for s.Next() {
		if s.Text() == "--" {
			fmt.Println("** Args:", strings.Join(s.Split(), ", "))
		} else {
			fmt.Println(s.Text())
		}
	}
	// Output:
	// cmd
	// -flag=t
	// ** Args: foo, bar, baz
}
