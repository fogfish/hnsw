//
// Copyright (C) 2024 Dmitry Kolesnikov
//
// This file may be modified and distributed under the terms
// of the MIT license.  See the LICENSE file for details.
// https://github.com/fogfish/hnsw
//

// Package num implements numeric computation utilities for HNSW dataset.
package num

import (
	"math"

	"github.com/danaugrs/go-tsne/tsne"
	"github.com/fogfish/hnsw"
	"github.com/fogfish/hnsw/vector"
	"github.com/sjwhitworth/golearn/pca"
	"gonum.org/v1/gonum/mat"
)

// Transform layer of HNSW index to matrix. Each row is vector.
func ToMatrix(level int, h *hnsw.HNSW[vector.VF32]) *mat.Dense {
	var (
		data       []float64
		rows, cols int
	)

	cols = len(h.Head().Vec)

	h.ForAll(level, func(rank int, vector vector.VF32, vertex []vector.VF32) error {
		for _, x := range vector.Vec {
			data = append(data, float64(x))
		}

		rows++
		return nil
	})

	if rows == 0 {
		return nil
	}

	return mat.NewDense(rows, cols, data)
}

// Reduces matrix dimension to (x, y) coordinates using t-SNE algorithms.
// t-SNE params optimized for GLoVe datasets
func Text2D(matrix *mat.Dense) mat.Matrix {
	t := tsne.NewTSNE(2, 50, 10, 300, true)

	if r, _ := matrix.Dims(); r > 5000 {
		pcaTransform := pca.NewPCA(50)
		return t.EmbedData(pcaTransform.FitTransform(matrix), nil)
	}

	return t.EmbedData(matrix, nil)
}

// Reduces matrix dimension to (x, y, z) coordinates using t-SNE algorithms.
// t-SNE params optimized for GLoVe datasets
func Text3D(matrix *mat.Dense) mat.Matrix {
	t := tsne.NewTSNE(3, 50, 10, 300, true)

	if r, _ := matrix.Dims(); r > 5000 {
		pcaTransform := pca.NewPCA(50)
		return t.EmbedData(pcaTransform.FitTransform(matrix), nil)
	}

	return t.EmbedData(matrix, nil)
}

// Calculate min-max value for each coordinate dimension.
// Return matrix - rows are dimensions, cols are (min, max) values
func MinMax(coords mat.Matrix) mat.Matrix {
	r, c := coords.Dims()

	data := make([]float64, c*2)
	minmax := mat.NewDense(c, 2, data)

	for j := 0; j < c; j++ {
		minmax.Set(j, 0, coords.At(0, j))
		minmax.Set(j, 1, coords.At(0, j))
	}

	for i := 1; i < r; i++ {
		for j := 0; j < c; j++ {
			v := coords.At(i, j)

			if v < minmax.At(j, 0) {
				minmax.Set(j, 0, v)
			}

			if v > minmax.At(j, 1) {
				minmax.Set(j, 1, v)
			}
		}
	}

	return minmax
}

// Calculate min-max value for node distances
func MinMaxDistance(level int, h *hnsw.HNSW[vector.VF32]) (min, max float32) {
	minD := math.Inf(1)
	maxD := math.Inf(-1)

	h.ForAll(level, func(rank int, vector vector.VF32, vertex []vector.VF32) error {
		for _, dst := range vertex {
			d := float64(h.Distance(vector, dst))
			if d < minD {
				minD = d
			}
			if d > maxD {
				maxD = d
			}
		}
		return nil
	})

	return float32(minD), float32(maxD)
}
