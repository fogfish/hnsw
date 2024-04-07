//
// Copyright (C) 2024 Dmitry Kolesnikov
//
// This file may be modified and distributed under the terms
// of the MIT license.  See the LICENSE file for details.
// https://github.com/fogfish/hnsw
//

package hnsw

// Pointer to Node
type Pointer = uint32

// Graph Node
type Node[Vector any] struct {
	Vector      Vector
	Connections [][]Pointer
}

// Graph Nodes
type Nodes[Vector any] struct {
	Rank int
	Head Pointer
	Heap []Node[Vector]
}

// Vertex to graph node
type Vertex struct {
	Distance float32
	Addr     Pointer
}

// Forward Vertex Ordering
type ordForwardVertex string

func (ordForwardVertex) Compare(a, b Vertex) int {
	if a.Distance < b.Distance {
		return -1
	}

	if a.Distance > b.Distance {
		return 1
	}

	return 0
}

// Reverse Vertex Ordering
type ordReverseVertex string

func (ordReverseVertex) Compare(a, b Vertex) int {
	if a.Distance > b.Distance {
		return -1
	}

	if a.Distance < b.Distance {
		return 1
	}

	return 0
}
