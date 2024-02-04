//
// Copyright (C) 2024 Dmitry Kolesnikov
//
// This file may be modified and distributed under the terms
// of the MIT license.  See the LICENSE file for details.
// https://github.com/fogfish/hnsw
//

package hnsw

import (
	"github.com/fogfish/hnsw/internal/pq"
	"github.com/willf/bitset"
)

// skip the graph to "nearest" node
func (h *HNSW[Vector]) skip(level int, addr Pointer, q Vector) Pointer {
	for {
		skip := h.skipToNearest(level, addr, q)
		if skip == addr {
			return skip
		}
		addr = skip
	}
}

// skip to "nearest" connection at the node.
// it return input address if no "movements" is possible
func (h *HNSW[Vector]) skipToNearest(level int, addr Pointer, q Vector) Pointer {
	dist := h.surface.Distance(h.heap[addr].Vector, q)

	for _, a := range h.heap[addr].Connections[level] {
		d := h.surface.Distance(h.heap[a].Vector, q)
		if d < dist {
			dist = d
			addr = a
		}
	}

	return addr
}

// Search "nearest" vectors on the layer
func (h *HNSW[Vector]) SearchLayer(level int, addr Pointer, q Vector, ef int) pq.Queue[Vertex] {
	var visited bitset.BitSet

	this := Vertex{
		Distance: h.surface.Distance(q, h.heap[addr].Vector),
		Addr:     addr,
	}

	candidates := pq.New(ordForwardVertex(""), this)
	setadidnac := pq.New(ordReverseVertex(""), this)

	for candidates.Len() > 0 {
		c := candidates.Deq()
		f := setadidnac.Head()

		if c.Distance > f.Distance {
			break
		}

		for _, e := range h.heap[c.Addr].Connections[level] {
			if !visited.Test(uint(e)) {
				visited.Set(uint(e))

				dist := h.surface.Distance(q, h.heap[e].Vector)
				item := Vertex{Distance: dist, Addr: e}

				if setadidnac.Len() < ef {
					if e != addr {
						setadidnac.Enq(item)
					}
					candidates.Enq(item)
				} else if dist < setadidnac.Head().Distance {
					setadidnac.Enq(item)
					setadidnac.Deq()
					candidates.Enq(item)
				}
			}
		}
	}

	return setadidnac
}

// Search K-nearest vectors from the graph
func (h *HNSW[Vector]) Search(q Vector, K int, efSearch int) []Vector {
	head := h.head
	for lvl := h.level - 1; lvl >= 0; lvl-- {
		head = h.skip(lvl, head, q)
	}

	w := h.SearchLayer(0, head, q, efSearch)
	for w.Len() > K {
		w.Deq()
	}

	v := make([]Vector, w.Len())
	for i := w.Len() - 1; i >= 0; i-- {
		x := w.Deq()
		v[i] = h.heap[x.Addr].Vector
	}

	return v
}
