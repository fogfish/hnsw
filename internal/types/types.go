//
// Copyright (C) 2024 Dmitry Kolesnikov
//
// This file may be modified and distributed under the terms
// of the MIT license.  See the LICENSE file for details.
// https://github.com/fogfish/hnsw
//

package types

// Vertex to graph node
type Vertex struct {
	Distance float32
	Addr     uint32
}

// Forward Vertex Ordering
type ordForwardVertex int

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
type ordReverseVertex int

func (ordReverseVertex) Compare(a, b Vertex) int {
	if a.Distance > b.Distance {
		return -1
	}

	if a.Distance < b.Distance {
		return 1
	}

	return 0
}

const (
	OrdForwardVertex = ordForwardVertex(0)
	OrdReverseVertex = ordReverseVertex(1)
)
