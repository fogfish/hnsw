//
// Copyright (C) 2024 Dmitry Kolesnikov
//
// This file may be modified and distributed under the terms
// of the MIT license.  See the LICENSE file for details.
// https://github.com/fogfish/hnsw
//

package hnsw

import (
	"errors"
	"io"
)

func (h *HNSW[Vector]) Write(
	nodes interface {
		Write(n Vector) error
	},
	edges interface {
		Write(v []Pointer) error
	},
) error {
	h.Lock()
	defer h.Unlock()

	for l := h.level - 1; l >= 0; l-- {
		hv := []Pointer{0, 0, 0, uint32(l)}
		if err := edges.Write(hv); err != nil {
			return err
		}

		for addr, node := range h.heap {
			if len(node.Connections) > l {
				iv := make([]Pointer, len(node.Connections[l])+1)
				iv[0] = Pointer(addr)
				copy(iv[1:], node.Connections[l])
				// for i, edge := range node.Connections[l] {
				// 	iv[i+1] = edge
				// }

				if err := edges.Write(iv); err != nil {
					return err
				}
			}
		}
	}

	for _, node := range h.heap {
		if err := nodes.Write(node.Vector); err != nil {
			return err
		}
	}

	return nil
}

func (h *HNSW[Vector]) Read(
	nodes interface {
		Read() (Vector, error)
	},
	edges interface {
		Read() ([]Pointer, error)
	},
) error {
	h.Lock()
	defer h.Unlock()

	if err := h.readNodes(nodes); err != nil {
		return err
	}

	if err := h.readEdges(edges); err != nil {
		return err
	}

	return nil
}

func (h *HNSW[Vector]) readNodes(
	nodes interface {
		Read() (Vector, error)
	},
) error {
	h.heap = []Node[Vector]{}

	for {
		nv, err := nodes.Read()
		switch {
		case err == nil:
			node := Node[Vector]{Vector: nv}
			h.heap = append(h.heap, node)
		case errors.Is(err, io.EOF):
			return nil
		default:
			return err
		}
	}
}

func (h *HNSW[Vector]) readEdges(
	edges interface {
		Read() ([]Pointer, error)
	},
) error {
	lvl := -1

	// fmt.Printf("%v\n", h.heap)

	for {
		iv, err := edges.Read()
		switch {
		case err == nil:
			if len(iv) == 4 && iv[0] == 0 && iv[1] == 0 && iv[2] == 0 {
				lvl = int(iv[3])
			} else {
				addr := iv[0]
				node := h.heap[addr]
				if node.Connections == nil {
					node.Connections = make([][]Pointer, lvl+1)
				}
				node.Connections[lvl] = iv[1:]
				// fmt.Printf("%v | %v\n", addr, node.Connections[lvl])
				h.heap[addr] = node
			}
		case errors.Is(err, io.EOF):
			return nil
		default:
			return err
		}
	}
}
