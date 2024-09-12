// Copyright (c) 2015, Michael J. Fromberger

package shell

import (
	"fmt"
	"testing"
)

func TestQuote(t *testing.T) {
	type testCase struct{ in, want string }
	tests := []testCase{
		{"", "''"},            // empty is special
		{"abc", "abc"},        // nothing to quote
		{"--flag", "--flag"},  // "
		{"'abc", `\'abc`},     // single quote only
		{"abc'", `abc\'`},     // "
		{`shan't`, `shan\'t`}, // "
		{"--flag=value", `'--flag=value'`},
		{"a b\tc", "'a b\tc'"},
		{`a"b"c`, `'a"b"c'`},
		{`'''`, `\'\'\'`},
		{`\`, `'\'`},
		{`'a=b`, `\''a=b'`},   // quotes and other stuff
		{`a='b`, `'a='\''b'`}, // "
		{`a=b'`, `'a=b'\'`},   // "
	}
	// Verify that all the designated special characters get quoted.
	for _, c := range shouldQuote + mustQuote {
		tests = append(tests, testCase{
			in:   string(c),
			want: fmt.Sprintf(`'%c'`, c),
		})
	}

	for _, test := range tests {
		got := Quote(test.in)
		if got != test.want {
			t.Errorf("Quote %q: got %q, want %q", test.in, got, test.want)
		}
	}
}
