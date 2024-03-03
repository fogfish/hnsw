//
// Copyright (C) 2024 Dmitry Kolesnikov
//
// This file may be modified and distributed under the terms
// of the MIT license.  See the LICENSE file for details.
// https://github.com/fogfish/hnsw
//

package vector

import (
	"encoding/binary"
	"os"
	"strconv"

	"github.com/fogfish/hnsw"
	"github.com/kshard/fvecs"
	"github.com/kshard/vector"
)

// Vector of float32 annotated with uint32 key
type VF32 struct {
	Key    uint32
	Vector vector.F32
}

func (v VF32) String() string { return strconv.Itoa(int(v.Key)) }

func toVector(n VF32) vector.F32 { return n.Vector }

// Zero vector of float32
func Zero(vs int) VF32 {
	return VF32{Key: 0, Vector: make(vector.F32, vs)}
}

// Surface defines distance measurement rules
func Surface(surface vector.Surface[vector.F32]) vector.Surface[VF32] {
	return vector.ContraMap[vector.F32, VF32]{
		Surface:   surface,
		ContraMap: toVector,
	}
}

// Write HNSW index to file. The structure is written into three files.
// *.fvecs - sequence of vectors (vector.F32), ordered by insert time (as preserved by data structure)
// *.bvecs - sequence of vectors ids (uint32), ordered by insert time (as sequence of vectors)
// *.ivecs - sequence of edges ([]Pointer) for each level between vectors
func Write(h *hnsw.HNSW[VF32], file string) error {
	hw, err := os.Create(file + ".json")
	if err != nil {
		return err
	}
	defer hw.Close()

	fw, err := os.Create(file + ".fvecs")
	if err != nil {
		return err
	}
	defer fw.Close()

	iw, err := os.Create(file + ".ivecs")
	if err != nil {
		return err
	}
	defer iw.Close()

	bw, err := os.Create(file + ".bvecs")
	if err != nil {
		return err
	}
	defer bw.Close()

	fe := fvecs.NewEncoder[float32](fw)
	ie := fvecs.NewEncoder[uint32](iw)
	be := fvecs.NewEncoder[byte](bw)

	if err := h.Write(hw, newWriter(fe, be), ie); err != nil {
		return err
	}

	return nil
}

// Read HNSW index from file.
func Read(h *hnsw.HNSW[VF32], file string) error {
	hr, err := os.Open(file + ".json")
	if err != nil {
		return err
	}
	defer hr.Close()

	fr, err := os.Open(file + ".fvecs")
	if err != nil {
		return err
	}
	defer fr.Close()

	ir, err := os.Open(file + ".ivecs")
	if err != nil {
		return err
	}
	defer ir.Close()

	br, err := os.Open(file + ".bvecs")
	if err != nil {
		return err
	}
	defer br.Close()

	fd := fvecs.NewDecoder[float32](fr)
	id := fvecs.NewDecoder[uint32](ir)
	bd := fvecs.NewDecoder[byte](br)

	if err := h.Read(hr, newReader(fd, bd), id); err != nil {
		return err
	}

	return nil
}

//
//
//

type Writer[T any] interface {
	Write(T) error
}

type writer struct {
	floats Writer[vector.F32]
	bytes  Writer[[]byte]
	b      []byte
}

func newWriter(floats Writer[vector.F32], bytes Writer[[]byte]) Writer[VF32] {
	return writer{
		floats: floats,
		bytes:  bytes,
		b:      []byte{0, 0, 0, 0},
	}
}

func (w writer) Write(v VF32) error {
	if err := w.floats.Write(v.Vector); err != nil {
		return err
	}

	binary.LittleEndian.PutUint32(w.b, v.Key)
	if err := w.bytes.Write(w.b); err != nil {
		return err
	}

	return nil
}

//
//
//

type Reader[T any] interface {
	Read() (T, error)
}

type reader struct {
	floats Reader[vector.F32]
	bytes  Reader[[]byte]
}

func newReader(floats Reader[vector.F32], bytes Reader[[]byte]) Reader[VF32] {
	return reader{
		floats: floats,
		bytes:  bytes,
	}
}

func (r reader) Read() (v VF32, err error) {
	v.Vector, err = r.floats.Read()
	if err != nil {
		return
	}

	b, err := r.bytes.Read()
	if err != nil {
		return
	}

	v.Key = binary.LittleEndian.Uint32(b)

	return
}
