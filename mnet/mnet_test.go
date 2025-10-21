package mnet_test

import (
	"context"
	"errors"
	"io"
	"net"
	"testing"
	"testing/synctest"
	"time"

	"github.com/creachadair/mds/mnet"
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
		n := mnet.New("empty")
		checkNetError(t, "Close empty", n.Close(), nil, false)
	})

	t.Run("ListenClosed", func(t *testing.T) {
		n := mnet.New("closed")
		n.Close()
		lst, err := n.Listen("tcp", "whatever")
		if !checkNetError(t, "Listen on closed", err, net.ErrClosed, false) {
			t.Logf("Got result: %+v", lst)
		}
	})

	t.Run("DialClosed", func(t *testing.T) {
		n := mnet.New("closed")
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
			n := mnet.New("test")

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
			n := mnet.New("test")
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
		n := mnet.New("empty")
		conn, err := n.Dial("tcp", "nonesuch")
		if !checkNetError(t, "Dial", err, mnet.ErrConnRefused, false) {
			t.Logf("Got result: %+v", conn)
		}
	})

	t.Run("DialOK", func(t *testing.T) {
		synctest.Test(t, func(t *testing.T) {
			n := mnet.New("test")
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
			n := mnet.New("test")
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
			n := mnet.New("test")
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
			n := mnet.New("test")
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
}
