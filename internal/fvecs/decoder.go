//
// Copyright (C) 2024 Dmitry Kolesnikov
//
// This file may be modified and distributed under the terms
// of the MIT license.  See the LICENSE file for details.
// https://github.com/fogfish/hnsw
//

package fvecs

import (
	"encoding/binary"
	"io"
	"math"
)

type Decoder[T float32 | uint32] struct {
	r      io.Reader
	reader func() (T, error)
}

func NewDecoder[T float32 | uint32](r io.Reader) Decoder[T] {
	d := Decoder[T]{r: r}

	switch any(*new(T)).(type) {
	case float32:
		d.reader = func() (T, error) {
			v, err := d.float32()
			return T(v), err
		}
	case uint32:
		d.reader = func() (T, error) {
			v, err := d.uint32()
			return T(v), err
		}
	}

	return d
}

func (d *Decoder[T]) uint32() (uint32, error) {
	bs := make([]byte, 4)
	_, err := d.r.Read(bs)
	return binary.LittleEndian.Uint32(bs), err
}

func (d *Decoder[T]) float32() (float32, error) {
	bs := make([]byte, 4)
	_, err := d.r.Read(bs)
	return float32(math.Float32frombits(binary.LittleEndian.Uint32(bs))), err
}

func (d *Decoder[T]) Read() ([]T, error) {
	s, err := d.uint32()
	if err != nil {
		return nil, err
	}

	v := make([]T, s)
	for i := 0; i < int(s); i++ {
		v[i], err = d.reader()
		if err != nil {
			return nil, err
		}
	}

	return v, nil
}
