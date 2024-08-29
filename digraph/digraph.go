package main

import (
	"iter"

	"github.com/creachadair/mds/mapset"
)

type Node[T any] struct {
	self *T
	ins  mapset.Set[*Node[T]]
	outs mapset.Set[*Node[T]]
}

type NodeEmbedder[T any] interface {
	getNodeInternal() *Node[T]
}

type NodeEmbedderPtr[T any] interface {
	NodeEmbedder[T]
	*T
}

func (n *Node[T]) getNodeInternal() *Node[T] { return n }

func (n *Node[T]) EdgesOut() iter.Seq[*T] {
	return func(yield func(val *T) bool) {
		for next := range n.outs {
			if !yield(next.self) {
				return
			}
		}
	}
}

func (n *Node[T]) EdgesIn() iter.Seq[*T] {
	return func(yield func(val *T) bool) {
		for prev := range n.ins {
			if !yield(prev.self) {
				return
			}
		}
	}
}

func (n *Node[T]) AddEdgeFrom(v NodeEmbedder[T]) {
	vnode := v.getNodeInternal()
	n.ins.Add(vnode)
	vnode.outs.Add(n)
}

func (n *Node[T]) RemoveEdgeFrom(v NodeEmbedder[T]) {
	vnode := v.getNodeInternal()
	n.ins.Remove(vnode)
	vnode.outs.Remove(n)
}

func (n *Node[T]) AddEdgeTo(v NodeEmbedder[T]) {
	vnode := v.getNodeInternal()
	n.outs.Add(vnode)
	vnode.ins.Add(n)
}

func (n *Node[T]) RemoveEdgeTo(v NodeEmbedder[T]) {
	vnode := v.getNodeInternal()
	n.outs.Remove(vnode)
	vnode.ins.Remove(n)
}

func (n *Node[T]) IsSuccessorOf(pred NodeEmbedder[T]) bool {
	return n.ins.Has(pred.getNodeInternal())
}

func (n *Node[T]) IsPredecessorOf(succ NodeEmbedder[T]) bool {
	return n.outs.Has(succ.getNodeInternal())
}

func New[T any, U NodeEmbedderPtr[T]](v U) U {
	v.getNodeInternal().self = v
	return v
}

type BuildStep struct {
	Node[BuildStep]

	Name     string
	Commands []string
}

func main() {
	g1 := New(&BuildStep{
		Name: "run something",
	})
	g2 := New(&BuildStep{
		Name: "run a second thing",
	})
	g2.AddEdgeFrom(g1)
}
