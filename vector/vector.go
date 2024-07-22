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

	"github.com/fogfish/guid/v2"
	"github.com/kshard/vector"
)

// Vector of float32 annotated with uint32 key
type VF32 struct {
	Key uint32     `json:"k"`
	Vec vector.F32 `json:"v"`
}

func (v VF32) String() string { return strconv.Itoa(int(v.Key)) }

// Create surface distance function for type VF32
func SurfaceVF32(surface vector.Surface[vector.F32]) vector.Surface[VF32] {
	return vector.ContraMap[vector.F32, VF32]{
		Surface:   surface,
		ContraMap: func(e VF32) vector.F32 { return e.Vec },
	}
}

//------------------------------------------------------------------------------

// Vector of float32 annotated with K-order number
type KF32 struct {
	Key guid.K     `json:"k"`
	Vec vector.F32 `json:"v"`
}

func (v KF32) String() string { return v.Key.String() }

// Create surface distance function for type VF32
func SurfaceKF32(surface vector.Surface[vector.F32]) vector.Surface[KF32] {
	return vector.ContraMap[vector.F32, KF32]{
		Surface:   surface,
		ContraMap: func(e KF32) vector.F32 { return e.Vec },
	}
}
