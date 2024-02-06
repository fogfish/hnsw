//
// Copyright (C) 2024 Dmitry Kolesnikov
//
// This file may be modified and distributed under the terms
// of the MIT license.  See the LICENSE file for details.
// https://github.com/fogfish/hnsw
//

package hnsw

type FMap[Vector any] func(level int, vector Vector, vertex []Vector) error

func (h *HNSW[Vector]) FMap(level int, fmap FMap[Vector]) error {
	for _, node := range h.heap {
		if len(node.Connections) > level {

			vertex := make([]Vector, len(node.Connections[level]))
			for i, addr := range node.Connections[level] {
				vertex[i] = h.heap[addr].Vector
			}

			if err := fmap(len(node.Connections), node.Vector, vertex); err != nil {
				return err
			}

		}
	}

	return nil
}
