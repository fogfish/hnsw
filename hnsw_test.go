//
// Copyright (C) 2024 Dmitry Kolesnikov
//
// This file may be modified and distributed under the terms
// of the MIT license.  See the LICENSE file for details.
// https://github.com/fogfish/hnsw
//

package hnsw_test

import (
	"math/rand"
	"testing"

	"github.com/fogfish/hnsw"
	"github.com/fogfish/hnsw/vector"
	surface "github.com/kshard/vector"
)

const (
	d = 128
	n = 1000
)

var (
	rnd     = rand.NewSource(0x211111111)
	vectors = rndVectors()
)

func sut(surface surface.Surface[surface.F32]) *hnsw.HNSW[vector.VF32] {
	return hnsw.New(
		vector.SurfaceVF32(surface),
		hnsw.WithRandomSource(rnd),
		hnsw.WithM0(64),
	)
}

func TestInsert(t *testing.T) {
	for _, df := range []surface.Surface[surface.F32]{
		surface.Euclidean(),
		surface.Cosine(),
	} {
		index := sut(df)
		for i, v := range vectors {
			index.Insert(vector.VF32{Key: uint32(i), Vec: v})
		}

		for _, q := range nodes(index) {
			seq := index.Search(q, 1, 100)
			if seq[0].Key != q.Key {
				t.Errorf("Not found %v in %v", q, seq)
			}
		}
	}
}

func TestUpdate(t *testing.T) {
	for _, df := range []surface.Surface[surface.F32]{
		surface.Euclidean(),
		surface.Cosine(),
	} {
		index := sut(df)
		for i, v := range vectors {
			index.Insert(vector.VF32{Key: uint32(i), Vec: v})
		}

		// Update vectors (set new key)
		for i, v := range vectors {
			key := uint32((1 << 31) | i)
			index.Insert(vector.VF32{Key: key, Vec: v})
		}

		// Check that key is updated
		for _, n := range nodes(index) {
			if (n.Key & (1 << 31)) == 0 {
				t.Errorf("Not updated %v", n.Key)
			}
		}
	}
}

//------------------------------------------------------------------------------

func random() float32 {
again:
	f := float64(rnd.Int63()) / (1 << 63)
	if f == 1 {
		goto again // resample; this branch is taken O(never)
	}
	return float32(f)
}

func rndVector() surface.F32 {
	v := make(surface.F32, d)
	for i := 0; i < d; i++ {
		v[i] = 2*random() - 1
	}
	return v
}

func rndVectors() []surface.F32 {
	vecs := make([]surface.F32, n)
	for i := 0; i < n; i++ {
		vecs[i] = rndVector()
	}
	return vecs
}

func nodes(index *hnsw.HNSW[vector.VF32]) []vector.VF32 {
	nodes := make([]vector.VF32, 0)
	index.ForAll(0,
		func(rank int, vector vector.VF32, edges []vector.VF32) error {
			nodes = append(nodes, vector)
			return nil
		},
	)
	return nodes
}
