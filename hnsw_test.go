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
	"github.com/kshard/vector"
)

const n = 128

func vZero() vector.F32 {
	v := make(vector.F32, n)
	for i := 0; i < n; i++ {
		v[i] = 0
	}
	return v
}

func vRand() vector.F32 {
	v := make(vector.F32, n)
	for i := 0; i < n; i++ {
		v[i] = rand.Float32()
	}
	return v
}

func BenchmarkInsert(b *testing.B) {
	h := hnsw.New[vector.F32](
		vector.Euclidean(),
		// vector.Cosine,
		vZero(),
		hnsw.WithEfConstruction(400),
		hnsw.WithM(16),
	)

	b.ReportAllocs()

	for n := b.N; n > 0; n-- {
		h.Insert(vRand())
	}
}
