package kv

import (
	"encoding/binary"
	"os"
	"strconv"

	"github.com/fogfish/hnsw"
	"github.com/kshard/fvecs"
	"github.com/kshard/vector"
)

type Vector struct {
	Key    uint32
	Vector vector.F32
}

func (v Vector) String() string { return strconv.Itoa(int(v.Key)) }

func toVector(n Vector) vector.F32 { return n.Vector }

func Zero(vs int) Vector {
	return Vector{Key: 0, Vector: make(vector.F32, vs)}
}

func Surface(surface vector.Surface[vector.F32]) vector.Surface[Vector] {
	return vector.ContraMap[vector.F32, Vector]{
		Surface:   surface,
		ContraMap: toVector,
	}
}

func Write(h *hnsw.HNSW[Vector], file string) error {
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

	nw := NewWriter(fe, be)

	return h.Write(nw, ie)
}

func Read(h *hnsw.HNSW[Vector], file string) error {
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

	nr := NewReader(fd, bd)

	return h.Read(nr, id)
}

//
//
//

type FWriter interface {
	Write(vector.F32) error
}

type BWriter interface {
	Write([]byte) error
}

type Writer struct {
	fw FWriter
	bw BWriter
	b  []byte
}

func NewWriter(fw FWriter, bw BWriter) Writer {
	return Writer{fw: fw, bw: bw, b: []byte{0, 0, 0, 0}}
}

func (w Writer) Write(v Vector) error {
	if err := w.fw.Write(v.Vector); err != nil {
		return err
	}

	binary.LittleEndian.PutUint32(w.b, v.Key)
	if err := w.bw.Write(w.b); err != nil {
		return err
	}

	return nil
}

//
//
//

type FReader interface {
	Read() (vector.F32, error)
}

type BReader interface {
	Read() ([]byte, error)
}

type Reader struct {
	fr FReader
	br BReader
}

func NewReader(fr FReader, br BReader) Reader {
	return Reader{fr: fr, br: br}
}

func (r Reader) Read() (v Vector, err error) {
	v.Vector, err = r.fr.Read()
	if err != nil {
		return
	}

	b, err := r.br.Read()
	if err != nil {
		return
	}

	v.Key = binary.LittleEndian.Uint32(b)

	// fmt.Printf("%v\n", v)

	return
}
