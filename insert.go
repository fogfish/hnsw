//
// Copyright (C) 2024 Dmitry Kolesnikov
//
// This file may be modified and distributed under the terms
// of the MIT license.  See the LICENSE file for details.
// https://github.com/fogfish/hnsw
//

package hnsw

import (
	"math"

	"github.com/fogfish/hnsw/internal/pq"
)

// generate random float from random source generator
func (h *HNSW[Vector]) rand() float64 {
again:
	f := float64(h.config.random.Int63()) / (1 << 63)
	if f == 1 {
		goto again // resample; this branch is taken O(never)
	}
	return f
}

// Insert new vector
func (h *HNSW[Vector]) Insert(v Vector) {
	level := int(math.Floor(-math.Log(h.rand() * h.config.mL)))

	node := Node[Vector]{
		Vector:      v,
		Connections: make([][]Pointer, level+1),
	}

	h.heap = append(h.heap, node)
	addr := Pointer(len(h.heap) - 1)

	// skip down through layers
	head := h.head
	for lvl := h.level - 1; lvl > level; lvl-- {
		head = h.skip(lvl, head, v)
	}

	//
	for lvl := min(level, h.level-1); lvl >= 0; lvl-- {
		M := h.config.mLayerN
		if lvl == 0 {
			M = h.config.mLayer0
		}

		w := h.SearchLayer(lvl, head, v, h.config.efConstruction)

		// TODO: Selector algorithm this one is Neighbor simple
		for w.Len() > M {
			w.Deq()
		}

		// Add Bi-Edges
		node.Connections[lvl] = make([]Pointer, w.Len())
		for i := w.Len() - 1; i >= 0; i-- {
			candidate := w.Deq()
			node.Connections[lvl][i] = candidate.Addr

			c := h.heap[candidate.Addr].Connections[lvl]
			h.heap[candidate.Addr].Connections[lvl] = append(c, addr)
		}

		// Shrink Connection
		for _, e := range node.Connections[lvl] {

			if len(h.heap[e].Connections[lvl]) > M {
				edges := pq.New(ordReverseVertex(""))

				for _, n := range h.heap[e].Connections[lvl] {
					dist := h.surface.Distance(h.heap[e].Vector, h.heap[n].Vector)
					item := Vertex{Distance: dist, Addr: n}
					edges.Enq(item)
				}

				for edges.Len() > M {
					edges.Deq()
				}

				conns := make([]Pointer, edges.Len())
				for i := edges.Len() - 1; i >= 0; i-- {
					conns[i] = edges.Deq().Addr
				}

				h.heap[e].Connections[lvl] = conns
			}
		}

	}

	if len(node.Connections) > h.level {
		h.level = len(node.Connections)
		h.head = addr
	}
}
