// Package mnet is an in-memory implementation of a subset of the [net] package.
//
// # Usage
//
// Create a new virtual [Network] using [mnet.New].
//
//	n := mnet.New("example")
//
// The name provided to the constructor is arbitrary, it is included in error
// messages to help with debugging. Each instance of [Network] is a separate
// connection namespace.
//
// To listen on the network, use [Network.Listen]:
//
//	lst, err := n.Listen("tcp", "example:12345")
//
// The network and address strings passed to Listen are not interpreted, but
// must match when dialing in order to reach the listener.
//
// To dial a connection, use [Network.Dial] or [Network.DialContext].
//
//	conn, err := n.DialContext(ctx, "tcp", "example:12345")
//
// If no listener exists for the specified network/address combination, it
// reports [ErrConnRefused]. If ctx ends before a connection was made, it
// reports a timeout. All errors reported by this package satisfy the
// [net.Error] interface.
//
// Once established, connections are the caller's responsibility and do not
// depend on the [Network] or [Listener] from which they were derived.  The
// underlying connection is provided by [net.Pipe] which is synchronous and
// nonblocking.
//
// When a [Network] is no longer needed, you may call its [Network.Close]
// method to close all its associated listeners and unblock any Dial or Accept
// calls pending. Once closed, a network is no longer usable, and future calls
// to [Network.Listen] and [Network.Dial] will report [net.ErrClosed].
package mnet

import (
	"context"
	"errors"
	"fmt"
	"io"
	"maps"
	"net"
	"path/filepath"
	"runtime"
	"slices"
	"strings"
	"sync"
)

// ErrConnRefused is a sentinel error reported when dialing an address not
// recognized by a [Network].
var ErrConnRefused = errors.New("connection refused")

// A Network is a virtual network that handles connections using synchronous
// in-memory pipes.
type Network struct {
	name string // immutable after initialization

	μ      sync.Mutex
	closed bool
	listen map[mnetAddr]Listener
}

// New constructs a new virtual network. The specified name is used only for
// diagnostics.
func New(name string) *Network {
	return &Network{name: name, listen: make(map[mnetAddr]Listener)}
}

// Name reports the name registered with construction of n.
func (n *Network) Name() string { return n.name }

// Close terminates all active listeners associated with n.
func (n *Network) Close() error {
	n.μ.Lock()
	all := slices.Collect(maps.Values(n.listen))
	n.closed = true
	n.μ.Unlock()

	for _, lst := range all {
		lst.Close()
	}
	return nil
}

// Listen returns a new [net.Listener] for the specified network and address.
// It reports an error if a listener already exists for the given address.
func (n *Network) Listen(network, addr string) (net.Listener, error) {
	n.μ.Lock()
	defer n.μ.Unlock()
	if n.closed {
		return nil, netErrorf(false, "[%s] listen: %w", n.name, net.ErrClosed)
	}

	key := mnetAddr{network: network, address: addr}
	if _, ok := n.listen[key]; ok {
		return nil, netErrorf(false, "[%s] listen %s %q: address already in use", n.name, network, addr)
	}
	stopCtx, cancel := context.WithCancel(context.Background())
	lst := Listener{
		netName: n.name,
		addr:    key,
		conns:   make(chan net.Conn),
		stopCtx: stopCtx,
		stop: func() {
			n.μ.Lock()
			defer n.μ.Unlock()
			if _, ok := n.listen[key]; ok {
				cancel()
				delete(n.listen, key)
			}
		},
	}
	n.listen[key] = lst
	return lst, nil
}

// MustListen returns a new [Listener] for the specified network and address.
// It panics if a listener already exists for the given address.
// This is intended for use in tests.
func (n *Network) MustListen(network, addr string) Listener {
	lst, err := n.Listen(network, addr)
	if err != nil {
		panic(err)
	}
	return lst.(Listener)
}

// Dial establishes a connection to the specified address on n.
// It reports [ErrConnRefused] if there is no active listener for the address.
// This is shorthand for [Network.DialContext] using a background context.
func (n *Network) Dial(network, addr string) (net.Conn, error) {
	lst, err := n.checkListener(network, addr)
	if err != nil {
		return nil, err // already wrapped
	}
	return lst.dialContext(context.Background())
}

// DialContext establishes a connection to the specified address on n.
// It reports [ErrConnRefused] if there is no active listener for the address.
// It reports a timeout if ctx ends before a connection can be established.
func (n *Network) DialContext(ctx context.Context, network, addr string) (net.Conn, error) {
	lst, err := n.checkListener(network, addr)
	if err != nil {
		return nil, err // already wrapped
	}
	return lst.dialContext(ctx)
}

