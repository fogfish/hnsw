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

	h.Lock()
	h.heap = append(h.heap, node)
	addr := Pointer(len(h.heap) - 1)
	h.Unlock()

	// skip down through layers
	h.RLock()
	head := h.head
	hLevel := h.level
	h.RUnlock()

	for lvl := hLevel - 1; lvl > level; lvl-- {
		head = h.skip(lvl, head, v)
	}

	//
	for lvl := min(level, hLevel-1); lvl >= 0; lvl-- {
		M := h.config.mLayerN
		if lvl == 0 {
			M = h.config.mLayer0
		}

		w := h.SearchLayer(lvl, head, v, h.config.efConstruction)

		// TODO: Selector algorithm this one is Neighbor simple
		for w.Len() > M {
			w.Deq()
		}
		// if w.Len() > M {
		// 	w = h.SelectNeighboursHeuristic(lvl, v, w, M)
		// }

		// Add Bi-Edges
		//h.Lock()
		edges := make([]Pointer, w.Len())
		// node.Connections[lvl] = make([]Pointer, w.Len())
		for i := w.Len() - 1; i >= 0; i-- {
			candidate := w.Deq()
			edges[i] = candidate.Addr
			//node.Connections[lvl][i] = candidate.Addr

			n := h.heap[candidate.Addr]
			c := n.Connections[lvl]
			h.Lock()
			n.Connections[lvl] = append(c, addr)
			h.Unlock()
		}
		h.Lock()
		node.Connections[lvl] = edges
		h.Unlock()

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
				// if edges.Len() > M {
				// 	edges = h.SelectNeighboursHeuristic(lvl, h.heap[e].Vector, edges, M)
				// }

				conns := make([]Pointer, edges.Len())
				for i := edges.Len() - 1; i >= 0; i-- {
					conns[i] = edges.Deq().Addr
				}

				h.Lock()
				h.heap[e].Connections[lvl] = conns
				h.Unlock()
			}
		}

	}

	h.Lock()
	if len(node.Connections) > h.level {
		h.level = len(node.Connections)
		h.head = addr
	}
	h.Unlock()
}

/*
func (h *HNSW[Vector]) SelectNeighboursHeuristic(level int, q Vector, c pq.Queue[Vertex], m int) pq.Queue[Vertex] {
	var inW bitset.BitSet

	w := pq.New(ordForwardVertex(""))

	// extend candidates by their neighbors
	for c.Len() > 0 {
		e := c.Deq()
		for _, eadj := range h.heap[e.Addr].Connections[level] {
			if !inW.Test(uint(eadj)) {
				inW.Set(uint(eadj))
				w.Enq(Vertex{
					Distance: h.surface.Distance(q, h.heap[eadj].Vector),
					Addr:     eadj,
				})
			}
		}
	}

	r := pq.New(ordForwardVertex(""))
	d := pq.New(ordForwardVertex(""))

	for w.Len() > 0 {
		if r.Len() > m {
			break
		}

		e := w.Deq()
		if r.Len() == 0 || r.Head().Distance > e.Distance {
			r.Enq(e)
		} else {
			d.Enq(e)
		}
	}

	for d.Len() > 0 {
		if r.Len() > m {
			break
		}
		r.Enq(d.Deq())
	}

	return r
}
*/
