// Copyright (c) 2015, Michael J. Fromberger

// Package shell supports splitting and joining of shell command strings.
//
// The Split function divides a string into whitespace-separated fields,
// respecting single and double quotation marks as defined by the Shell Command
// Language section of IEEE Std 1003.1 2013.  The Quote function quotes
// characters that would otherwise be subject to shell evaluation, and the Join
// function concatenates quoted strings with spaces between them.
//
// The relationship between Split and Join is that given
//
//	fields, ok := shell.Split(shell.Join(ss))
//
// the following relationship will hold:
//
//	fields == ss && ok
package shell

import (
	"bufio"
	"bytes"
	"io"
	"strings"
	"sync"
)

// These characters must be quoted to escape special meaning.  This list
// doesn't include the single quote.
const mustQuote = "|&;<>()$`\\\"\t\n"

// These characters should be quoted to escape special meaning, since in some
// contexts they are special (e.g., "x=y" in command position, "*" for globs).
const shouldQuote = `*?[#~=%`

// These are the separator characters in unquoted text.
const spaces = " \t\n"

const allQuote = mustQuote + shouldQuote + spaces

type state int

const (
	stNone state = iota
	stBreak
	stBreakQ
	stWord
	stWordQ
	stSingle
	stDouble
	stDoubleQ
)

type class int

const (
	clOther class = iota
	clBreak
	clNewline
	clQuote
	clSingle
	clDouble
)

type action int

const (
	drop action = iota
	push
	xpush
	emit
)

// N.B. Benchmarking shows that array lookup is substantially faster than map
// lookup here, but it requires caution when changing the state machine.  In
// particular:
//
// 1. The state and action values must be small integers.
// 2. The update table must completely cover the state values.
// 3. Each action slice must completely cover the action values.
var update = [...][]struct {
	state
	action
}{
	stNone: {},
	stBreak: {
		clBreak:   {stBreak, drop},
		clNewline: {stBreak, drop},
		clQuote:   {stBreakQ, drop},
		clSingle:  {stSingle, drop},
		clDouble:  {stDouble, drop},
		clOther:   {stWord, push},
	},
	stBreakQ: {
		clBreak:   {stWord, push},
		clNewline: {stBreak, drop},
		clQuote:   {stWord, push},
		clSingle:  {stWord, push},
		clDouble:  {stWord, push},
		clOther:   {stWord, push},
	},
	stWord: {
		clBreak:   {stBreak, emit},
		clNewline: {stBreak, emit},
		clQuote:   {stWordQ, drop},
		clSingle:  {stSingle, drop},
		clDouble:  {stDouble, drop},
		clOther:   {stWord, push},
	},
	stWordQ: {
		clBreak:   {stWord, push},
		clNewline: {stWord, drop},
		clQuote:   {stWord, push},
		clSingle:  {stWord, push},
		clDouble:  {stWord, push},
		clOther:   {stWord, push},
	},
	stSingle: {
		clBreak:   {stSingle, push},
		clNewline: {stSingle, push},
		clQuote:   {stSingle, push},
		clSingle:  {stWord, drop},
		clDouble:  {stSingle, push},
		clOther:   {stSingle, push},
	},
	stDouble: {
		clBreak:   {stDouble, push},
		clNewline: {stDouble, push},
		clQuote:   {stDoubleQ, drop},
		clSingle:  {stDouble, push},
		clDouble:  {stWord, drop},
		clOther:   {stDouble, push},
	},
	stDoubleQ: {
		clBreak:   {stDouble, xpush},
		clNewline: {stDouble, drop},
		clQuote:   {stDouble, push},
		clSingle:  {stDouble, xpush},
		clDouble:  {stDouble, push},
		clOther:   {stDouble, xpush},
	},
}

var classOf = [256]class{
	' ':  clBreak,
	'\t': clBreak,
	'\n': clNewline,
	'\\': clQuote,
	'\'': clSingle,
	'"':  clDouble,
}

// A Scanner partitions input from a reader into tokens divided on space, tab,
// and newline characters.  Single and double quotation marks are handled as
// described in http://pubs.opengroup.org/onlinepubs/9699919799/utilities/V3_chap02.html#tag_18_02.
type Scanner struct {
	buf *bufio.Reader
	cur bytes.Buffer
	st  state
	err error
}

// NewScanner returns a Scanner that reads input from r.
func NewScanner(r io.Reader) *Scanner {
	return &Scanner{buf: bufio.NewReader(r), st: stBreak}
}

// Reset discards the current token (if any) and all remaining input from s,
// and resets its internal state to consume tokens from r.
//
// Reset retains the internal buffers of s, so when processing multiple
// consecutive inputs, it will be more memory-efficient to read to allocate a
// single scanner and reset it for each reader in sequence.
func (s *Scanner) Reset(r io.Reader) {
	s.buf.Reset(r)
	s.cur.Reset()
	s.st = stBreak
	s.err = nil
}

