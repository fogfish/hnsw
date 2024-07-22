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
	rnd = rand.NewSource(0x211111111)
)

func random() float32 {
again:
	f := float32(rnd.Int63()) / (1 << 63)
	if f == 1 {
		goto again // resample; this branch is taken O(never)
	}
	return f
}

func vRand() surface.F32 {
	v := make(surface.F32, d)
	for i := 0; i < d; i++ {
		v[i] = random()
	}
	return v
}

func TestHNSW(t *testing.T) {
	vecs := make([]surface.F32, n)
	for i := 0; i < n; i++ {
		vecs[i] = vRand()
	}

	index := hnsw.New(
		vector.SurfaceVF32(surface.Euclidean()),
		hnsw.WithRandomSource(rnd),
	)

	for i, v := range vecs {
		index.Insert(vector.VF32{Key: uint32(i), Vec: v})
	}

	nodes := make([]vector.VF32, 0)
	index.ForAll(0,
		func(rank int, vector vector.VF32, edges []vector.VF32) error {
			nodes = append(nodes, vector)
			return nil
		},
	)

	for _, q := range nodes {
		seq := index.Search(q, 4, 100)
		if seq[0].Key != q.Key {
			t.Errorf("Not found %v in %v", q, seq)
		}
	}
}
