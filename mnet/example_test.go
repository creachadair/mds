package mnet_test

import (
	"fmt"
	"io"
	"log"
	"os"

	"github.com/creachadair/mds/mnet"
)

func Example() {
	net := mnet.New("example")

	// Listen on an arbitrary network and address.
	// These values are not interpreted, but must match when dialing.
	lst, err := net.Listen("tcp", "localhost:12345")
	if err != nil {
		log.Fatalf("Listen: %v", err)
	}
	defer lst.Close()

	// Simulate a server by accepting a connection on lst and sending some text
	// to the caller.
	go func() {
		c, err := lst.Accept()
		if err != nil {
			log.Fatalf("Accept: %v", err)
		}
		fmt.Fprintln(c, "hello, world")
		c.Close()
	}()

	// Dial the listener to get a connection. The address must match one of the
	// listeners attached to the virtual network.
	c, err := net.Dial("tcp", "localhost:12345")
	if err != nil {
		log.Fatalf("Dial: %v", err)
	}
	io.Copy(os.Stdout, c)
	c.Close()

	// Output:
	// hello, world
}
