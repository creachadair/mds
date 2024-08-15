# mds

[![GoDoc](https://img.shields.io/static/v1?label=godoc&message=reference&color=white)](https://pkg.go.dev/github.com/creachadair/mds)
[![CI](https://github.com/creachadair/mds/actions/workflows/go-presubmit.yml/badge.svg?event=push&branch=main)](https://github.com/creachadair/mds/actions/workflows/go-presubmit.yml)

This repository defines generic data structures in Go.

## Data Structures

- [heapq](./heapq) a heap-structured priority queue ([package docs](https://godoc.org/github.com/creachadair/mds/heapq))
- [mapset](./mapset) a basic map-based set implementation ([package docs](https://godoc.org/github.com/creachadair/mds/mapset))
- [mlink](./mlink) basic linked sequences (list, queue, ring) ([package docs](https://godoc.org/github.com/creachadair/mds/mlink))
- [omap](./omap) ordered key-value map ([package docs](https://godoc.org/github.com/creachadair/mds/omap))
- [oset](./oset) ordered set ([package docs](https://godoc.org/github.com/creachadair/mds/oset))
- [queue](./queue) an array-based FIFO queue ([package docs](https://godoc.org/github.com/creachadair/mds/queue))
- [stack](./stack) an array-based LIFO stack ([package docs](https://godoc.org/github.com/creachadair/mds/stack))
- [stree](./stree) self-balancing binary-search tree ([package docs](https://godoc.org/github.com/creachadair/mds/stree))

## Utilities

- [distinct](./distinct) a probabilistic distinct-elements counter (CVM) ([package docs](https://godoc.org/github.com/creachadair/mds/distinct))
- [slice](./slice) helpful functions for manipulating slices ([package docs](https://godoc.org/github.com/creachadair/mds/slice))
- [mbits](./mbits) helpful functions for manipulating bits and bytes ([package docs](https://godoc.org/github.com/creachadair/mds/mbits))
- [mdiff](./mdiff) supports creating textual diffs ([package docs](https://godoc.org/github.com/creachadair/mds/mdiff), [example](https://go.dev/play/p/DFL6f7WLcff))
- [mstr](./mstr) helpful functions for manipulating strings ([package docs](https://godoc.org/github.com/creachadair/mds/mstr))
- [mtest](./mtest) a support library for writing tests ([package docs](https://godoc.org/github.com/creachadair/mds/mtest))
- [shell](./shell) POSIX shell quoting and splitting ([package docs](https://godoc.org/github.com/creachadair/mds/shell))
- [value](./value) helpful functions for basic values and pointers ([package docs](https://godoc.org/github.com/creachadair/mds/value))
