package mnet_test

import (
	"context"
	"errors"
	"io"
	"net"
	"strconv"
	"testing"
	"testing/synctest"
	"time"

	"github.com/creachadair/mds/mnet"
	"github.com/creachadair/mds/mtest"
)

func checkAddr(t *testing.T, label string, addr net.Addr, wantNet, wantAddr string) {
	t.Helper()
	n, s := addr.Network(), addr.String()
	if n != wantNet || s != wantAddr {
		t.Errorf("%s: got (%q, %q), want (%q, %q)", label, n, s, wantNet, wantAddr)
	}
}

func checkNetError(t *testing.T, label string, got, want error, isTimeout bool) bool {
	t.Helper()
	if got != nil {
		var ne net.Error
		if !errors.As(got, &ne) {
			t.Errorf("%s error type %T is not net.Error", label, got)
		} else if ne.Timeout() != isTimeout {
			t.Errorf("%s net error: timeout=%v, want %v", label, ne.Timeout(), isTimeout)
		}
	}
	if want != nil && !errors.Is(got, want) {
		t.Errorf("%s error: got %v, want %v", label, got, want)
	}
	return !t.Failed()
}

func TestNetwork(t *testing.T) {
	t.Run("CloseEmpty", func(t *testing.T) {
		n := mnet.New(t.Name())
		checkNetError(t, "Close empty", n.Close(), nil, false)
	})

	t.Run("ListenClosed", func(t *testing.T) {
		n := mnet.New(t.Name())
		n.Close()
		lst, err := n.Listen("tcp", "whatever")
		if !checkNetError(t, "Listen on closed", err, net.ErrClosed, false) {
			t.Logf("Got result: %+v", lst)
		}
	})

	t.Run("ListenRandom", func(t *testing.T) {
		synctest.Test(t, func(t *testing.T) {
			n := mnet.New(t.Name())
			defer n.Close()

			lst1, err := n.Listen("tcp", "foo:0")
			if err != nil {
				t.Fatalf("Listen 1: unexpected error: %v", err)
			}
			p1 := checkHostPort(t, lst1.Addr().String(), "foo")
			t.Logf("Listener 1 assigned port %d", p1)

			lst2, err := n.Listen("tcp", "foo:0")
			if err != nil {
				t.Fatalf("Listen 2: unexpected error: %v", err)
			}
			p2 := checkHostPort(t, lst2.Addr().String(), "foo")
			t.Logf("Listener 2 assigned port %d", p2)

			if p1 == p2 {
				t.Errorf("Duplicate port assignment %d for both listeners", p1)
			}

			// Verify that we can actually dial the assigned address.
			go func() {
				cli, err := lst1.Accept()
				if err != nil {
					t.Errorf("Accept: unexpected error: %v", err)
				}
				t.Logf("Connection to %q accepted from %q OK", cli.LocalAddr(), cli.RemoteAddr())
				cli.Close()
			}()

			srv, err := n.Dial("tcp", lst1.Addr().String())
			if err != nil {
				t.Fatalf("Dial %q: unexpected error: %v", lst1.Addr(), err)
			}
			t.Logf("Connected to %q OK", srv.RemoteAddr())
			srv.Close()

			synctest.Wait() // allow logging to finish
		})
	})

	t.Run("MustListenRandom", func(t *testing.T) {
		n := mnet.New(t.Name())
		lst := n.MustListen("tcp4", "127.0.0.1:0")
		checkHostPort(t, lst.Addr().String(), "127.0.0.1")
	})

	t.Run("ListenRandomUnique", func(t *testing.T) {
		n := mnet.New(t.Name())
		n.MustListen("tcp", "foo:1024")
		n.MustListen("tcp", "bar:1024")
		n.MustListen("tcp", "foo:1026")
		lst1 := n.MustListen("tcp", "foo:0")
		checkHostPort(t, lst1.Addr().String(), "foo")
		t.Logf("Address 1: %q", lst1.Addr())
		lst2 := n.MustListen("tcp", "bar:0")
		checkHostPort(t, lst2.Addr().String(), "bar")
		t.Logf("Address 2: %q", lst2.Addr())

		if bad, err := n.Listen("tcp", "foo:1026"); err == nil {
			t.Errorf("Listen: got %v, want error", bad)
		}
	})

	t.Run("DialClosed", func(t *testing.T) {
		n := mnet.New(t.Name())
		defer n.Close()

		lst, err := n.Listen("tcp", "whatever")
		if !checkNetError(t, "Listen", err, nil, false) {
			t.Fatal("Listen failed")
		}

		n.Close()

		// After closing the network, dials to it should fail even if there was
		// previously a valid listener.
		if conn, err := n.Dial("tcp", "whatever"); !checkNetError(t, "Dial on closed", err, net.ErrClosed, false) {
			t.Logf("Got result: %+v", conn)
		}

		// Also after closing the network, its listeners should act closed.
		if acc, err := lst.Accept(); !checkNetError(t, "Listen on closed", err, net.ErrClosed, false) {
			t.Logf("Got result: %+v", acc)
		}
	})

	t.Run("CloseDialInFlight", func(t *testing.T) {
		synctest.Test(t, func(t *testing.T) {
			n := mnet.New(t.Name())

			if _, err := n.Listen("tcp", "whatever"); !checkNetError(t, "Listen", err, nil, false) {
				t.Fatal("Listen failed")
			}

			// Close the network while a connection is pending, to verify that
			// Dial gets unblocked.
			time.AfterFunc(3*time.Second, func() { n.Close() })

			conn, err := n.Dial("tcp", "whatever")
			if !checkNetError(t, "Dial", err, mnet.ErrConnRefused, false) {
				t.Logf("Got result: %+v", conn)
			}
		})
	})

	t.Run("CloseAcceptInFlight", func(t *testing.T) {
		synctest.Test(t, func(t *testing.T) {
			n := mnet.New(t.Name())
			defer n.Close()

			lst, err := n.Listen("tcp", "whatever")
			if !checkNetError(t, "Listen", err, nil, false) {
				t.Fatal("Listen failed")
			}

			// Close the network while a connection is pending, to verify that
			// Accept gets unblocked.
			time.AfterFunc(3*time.Second, func() { n.Close() })

			acc, err := lst.Accept()
			if !checkNetError(t, "Accept", err, net.ErrClosed, false) {
				t.Logf("Got result: %+v", acc)
			}
		})
	})

	t.Run("DialMissing", func(t *testing.T) {
		n := mnet.New(t.Name())
		defer n.Close()

		conn, err := n.Dial("tcp", "nonesuch")
		if !checkNetError(t, "Dial", err, mnet.ErrConnRefused, false) {
			t.Logf("Got result: %+v", conn)
		}
	})

	t.Run("DialListener", func(t *testing.T) {
		synctest.Test(t, func(t *testing.T) {
			n := mnet.New(t.Name())
			defer n.Close()

			lst := n.MustListen("tcp", "whatever")
			go func() {
				acc, err := lst.Accept()
				if err != nil {
					t.Errorf("Accept failed: %v", err)
				} else {
					t.Logf("Accept OK: %T (%v)", acc, acc.RemoteAddr())
					acc.Close()
				}
			}()

			conn, err := lst.Dial()
			if !checkNetError(t, "Dial", err, nil, false) {
				t.Fatal("Dial failed")
			}
			conn.Close()

			synctest.Wait() // allow logging to finish
		})
	})

	t.Run("MustListenPanic", func(t *testing.T) {
		n := mnet.New(t.Name())
		defer n.Close()

		n.MustListen("test", "xyzzy") // succeed

		v := mtest.MustPanicf(t, func() {
			n.MustListen("test", "xyzzy")
		}, "duplicate listen should panic")
		t.Logf("Got expected panic: %v", v)
	})

	t.Run("DialOK", func(t *testing.T) {
		synctest.Test(t, func(t *testing.T) {
			n := mnet.New(t.Name())
			defer n.Close()

			const testNet, testAddr = "tcp", "example.net:12345"

			lst, err := n.Listen(testNet, testAddr)
			if !checkNetError(t, "Listen", err, nil, false) {
				t.Fatal("Listen failed")
			}
			defer lst.Close()

			checkAddr(t, "Listener", lst.Addr(), testNet, testAddr)

			go func() {
				acc, err := lst.Accept()
				if err != nil {
					t.Errorf("Accept failed: %v", err)
					return
				}
				checkAddr(t, "Accepted local", acc.LocalAddr(), testNet, testAddr)
				t.Logf("Accept OK: %T (%v)", acc, acc.RemoteAddr())
				acc.Close()
			}()

			// Dial the address advertised by the listener.
			conn, err := n.DialContext(t.Context(), lst.Addr().Network(), lst.Addr().String())
			if !checkNetError(t, "Dial", err, nil, false) {
				t.Fatal("Dial failed")
			}
			defer conn.Close()

			checkAddr(t, "Conn remote", conn.RemoteAddr(), testNet, testAddr)
			synctest.Wait() // allow logging to finish
		})
	})

	t.Run("DialTimeout", func(t *testing.T) {
		synctest.Test(t, func(t *testing.T) {
			n := mnet.New(t.Name())
			defer n.Close()

			lst, err := n.Listen("tcp", "example")
			if !checkNetError(t, "Listen", err, nil, false) {
				t.Fatal("Listen failed")
			}
			defer lst.Close()

			ctx, cancel := context.WithTimeout(t.Context(), time.Minute)
			defer cancel()

			conn, err := n.DialContext(ctx, "tcp", "example")
			if !checkNetError(t, "Dial", err, context.DeadlineExceeded, true) {
				t.Logf("Got result: %+v", conn)
			}
		})
	})

	t.Run("DialYN", func(t *testing.T) {
		synctest.Test(t, func(t *testing.T) {
			n := mnet.New(t.Name())
			defer n.Close()

			lst, err := n.Listen("tcp", "example")
			if !checkNetError(t, "Listen", err, nil, false) {
				t.Fatal("Listen failed")
			}
			go func() {
				for {
					acc, err := lst.Accept()
					if err != nil {
						t.Logf("NOTE: Accept: %v", err)
						return
					} else {
						acc.Close()
					}
				}
			}()

			// There is an active listener for this address.
			conn, err := n.Dial("tcp", "example")
			if err != nil {
				t.Fatalf("Dial failed: %v", err)
			}
			conn.Close()

			// After closing the listener, there is no longer an active listener for
			// the address, and a connection attempt for that address should fail.
			lst.Close()

			conn2, err := n.Dial("tcp", "example")
			if !checkNetError(t, "Dial 2", err, mnet.ErrConnRefused, false) {
				t.Logf("Got result: %+v", conn2)
			}

			synctest.Wait() // allow logging to finish
		})
	})

	t.Run("Connect", func(t *testing.T) {
		synctest.Test(t, func(t *testing.T) {
			n := mnet.New(t.Name())
			defer n.Close()

			lst, err := n.Listen("unix", "example")
			if !checkNetError(t, "Listen", err, nil, false) {
				t.Fatal("Listen failed")
			}

			// Server: Accept a connection, then read what it sends so we can
			// verify we got what we expected.
			var req string
			go func() {
				acc, err := lst.Accept()
				if !checkNetError(t, "Accept", err, nil, false) {
					t.Logf("Got result: %v", acc)
					return
				}
				t.Log("[srv] connection accepted")
				data, err := io.ReadAll(acc)
				if err != nil {
					t.Errorf("Read failed; %v", err)
					return
				}
				t.Logf("[srv] received: %q", data)
				req = string(data)
			}()

			// Client: Connect to the server and send some text, then close.
			conn, err := n.Dial("unix", "example")
			if !checkNetError(t, "Dial", err, nil, false) {
				t.Fatal("Dial failed")
			}
			t.Logf("[cli] connected: %v", conn.RemoteAddr())

			const testProbe = "squeaky wheel gets the kick"
			io.WriteString(conn, testProbe)
			conn.Close()

			// Verify that the data flowed through the pipes.
			synctest.Wait()

			if req != testProbe {
				t.Errorf("Request: got %q, want %q", req, testProbe)
			}
		})
	})

	t.Run("Dialer", func(t *testing.T) {
		synctest.Test(t, func(t *testing.T) {
			n := mnet.New(t.Name())
			defer n.Close()

			lst := n.MustListen("tcp", "server")
			d := n.Dialer("tcp", "client")

			go func() {
				cli, err := lst.Accept()
				if err != nil {
					t.Errorf("Accept: %v", err)
					return
				}
				defer cli.Close()

				// The remote address should match what was given to the dialer.
				addr := cli.RemoteAddr()
				if gn, ga := addr.Network(), addr.String(); gn != "tcp" || ga != "client" {
					t.Errorf("Accept: got addr (%q, %q); want (tcp, client)", gn, ga)
				}
			}()

			srv, err := d.DialContext(t.Context(), "tcp", "server")
			if err != nil {
				t.Fatalf("Dial: %v", err)
			}

			// The remote address should match what was dialed.
			if got := srv.RemoteAddr().String(); got != "server" {
				t.Errorf("Dial: got addr %q, want server", got)
			}
			srv.Close()
		})
	})
}

func checkHostPort(t *testing.T, addr, wantHost string) uint16 {
	hs, ps, err := net.SplitHostPort(addr)
	if err != nil {
		t.Fatalf("Invalid address format: %v", err)
	}
	if hs != wantHost {
		t.Errorf("Got host %q, want %q", hs, wantHost)
	}
	v, err := strconv.ParseInt(ps, 10, 16)
	if err != nil || v <= 0 {
		t.Errorf("Port %q: got (%v, %v), want (>0, nil)", ps, v, err)
	}
	return uint16(v)
}
