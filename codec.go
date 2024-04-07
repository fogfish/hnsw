//
// Copyright (C) 2024 Dmitry Kolesnikov
//
// This file may be modified and distributed under the terms
// of the MIT license.  See the LICENSE file for details.
// https://github.com/fogfish/hnsw
//

package hnsw

import (
	"github.com/fogfish/faults"
	"github.com/kelindar/binary"
)

const (
	errIO    = faults.Type("i/o error")
	errCodec = faults.Type("codec failed")
)

// Getter interface abstract storage
type Reader interface{ Get([]byte) ([]byte, error) }

// Setter interface abstract storage
type Writer interface{ Put([]byte, []byte) error }

type header struct {
	EfConstruction int
	MLayerN        int
	MLayer0        int
	ML             float64
	Size           int
	Head           Pointer
	Level          int
}

// Write index to storage
func (h *HNSW[Vector]) Write(w Writer) error {
	h.rwCore.Lock()
	defer h.rwCore.Unlock()

	for i := 0; i < heapRWSlots; i++ {
		h.rwHeap[i].Lock()
		defer h.rwHeap[i].Unlock()
	}

	if err := h.writeHeader(w); err != nil {
		return err
	}

	if err := h.writeNodes(w); err != nil {
		return err
	}

	return nil
}

func (h *HNSW[Vector]) writeHeader(w Writer) error {
	v := header{
		EfConstruction: h.config.efConstruction,
		MLayerN:        h.config.mLayerN,
		MLayer0:        h.config.mLayer0,
		ML:             h.config.mL,
		Size:           len(h.heap),
		Head:           h.head,
		Level:          h.level,
	}

	b, err := binary.Marshal(v)
	if err != nil {
		return errCodec.New(err)
	}

	err = w.Put([]byte("&root"), b)
	if err != nil {
		return errIO.New(err)
	}

	return nil
}

func (h *HNSW[Vector]) writeNodes(w Writer) error {
	var bkey [5]byte
	bkey[0] = '&'

	for key, node := range h.heap {
		binary.LittleEndian.PutUint32(bkey[1:], uint32(key))

		b, err := binary.Marshal(node)
		if err != nil {
			return errCodec.New(err)
		}

		err = w.Put(bkey[:], b)
		if err != nil {
			return errIO.New(err)
		}
	}

	return nil
}

func (h *HNSW[Vector]) Read(r Reader) error {
	h.rwCore.Lock()
	defer h.rwCore.Unlock()

	for i := 0; i < heapRWSlots; i++ {
		h.rwHeap[i].Lock()
		defer h.rwHeap[i].Unlock()
	}

	if err := h.readHeader(r); err != nil {
		return err
	}

	if err := h.readNodes(r); err != nil {
		return err
	}

	return nil
}

func (h *HNSW[Vector]) readHeader(r Reader) error {
	var v header

	b, err := r.Get([]byte("&root"))
	if err != nil {
		return errIO.New(err)
	}

	if err := binary.Unmarshal(b, &v); err != nil {
		return errCodec.New(err)
	}

	h.config.efConstruction = v.EfConstruction
	h.config.mLayerN = v.MLayerN
	h.config.mLayer0 = v.MLayer0
	h.config.mL = v.ML
	h.heap = make([]Node[Vector], v.Size)
	h.head = v.Head
	h.level = v.Level

	return nil
}

func (h *HNSW[Vector]) readNodes(r Reader) error {
	var bkey [5]byte
	bkey[0] = '&'

	for key := 0; key < len(h.heap); key++ {
		binary.LittleEndian.PutUint32(bkey[1:], uint32(key))

		b, err := r.Get(bkey[:])
		if err != nil {
			return errIO.New(err)
		}

		if err := binary.Unmarshal(b, &h.heap[key]); err != nil {
			return errCodec.New(err)
		}
	}

	return nil
}
