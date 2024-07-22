//
// Copyright (C) 2024 Dmitry Kolesnikov
//
// This file may be modified and distributed under the terms
// of the MIT license.  See the LICENSE file for details.
// https://github.com/fogfish/hnsw
//

package hnsw

import (
	"fmt"
	"io"

	"github.com/bits-and-blooms/bitset"
)

// Node visitor function.
type FMap[Vector any] func(rank int, vector Vector, edges []Vector) error

// Breadth-first search iterator over all nodes linked at the level.
//
// This method provides an iterator that traverses all nodes linked at a specific
// level of the graph. By performing a full scan, the `ForAll` method ensures
// comprehensive exploration of the graph's nodes, making it useful for
// applications that require a complete overview of the graph structure at a given level.
func (h *HNSW[Vector]) ForAll(level int, fmap FMap[Vector]) error {
	var visited bitset.BitSet

	return h.forNode(level, h.head, &visited, fmap)
}

func (h *HNSW[Vector]) forNode(level int, addr Pointer, visited *bitset.BitSet, fmap FMap[Vector]) error {
	if visited.Test(uint(addr)) {
		return nil
	}
	visited.Set(uint(addr))

	node := h.heap[addr]

	var edges []Vector
	if len(node.Connections) > level {
		edges = make([]Vector, len(node.Connections[level]))
		for i, addr := range node.Connections[level] {
			edges[i] = h.heap[addr].Vector
		}
	}

	if err := fmap(len(node.Connections), node.Vector, edges); err != nil {
		return err
	}

	if len(node.Connections) > level {
		for _, addr := range node.Connections[level] {
			if err := h.forNode(level, addr, visited, fmap); err != nil {
				return err
			}
		}
	}

	return nil
}

// Heap iterator over data structure
func (h *HNSW[Vector]) FMap(level int, fmap FMap[Vector]) error {
	for _, node := range h.heap {
		if len(node.Connections) > level {
			edges := make([]Vector, len(node.Connections[level]))
			for i, addr := range node.Connections[level] {
				edges[i] = h.heap[addr].Vector
			}

			if err := fmap(len(node.Connections), node.Vector, edges); err != nil {
				return err
			}
		}
	}

	return nil
}

// Dump index as text
func (h *HNSW[Vector]) Dump(w io.Writer, f func(Vector) string) {
	for lvl := h.level - 1; lvl >= 0; lvl-- {
		w.Write([]byte(fmt.Sprintf("\n\n==> %v\n", lvl)))

		h.FMap(lvl, func(level int, vector Vector, vertex []Vector) error {

			w.Write([]byte(fmt.Sprintf("%s | ", f(vector))))
			for _, e := range vertex {
				w.Write([]byte(fmt.Sprintf("%s ", f(e))))
			}
			w.Write([]byte("\n"))

			return nil
		})
	}
}
