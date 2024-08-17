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
	"sync"

	"github.com/kshard/vector"
)

// Slots to coordinate concurrent I/O
const heapRWSlots = 1024

// Pointer to Node
type Pointer = uint32

// Node of Hierarchical Navigable Small World Graph
type Node[Vector any] struct {
	Vector      Vector
	Connections [][]Pointer
}

// Collection of serializable Hierarchical Navigable Small World Graph Nodes
type Nodes[Vector any] struct {
	Rank int
	Head Pointer
	Heap []Node[Vector]
}

// Hierarchical Navigable Small World Graph
type HNSW[Vector any] struct {
	rwCore sync.RWMutex
	rwHeap [heapRWSlots]sync.RWMutex

	config  Config
	surface vector.Surface[Vector]

	heap  []Node[Vector]
	head  Pointer
	level int
}

// Creates Hierarchical Navigable Small World Graph
func New[Vector any](
	surface vector.Surface[Vector],
	opts ...Option,
) *HNSW[Vector] {
	config := Config{}
	WithDefault()(&config)
	for _, opt := range opts {
		opt(&config)
	}

	hnsw := &HNSW[Vector]{
		config:  config,
		surface: surface,
	}

	hnsw.level = 0
	hnsw.heap = []Node[Vector]{}
	hnsw.head = 0

	return hnsw
}

// Hierarchical Navigable Small World Graph from exported nodes
func FromNodes[Vector any](
	surface vector.Surface[Vector],
	nodes Nodes[Vector],
	opts ...Option,
) *HNSW[Vector] {
	config := Config{}
	WithDefault()(&config)
	for _, opt := range opts {
		opt(&config)
	}

	hnsw := &HNSW[Vector]{
		config:  config,
		surface: surface,
	}

	hnsw.level = nodes.Rank
	hnsw.heap = nodes.Heap
	hnsw.head = nodes.Head

	return hnsw
}

func (h *HNSW[Vector]) String() string {
	return fmt.Sprintf("{ %d | Levels: %d  M: %d  M0: %d  mL: %f  efC: %d}",
		len(h.heap), h.level, h.config.mLayerN, h.config.mLayer0, h.config.mL, h.config.efConstruction)
}

// Return data structure nodes as serializable container.
func (h *HNSW[Vector]) Nodes() Nodes[Vector] {
	return Nodes[Vector]{
		Rank: h.level,
		Head: h.head,
		Heap: h.heap,
	}
}

// Return current head (entry point)
func (h *HNSW[Vector]) Head() Vector { return h.heap[h.head].Vector }

// Return current level
func (h *HNSW[Vector]) Level() int { return h.level }

// Return number of vectors in the data structure
func (h *HNSW[Vector]) Size() int { return len(h.heap) }

// Calculate distance between two vectors using defined surface distance function.
//
// The function is useful to fine-tune neighbors results, filtering non relevant values.
//
//	for _, vector := range neighbors {
//		if index.Distance(query, vector) < 0.2 {
//			// do something
//		}
//	}
func (h *HNSW[Vector]) Distance(a, b Vector) float32 {
	return h.surface.Distance(a, b)
}
