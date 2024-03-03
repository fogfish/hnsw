//
// Copyright (C) 2024 Dmitry Kolesnikov
//
// This file may be modified and distributed under the terms
// of the MIT license.  See the LICENSE file for details.
// https://github.com/fogfish/hnsw
//

package hnsw

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
