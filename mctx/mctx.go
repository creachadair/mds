// Copyright (C) Michael J. Fromberger. All Rights Reserved.

// Package mctx provides support for attaching data to [context.Context] values.
//
// # Overview
//
// A [context.Context] allows the caller to attach key-value pairs that can be
// recovered by code receiving the context. This package provides a lightweight
// typed wrapper to the standard key/value plumbing.
//
// To attach values to a context, create a [mctx.Key] for the given value type:
//
//	var infoKey mctx.Key[Info]
//
// To attach a value to a context, use [Key.Attach]:
//
//	infoCtx := infoKey.Attach(ctx, Info{ ... })
//
// To recover a value from a context, use [Key.Lookup]:
//
//	v := infoKey.Lookup(infoCtx)
//
// This reports a [value.Maybe] that contains the value for the key, if
// one is present.
//
// # Examples
//
// Check for a value:
//
//	ok := infoKey.Lookup(infoCtx).Present()
//
// Supply a default:
//
//	info := infoKey.Lookup(infoCtx).Or(defaultInfo).Get()
//
// Check for presence and return the value:
//
//	info, ok := infoKey.Lookup(infoCtx).GetOK()
package mctx

import (
	"context"
	"fmt"

	"github.com/creachadair/mds/value"
)

// A Key is a context key used to idetify values attached to a [context.Context].
//
// A zero Key is valid. Multiple keys for a given type may be distinguished by
// constructing non-zero values of the Key type.
type Key[T any] string

// Attach returns a context derived from ctx with value attached at the given key.
func (k Key[T]) Attach(ctx context.Context, value T) context.Context {
	return context.WithValue(ctx, k, value)
}

// Lookup reports whether ctx has a value for the given key, and if so returns
// that value. If the key is not present, it reports a zero value.
func (k Key[T]) Lookup(ctx context.Context) value.Maybe[T] {
	if v := ctx.Value(k); v != nil {
		return value.Just(v.(T))
	}
	return value.Absent[T]()
}

// String returns a human-readable representation of k.
func (k Key[T]) String() string {
	var zero T
	return fmt.Sprintf("Key[%T](%q)", zero, string(k))
}