func (n *Network) checkListener(network, addr string) (Listener, error) {
	n.μ.Lock()
	key := mnetAddr{network: network, address: addr}
	lst, ok := n.listen[key]
	isClosed := n.closed
	n.μ.Unlock()

	if isClosed {
		return Listener{}, netErrorf(false, "[%s] dial %s %q: %w", n.name, network, addr, net.ErrClosed)
	} else if !ok {
		return Listener{}, netErrorf(false, "[%s] dial %s %q: %w", n.name, network, addr, ErrConnRefused)
	}
	return lst, nil
}

// A Listener implements the [net.Listener] interface accepting connections
// from calls to [Network.Dial] and [Network.DialContext]. It is the concrete
// type of listeners returned by the [Network.Listen] method.
type Listener struct {
	netName string
	addr    mnetAddr
	conns   chan net.Conn

	stopCtx context.Context
	stop    func()
}

// Accept returns a connection from ln, or reports [net.ErrClosed] if the
// listener is closed before a connection is available.
// It implements part of [net.Listener].
func (ln Listener) Accept() (net.Conn, error) {
	select {
	case conn := <-ln.conns:
		return conn, nil
	case <-ln.stopCtx.Done():
		return nil, netErrorf(false, "[%s] accept: %w", ln.netName, net.ErrClosed)
	}
}

// Close implements part of [net.Listener]. It never reports an error.
func (ln Listener) Close() error { ln.stop(); return nil }

// Addr implements part of [net.Listener].
func (ln Listener) Addr() net.Addr { return ln.addr }

// Dial dials the address hosted by ln.
func (ln Listener) Dial() (net.Conn, error) { return ln.dialContext(context.Background()) }

// DialContext dials the address hosted by ln.
func (ln Listener) DialContext(ctx context.Context) (net.Conn, error) { return ln.dialContext(ctx) }

func (ln Listener) dialContext(ctx context.Context) (_ net.Conn, err error) {
	// Synthesize an "address" for the dialer based on its calling location.
	dialer := mnetAddr{network: ln.addr.network, address: "dial:unknown"}
	pc, fpath, line, _ := runtime.Caller(2)
	if f := runtime.FuncForPC(pc); f != nil {
		dialer.address = fmt.Sprintf("dial:%s:%s:%d", funcPackageName(f.Name()), filepath.Base(fpath), line)
	}

	lhs, rhs := net.Pipe()
	defer func() {
		if err != nil {
			lhs.Close()
			rhs.Close()
		}
	}()
	select {
	case ln.conns <- addrPipe{Conn: rhs, local: ln.addr, remote: dialer}:
		return addrPipe{Conn: lhs, local: dialer, remote: ln.addr}, nil
	case <-ln.stopCtx.Done():
		return nil, netErrorf(false, "[%s] dial %s %q: %w", ln.netName, ln.addr.network, ln.addr.address, ErrConnRefused)
	case <-ctx.Done():
		return nil, netErrorf(true, "[%s] dial %s %q: %w", ln.netName, ln.addr.network, ln.addr.address, ctx.Err())
	}
}

// mnetAddr implements the [net.Addr] interface.
type mnetAddr struct {
	network, address string
}

func (m mnetAddr) Network() string { return m.network }
func (m mnetAddr) String() string  { return m.address }

type addrPipe struct {
	net.Conn
	local, remote mnetAddr
}

// Read delegates to the underlying pipe, but treats [io.ErrClosedPipe] as
// equivalent to [io.EOF] since most callers do not know how to deal with that.
func (p addrPipe) Read(data []byte) (int, error) {
	n, err := p.Conn.Read(data)
	if errors.Is(err, io.ErrClosedPipe) {
		err = io.EOF
	}
	return n, err
}

func (p addrPipe) LocalAddr() net.Addr  { return p.local }
func (p addrPipe) RemoteAddr() net.Addr { return p.remote }

// netError satisfies the [net.Error] interface.
type netError struct {
	err       error
	isTimeout bool
}

func netErrorf(timeout bool, msg string, args ...any) error {
	return netError{
		err:       fmt.Errorf(msg, args...),
		isTimeout: timeout,
	}
}

func (e netError) Error() string { return e.err.Error() }
func (e netError) Timeout() bool { return e.isTimeout }
func (e netError) Unwrap() error { return e.err }
func (netError) Temporary() bool { return false }

func funcPackageName(funcName string) string {
	ls := max(strings.LastIndex(funcName, "/"), 0)
	for {
		i := strings.LastIndex(funcName, ".")
		if i <= ls {
			return funcName
		}
		funcName = funcName[:i]
	}
}