// Next advances the scanner and reports whether there are any further tokens
// to be consumed.
func (s *Scanner) Next() bool {
	if s.err != nil {
		return false
	}
	s.cur.Reset()
	for {
		c, err := s.buf.ReadByte()
		s.err = err
		if err == io.EOF {
			break
		} else if err != nil {
			return false
		}
		next := update[s.st][classOf[c]]
		s.st = next.state
		switch next.action {
		case push:
			s.cur.WriteByte(c)
		case xpush:
			s.cur.Write([]byte{'\\', c})
		case emit:
			return true // s.cur has a complete token
		case drop:
			continue
		default:
			panic("unknown action")
		}
	}
	return s.st != stBreak
}

// Text returns the text of the current token, or "" if there is none.
func (s *Scanner) Text() string { return s.cur.String() }

// Err returns the error, if any, that resulted from the most recent action.
func (s *Scanner) Err() error { return s.err }

// Complete reports whether the current token is complete, meaning either that
// it is unquoted or that its quotes were balanced.
func (s *Scanner) Complete() bool { return s.st == stBreak || s.st == stWord }

// Rest returns an io.Reader for the remainder of the unconsumed input in s.
// After calling this method, Next will always return false.  The remainder
// does not include the text of the current token at the time Rest is called.
func (s *Scanner) Rest() io.Reader {
	s.st = stNone
	s.cur.Reset()
	s.err = io.EOF
	return s.buf
}

// Each is a range function that calls f for each token in the scanner.  It
// continues until the input is exhausted, f returns false, or an error other
// than [io.EOF] occurs. Use the Err method to check for an error.
func (s *Scanner) Each(f func(tok string) bool) {
	for s.Next() {
		if !f(s.Text()) {
			return
		}
	}
}

// Split returns the remaining tokens in s, not including the current token if
// there is one.  Any tokens already consumed are still returned, even if there
// is an error.
func (s *Scanner) Split() []string {
	var tokens []string
	for s.Next() {
		tokens = append(tokens, s.Text())
	}
	return tokens
}

// scanPool buffers reusable scanners so that Split does not need to allocate
// new buffers in the common case.
var scanPool = &sync.Pool{
	New: func() any { return NewScanner(nil) },
}

// Split partitions s into tokens divided on space, tab, and newline characters
// using a *Scanner.  Leading and trailing whitespace are ignored.
//
// The Boolean flag reports whether the final token is "valid", meaning there
// were no unclosed quotations in the string.
func Split(s string) ([]string, bool) {
	sc := scanPool.Get().(*Scanner)
	defer scanPool.Put(sc)

	sc.Reset(strings.NewReader(s))
	ss := sc.Split()
	return ss, sc.Complete()
}

func quotable(s string) (hasQ, hasOther bool) {
	const (
		quote = 1
		other = 2
		all   = quote + other
	)
	var v uint
	for i := 0; i < len(s) && v < all; i++ {
		if s[i] == '\'' {
			v |= quote
		} else if strings.IndexByte(allQuote, s[i]) >= 0 {
			v |= other
		}
	}
	return v&quote != 0, v&other != 0
}

var bufPool = &sync.Pool{
	New: func() any { return new(bytes.Buffer) },
}

// Quote returns a copy of s in which shell metacharacters are quoted to
// protect them from evaluation.
func Quote(s string) string {
	if s == "" {
		return "''"
	}
	if hasQ, hasOther := quotable(s); !hasQ && !hasOther {
		return s // no quotation needed
	}
	buf := bufPool.Get().(*bytes.Buffer)
	defer bufPool.Put(buf)
	buf.Reset()

	quote(s, buf)
	return buf.String()
}

// quote appends the quoted representation of s to buf.
// Any existing contents in buf are not modified.
func quote(s string, buf *bytes.Buffer) {
	if s == "" {
		buf.WriteString("''")
		return
	}

	hasQ, hasOther := quotable(s)
	if !hasQ && !hasOther {
		buf.WriteString(s)
		return // fast path: nothing needs quotation
	}

	buf.Grow(len(s) + 2) // +2 for expected quotes; it may be more
	inq := false
	for i := range len(s) {
		ch := s[i]
		if ch == '\'' {
			if inq {
				buf.WriteByte('\'')
				inq = false
			}
			buf.WriteByte('\\')
		} else if !inq && hasOther {
			buf.WriteByte('\'')
			inq = true
		}
		buf.WriteByte(ch)
	}
	if inq {
		buf.WriteByte('\'')
	}
}

// Join quotes each element of ss with Quote and concatenates the resulting
// strings separated by spaces.
func Join(ss []string) string {
	if len(ss) == 0 {
		return ""
	}
	buf := bufPool.Get().(*bytes.Buffer)
	defer bufPool.Put(buf)
	buf.Reset()

	quote(ss[0], buf)
	for _, s := range ss[1:] {
		buf.WriteByte(' ')
		quote(s, buf)
	}
	return buf.String()
}
