//
// Copyright (C) 2024 Dmitry Kolesnikov
//
// This file may be modified and distributed under the terms
// of the MIT license.  See the LICENSE file for details.
// https://github.com/fogfish/hnsw
//

package try

import (
	"errors"
	"fmt"
	"io"
	"os"
	"time"

	"github.com/fogfish/hnsw"
	"github.com/fogfish/hnsw/internal/fvecs"
	"github.com/fogfish/hnsw/vector"
)

type Node struct {
	ID     int
	Vector vector.V32
}

func New() *hnsw.HNSW[Node] {
	surface := vector.ContraMap[vector.V32, Node]{
		Surface:   vector.Euclidean,
		ContraMap: func(n Node) []float32 { return n.Vector },
	}

	zero := Node{ID: 0, Vector: make(vector.V32, 128)}

	return hnsw.New[Node](surface, zero)
}

func Create(h *hnsw.HNSW[Node], dataset string) error {
	fmt.Printf("==> reading dataset %s\n", dataset)

	f, err := os.Open(fmt.Sprintf("%s/%s_base.fvecs", dataset, dataset))
	if err != nil {
		return err
	}
	defer f.Close()

	t := time.Now()
	c := 1
	d := fvecs.NewDecoder[float32](f)
	for {
		vec, err := d.Read()
		switch {
		case err == nil:
			h.Insert(Node{ID: c, Vector: vec})
		case errors.Is(err, io.EOF):
			return nil
		default:
			return err
		}

		c++

		if c%1000 == 0 {
			fmt.Printf("==> read %9d vectors in %s (%d ns/op)\n", c, time.Since(t), int(time.Since(t).Nanoseconds())/c)
		}
	}
}

func Query(h *hnsw.HNSW[Node], query []float32, truth []uint32) {
	result := h.Search(Node{Vector: query}, 10, 100)

	errors := 0
	for i, vector := range result {
		if truth[i] != uint32(vector.ID-1) {
			errors++
		}
	}

	if errors > 0 {
		fmt.Printf("FAIL: %2d of  %2d (%.2f %%)\n", errors, len(result), 100.0*float32(errors)/float32(len(result)))
	}
}

func Test(h *hnsw.HNSW[Node], dataset string) error {
	fmt.Printf("==> testing dataset %s\n", dataset)

	qf, err := os.Open(fmt.Sprintf("%s/%s_query.fvecs", dataset, dataset))
	if err != nil {
		return err
	}

	defer qf.Close()

	tf, err := os.Open(fmt.Sprintf("%s/%s_groundtruth.ivecs", dataset, dataset))
	if err != nil {
		return err
	}

	defer tf.Close()

	query := fvecs.NewDecoder[float32](qf)
	truth := fvecs.NewDecoder[uint32](tf)

	t := time.Now()
	c := 0

	for {
		q, err := query.Read()
		if err != nil {
			break
		}

		t, err := truth.Read()
		if err != nil {
			break
		}

		Query(h, q, t)
		c++
	}

	fmt.Printf("\n%d queries in %v (%d ns/op)\n", c, time.Since(t), int(time.Since(t).Nanoseconds())/c)
	return nil
}
