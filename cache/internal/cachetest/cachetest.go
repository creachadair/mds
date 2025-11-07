// Package cachetest implements a test harness for cache implementations.
package cachetest

import (
	"errors"
	"fmt"
	"strconv"
	"strings"
	"testing"

	"github.com/creachadair/mds/cache"
)

// An Op represents the operation code of an instruction.
type Op string

// The operation codes supported by a cache.
const (
	OpHas    Op = "has"
	OpGet    Op = "get"
	OpPut    Op = "put"
	OpRemove Op = "remove"
	OpClear  Op = "clear"
	OpLen    Op = "len"
	OpSize   Op = "size"
	OpPop    Op = "pop"
)

// An insn is a single instruction in a cache test program.  Each instruction
// describes an operation to apply to the cache, the arguments to that
// operation, and the expected results.
type insn struct {
	Op    Op     // the operation to apply
	Key   string // for has, get, put
	Value string // for put, remove

	resV  string // for get, the expected value
	resK  string // for pop, the expected key
	resOK bool   // for has, get, put, remove
	resZ  int64  // for len, size
	text  string // for pretty-printing the instruction
}

func (in insn) String() string { return in.text }

// Run compiles and evaluates the given test program on c.  If the compilation
// step fails, no operations are applied to c, and the test fails immediately.
// Otherwise, the whole program is run and errors are logged as appropriate.
//
// The general format of a test program instruction is:
//
//	opcode [args...] [= results ...]
//
// Arguments and results are separated by spaces.  The number and types of the
// arguments correspond to the operations on a cache, for example "get" takes a
// single key and returns a value and a bool, while "len" takes no arguments
// and returns an int.  ParseInsn will report an error if the arguments and
// results do not match the opcode.
//
// As a special case, the empty string can be written as ”.
//
// Examples
//
//	len = 0
//	get foo = bar true
//	get quux = '' false
//	has nonesuch = false
//	clear
func Run(t *testing.T, c *cache.Cache[string, string], prgm ...string) {
	t.Helper()

	var insn []insn
	for i, p := range prgm {
		ins, err := parseInsn(p)
		if err != nil {
			t.Fatalf("Line %d: parse %q: %v", i+1, p, err)
		}
		insn = append(insn, ins)
	}

	for i, ins := range insn {
		if err := ins.eval(c); err != nil {
			t.Errorf("Line %d: %s: %v", i+1, ins, err)
		}
	}
}

func (in insn) eval(c *cache.Cache[string, string]) error {
	switch in.Op {
	case OpHas:
		got := c.Has(in.Key)
		if got != in.resOK {
			return fmt.Errorf("c.Has(%q): got %v, want %v", in.Key, got, in.resOK)
		}
	case OpGet:
		got, ok := c.Get(in.Key)
		if got != in.resV || ok != in.resOK {
			return fmt.Errorf("c.Get(%q): got (%q, %v), want (%q, %v)", in.Key, got, ok, in.resV, in.resOK)
		}
	case OpPut:
		if got, want := c.Put(in.Key, in.Value), in.resOK; got != want {
			return fmt.Errorf("c.Put(%q, %q): got %v, want %v", in.Key, in.Value, got, want)
		}
	case OpRemove:
		if got, want := c.Remove(in.Key), in.resOK; got != want {
			return fmt.Errorf("c.Remove(%q): got %v, want %v", in.Key, got, want)
		}
	case OpClear:
		c.Clear() // cannot fail
		return nil
	case OpLen:
		if got, want := c.Len(), int(in.resZ); got != want {
			return fmt.Errorf("c.Len(): got %d, want %d", got, want)
		}
	case OpSize:
		if got, want := c.Size(), in.resZ; got != want {
			return fmt.Errorf("c.Size(): got %d, want %d", got, want)
		}
	case OpPop:
		if gotK, gotV, ok := c.Pop(); gotK != in.resK || gotV != in.resV || ok != in.resOK {
			return fmt.Errorf("c.Pop(): got (%q, %q, %v); want (%q, %q, %v)",
				gotK, gotV, ok, in.resK, in.resV, in.resOK)
		}
	default:
		panic(fmt.Sprintf("eval: unknown opcode %q", in.Op))
	}
	return nil
}

// parseInsn parses an instruction from a string format.
func parseInsn(s string) (insn, error) {
	op, tail, _ := strings.Cut(s, "=")
	args := strings.Fields(op)
	resp := strings.Fields(tail)
	if len(args) == 0 {
		return insn{}, errors.New("missing opcode")
	}

	out := insn{
		Op:   Op(args[0]),
		text: strings.Join(args, " "), // for the String method
	}
	if len(resp) != 0 {
		out.text += " = " + strings.Join(resp, " ")
	}

	// Check argument counts.
	var narg, nres int
	switch out.Op {
	case "":
		return insn{}, errors.New("missing opcode")
	case OpGet:
		narg, nres = 1, 2
	case OpHas, OpRemove:
		narg, nres = 1, 1
	case OpPut:
		narg, nres = 2, 1
	case OpClear:
	case OpLen, OpSize:
		narg, nres = 0, 1
	case OpPop:
		narg, nres = 0, 3
	default:
		return insn{}, fmt.Errorf("unknown opcode %q", args[0])
	}
	if len(args) != narg+1 {
		return insn{}, fmt.Errorf("op %q has %d args, want %d", args[0], len(args)-1, narg)
	}
	if len(resp) != nres {
		return insn{}, fmt.Errorf("op %q has %d results, want %d", args[0], len(resp), nres)
	}

	// Check argument and result types.
	switch out.Op {
	case OpHas, OpGet, OpPut, OpRemove:
		out.Key = args[1]
		b, err := strconv.ParseBool(resp[len(resp)-1])
		if err != nil {
			return insn{}, fmt.Errorf("op %q result: %w", out.Op, err)
		}
		out.resOK = b
	case OpLen, OpSize:
		v, err := strconv.ParseInt(resp[0], 10, 64)
		if err != nil {
			return insn{}, fmt.Errorf("op %q result: %w", out.Op, err)
		}
		out.resZ = v
	case OpPop:
		v, err := strconv.ParseBool(resp[2])
		if err != nil {
			return insn{}, fmt.Errorf("op %q result: %w", out.Op, err)
		}
		out.resK = unquoteEmpty(resp[0])
		out.resV = unquoteEmpty(resp[1])
		out.resOK = v
	}
	if out.Op == OpGet {
		out.resV = unquoteEmpty(resp[0])
	}
	if out.Op == OpPut {
		out.Value = args[2]
	}
	return out, nil
}

func unquoteEmpty(s string) string {
	if s == "''" {
		return ""
	}
	return s
}
