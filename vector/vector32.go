//
// Copyright (C) 2024 Dmitry Kolesnikov
//
// This file may be modified and distributed under the terms
// of the MIT license.  See the LICENSE file for details.
// https://github.com/fogfish/hnsw
//

package vector

import (
	"math"
)

// Vector of Floats32
type V32 = []float32

// Euclidean surface for vector of floats32
type euclidean string

func (euclidean) Distance(a V32, b V32) (d float32) {
	for i := 0; i < len(a); i++ {
		d += (a[i] - b[i]) * (a[i] - b[i])
	}
	return
}

const Euclidean = euclidean("")

// Cosine surface for vector of floats32
type cosine string

func (cosine) Distance(a V32, b V32) (d float32) {
	// https://en.wikipedia.org/wiki/Cosine_similarity

	ab := 0.0
	aa := 0.0
	bb := 0.0

	for i := 0; i < len(a); i++ {
		ab += float64(a[i] * b[i])
		aa += float64(a[i] * a[i])
		bb += float64(b[i] * b[i])
	}

	d = float32(ab / (math.Sqrt(aa) * math.Sqrt(bb)))

	return
}

const Cosine = cosine("")
