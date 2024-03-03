//
// Copyright (C) 2024 Dmitry Kolesnikov
//
// This file may be modified and distributed under the terms
// of the MIT license.  See the LICENSE file for details.
// https://github.com/fogfish/hnsw
//

package hnsw

import (
	"encoding/json"
	"errors"
	"io"
)

type header struct {
	EfConstruction int     `json:"efConstruction"`
	MLayerN        int     `json:"mLayerN"`
	MLayer0        int     `json:"mLayer0"`
	ML             float64 `json:"mL"`
	Head           Pointer `json:"head"`
	Level          int     `json:"level"`
}

func (h *HNSW[Vector]) Write(
	w io.Writer,
	nodes interface {
		Write(n Vector) error
	},
	edges interface {
		Write(v []Pointer) error
	},
) error {
	h.rwCore.Lock()
	defer h.rwCore.Unlock()

	for i := 0; i < heapRWSlots; i++ {
		h.rwHeap[i].Lock()
		defer h.rwHeap[i].Unlock()
	}

	if err := h.writeEdges(edges); err != nil {
		return err
	}

	if err := h.writeNodes(nodes); err != nil {
		return err
	}

	if err := h.writeHeader(w); err != nil {
		return err
	}

	return nil
}

func (h *HNSW[Vector]) writeEdges(
	edges interface {
		Write(v []Pointer) error
	},
) error {
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

				if err := edges.Write(iv); err != nil {
					return err
				}
			}
		}
	}

	return nil
}

func (h *HNSW[Vector]) writeNodes(
	nodes interface {
		Write(n Vector) error
	},
) error {
	for _, node := range h.heap {
		if err := nodes.Write(node.Vector); err != nil {
			return err
		}
	}

	return nil
}

func (h *HNSW[Vector]) writeHeader(w io.Writer) error {
	v := header{
		EfConstruction: h.config.efConstruction,
		MLayerN:        h.config.mLayerN,
		MLayer0:        h.config.mLayer0,
		ML:             h.config.mL,
		Head:           h.head,
		Level:          h.level,
	}

	if err := json.NewEncoder(w).Encode(v); err != nil {
		return err
	}

	return nil
}

func (h *HNSW[Vector]) Read(
	r io.Reader,
	nodes interface {
		Read() (Vector, error)
	},
	edges interface {
		Read() ([]Pointer, error)
	},
) error {
	h.rwCore.Lock()
	defer h.rwCore.Unlock()

	for i := 0; i < heapRWSlots; i++ {
		h.rwHeap[i].Lock()
		defer h.rwHeap[i].Unlock()
	}

	if err := h.readNodes(nodes); err != nil {
		return err
	}

	if err := h.readEdges(edges); err != nil {
		return err
	}

	if err := h.readHeader(r); err != nil {
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
				h.heap[addr] = node
			}
		case errors.Is(err, io.EOF):
			return nil
		default:
			return err
		}
	}
}

func (h *HNSW[Vector]) readHeader(r io.Reader) error {
	var v header

	if err := json.NewDecoder(r).Decode(&v); err != nil {
		return err
	}

	h.config.efConstruction = v.EfConstruction
	h.config.mLayerN = v.MLayerN
	h.config.mLayer0 = v.MLayer0
	h.config.mL = v.ML
	h.head = v.Head
	h.level = v.Level

	return nil
}
