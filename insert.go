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
	//
	// allocate new node
	//

	level := int(math.Floor(-math.Log(h.rand() * h.config.mL)))

	addr := Pointer(0)
	node := Node[Vector]{
		Vector:      v,
		Connections: make([][]Pointer, level+1),
	}

	//
	// skip down through layers
	//

	h.rwCore.RLock()
	head := h.head
	hLevel := h.level
	h.rwCore.RUnlock()

	for lvl := hLevel - 1; lvl > level; lvl-- {
		head = h.skip(lvl, head, v)
	}

	//
	// start building neighborhood
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

		// Add Edges from new node to existing one
		edges := make([]Pointer, w.Len())
		for i := w.Len() - 1; i >= 0; i-- {
			candidate := w.Deq()
			edges[i] = candidate.Addr
		}
		node.Connections[lvl] = edges
	}

	// if w.Len() > M {
	// 	w = h.SelectNeighboursHeuristic(lvl, v, w, M)
	// }

	//
	// Append new node
	//

	h.rwCore.Lock()
	addr = Pointer(len(h.heap))
	h.rwHeap[addr%heapRWSlots].Lock()
	h.heap = append(h.heap, node)
	h.rwHeap[addr%heapRWSlots].Unlock()
	h.rwCore.Unlock()

	for lvl, edges := range node.Connections {
		for i := 0; i < len(edges); i++ {
			h.addConnection(lvl, edges[i], addr)
		}
	}

	//
	// Shrink Connections
	//

	for lvl, edges := range node.Connections {
		M := h.config.mLayerN
		if lvl == 0 {
			M = h.config.mLayer0
		}

		for _, e := range edges {
			slot := e % heapRWSlots
			h.rwHeap[slot].RLock()
			enode := h.heap[e]
			eedges := enode.Connections[lvl]
			h.rwHeap[slot].RUnlock()

			if len(eedges) > M {
				edges := pq.New(ordReverseVertex(""))

				for _, n := range eedges {
					nnode := h.heap[n]

					dist := h.surface.Distance(enode.Vector, nnode.Vector)
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

				h.rwHeap[slot].Lock()
				h.heap[e].Connections[lvl] = conns
				h.rwHeap[slot].Unlock()
			}
		}
	}

	//
	// Update Heap
	//

	h.rwCore.Lock()
	if len(node.Connections) > h.level {
		h.level = len(node.Connections)
		h.head = addr
	}
	h.rwCore.Unlock()
}

func (h *HNSW[Vector]) addConnection(level int, src, dst Pointer) {
	slot := src % heapRWSlots

	h.rwHeap[slot].RLock()
	n := h.heap[src]
	c := n.Connections[level]
	h.rwHeap[slot].RUnlock()

	h.rwHeap[slot].Lock()
	n.Connections[level] = append(c, dst)
	h.rwHeap[slot].Unlock()
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
