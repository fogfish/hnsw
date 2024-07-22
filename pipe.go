//
// Copyright (C) 2024 Dmitry Kolesnikov
//
// This file may be modified and distributed under the terms
// of the MIT license.  See the LICENSE file for details.
// https://github.com/fogfish/hnsw
//

package hnsw

// Create pipe ("channel") for  batch insert
//
//	ch := index.Pipe(runtime.NumCPU())
//	ch <- vector.VF32{Key: 1, Vec: []float32{0.1, 0.2, /* ... */ 0.128}}
//
// The HNSW library supports batch insert operations, making it efficient to
// add large datasets. It leverages Golang channels to handle parallel writes,
// ensuring that multiple data points can be inserted concurrently.
func (h *HNSW[Vector]) Pipe(workers int) chan<- Vector {
	pipe := make(chan Vector, workers)

	for i := 1; i <= workers; i++ {
		go func() {
			for v := range pipe {
				h.Insert(v)
			}
		}()
	}

	return pipe
}
