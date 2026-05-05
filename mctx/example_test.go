// Copyright (C) Michael J. Fromberger. All Rights Reserved.

package mctx_test

import (
	"context"
	"fmt"

	"github.com/creachadair/mds/mctx"
)

type Info struct {
	Address string
	Port    int
	Purpose string
}

var infoKey = mctx.NewKey[Info]("metadata")

func Example() {
	ctx := infoKey.Attach(context.Background(), Info{
		Address: "example.com",
		Port:    8080,
		Purpose: "debug service",
	})

	handleRequest(ctx, "hello")
	fmt.Println()
	handleRequest(context.Background(), "goodbye")

	// Output:
	//
	// Message: hello
	// * Details
	//   Address: example.com port 8080
	//   Purpose: debug service
	//
	// Message: goodbye
	// - (no context)
}

func handleRequest(ctx context.Context, msg string) {
	fmt.Printf("Message: %s\n", msg)

	if v, ok := infoKey.Lookup(ctx).GetOK(); ok {
		fmt.Printf("* Details\n  Address: %s port %d\n", v.Address, v.Port)
		if v.Purpose != "" {
			fmt.Printf("  Purpose: %s\n", v.Purpose)
		}
	} else {
		fmt.Println("- (no context)")
	}
}
