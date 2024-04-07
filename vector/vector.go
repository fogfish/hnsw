//
// Copyright (C) 2024 Dmitry Kolesnikov
//
// This file may be modified and distributed under the terms
// of the MIT license.  See the LICENSE file for details.
// https://github.com/fogfish/hnsw
//

package vector

import (
	"strconv"

	"github.com/kshard/vector"
)

// Vector of float32 annotated with uint32 key
type VF32 struct {
	Key    uint32     `json:"key"`
	Vector vector.F32 `json:"vector"`
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
