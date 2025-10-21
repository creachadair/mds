# mds

[![GoDoc](https://img.shields.io/static/v1?label=godoc&message=reference&color=white)](https://pkg.go.dev/github.com/creachadair/mds)
[![CI](https://github.com/creachadair/mds/actions/workflows/go-presubmit.yml/badge.svg?event=push&branch=main)](https://github.com/creachadair/mds/actions/workflows/go-presubmit.yml)

This repository defines generic data structures and utility types in Go.

The packages in this module are intended to be leaf libraries, and MUST NOT
depend on packages outside this module, excepting only packages from the Go
standard library, as well as selected golang.org/x/* packages on a case-by-case
basis.  Separate-package tests in this repository _may_ depend on other
packages, but such dependencies must be minimized.

Packages within the module _may_ depend on each other, where appropriate.

This module is currently versioned v0, and thus packages here are subject to
breaking changes. I will attempt to minimize such changes where practical, but
I may revise package APIs from time to time.  When making such changes, I will
increment the minor version number as a signal that more substantial changes
are included.

Although I wrote these packages mainly for my own use, they are intended to be
general-purpose, and you are welcome to use and depend on them.  If you do so,
I would be grateful for you to file
[issues](https://github.com/creachadair/mds/issues) for any problems you may
encounter.  While I cannot promise extensive support, I will do my best to
accommodate reasonable requests.

## Data Structures

Several of the data-types in this module share common behaviors:

- An `Add` method to add or update one or more elements in the container.
- A `Remove` method to remove one or more elements from the container.
- A `Clear` method that discards all the contents of the container.
- A `Peek` method that returns an order statistic of the container.
- An `Each` method that iterates the container in its natural order (usable as a [range function](https://go.dev/blog/range-functions)).
- An `IsEmpty` method that reports whether the container is empty.
- A `Len` method that reports the number of elements in the container.

### Packages

- [heapq](./heapq) a heap-structured priority queue ([package docs](https://godoc.org/github.com/creachadair/mds/heapq))
- [mapset](./mapset) a basic map-based set implementation ([package docs](https://godoc.org/github.com/creachadair/mds/mapset))
- [mlink](./mlink) basic linked sequences (list, queue) ([package docs](https://godoc.org/github.com/creachadair/mds/mlink))
- [omap](./omap) ordered key-value map ([package docs](https://godoc.org/github.com/creachadair/mds/omap))
- [queue](./queue) an array-based FIFO queue ([package docs](https://godoc.org/github.com/creachadair/mds/queue))
- [ring](./ring) a circular doubly-linked sequence ([package docs](https://godoc.org/github.com/creachadair/mds/ring))
- [stack](./stack) an array-based LIFO stack ([package docs](https://godoc.org/github.com/creachadair/mds/stack))
- [stree](./stree) self-balancing binary-search tree ([package docs](https://godoc.org/github.com/creachadair/mds/stree))

## Utilities

- [cache](./cache) an in-memory key/value cache ([package docs](https://godoc.org/github.com/creachadair/mds/cache))
- [compare](./compare) provides helpers for comparison of values ([package docs](https://godoc.org/github.com/creachadair/mds/compare))
- [distinct](./distinct) a probabilistic distinct-elements counter (CVM) ([package docs](https://godoc.org/github.com/creachadair/mds/distinct))
- [mbits](./mbits) helpful functions for manipulating bits and bytes ([package docs](https://godoc.org/github.com/creachadair/mds/mbits))
- [mdiff](./mdiff) supports creating textual diffs ([package docs](https://godoc.org/github.com/creachadair/mds/mdiff), [example](https://go.dev/play/p/xUYbbwnMkw3))
- [mnet](./mnet) provides an in-memory network implementation compatible with the `net` package ([package docs](https://godoc.org/github.com/creachadair/mds/mnet))
- [mstr](./mstr) helpful functions for manipulating strings ([package docs](https://godoc.org/github.com/creachadair/mds/mstr))
- [mtest](./mtest) a support library for writing tests ([package docs](https://godoc.org/github.com/creachadair/mds/mtest))
- [shell](./shell) POSIX shell quoting and splitting ([package docs](https://godoc.org/github.com/creachadair/mds/shell))
- [slice](./slice) helpful functions for manipulating slices ([package docs](https://godoc.org/github.com/creachadair/mds/slice))
- [value](./value) helpful functions for basic values and pointers ([package docs](https://godoc.org/github.com/creachadair/mds/value))
