//
// Copyright (C) 2024 Dmitry Kolesnikov
//
// This file may be modified and distributed under the terms
// of the MIT license.  See the LICENSE file for details.
// https://github.com/fogfish/hnsw
//

package sift

import (
	"errors"
	"fmt"
	"io"
	"math/rand"
	"os"
	"path/filepath"
	"time"

	"github.com/fogfish/hnsw"
	"github.com/fogfish/hnsw/vector"
	"github.com/kshard/fvecs"
	surface "github.com/kshard/vector"
)

// New HNSW Index for given vector's dimension
func New(m, m0, efC int, seed int64) *hnsw.HNSW[vector.VF32] {
	opts := hnsw.With(
		hnsw.WithEfConstruction(efC),
		hnsw.WithM(m),
		hnsw.WithM0(m0),
	)
	if seed != 0 {
		opts = hnsw.With(opts, hnsw.WithRandomSource(rand.NewSource(seed)))
	}

	return hnsw.New(vector.SurfaceVF32(surface.Euclidean()), opts)
}

func scanner(dataset string, f func(vector.VF32) error) error {
	os.Stderr.WriteString(fmt.Sprintf("==> reading %s\n", dataset))
	fd, err := os.Open(dataset)
	if err != nil {
		return err
	}
	defer fd.Close()

	t := time.Now()
	c := uint32(0)

	progress := func() {
		os.Stderr.WriteString(
			fmt.Sprintf("==> %9d vectors in %s (%d ns/op)\n", c, time.Since(t), time.Since(t).Nanoseconds()/int64(c)),
		)
	}

	scanner := fvecs.NewDecoder[float32](fd)
	for {
		vec, err := scanner.Read()
		if err != nil {
			if errors.Is(err, io.EOF) {
				break
			}
			return nil
		}

		c++
		if err := f(vector.VF32{Key: c, Vec: vec}); err != nil {
			return err
		}

		if c%10000 == 0 {
			progress()
		}
	}

	progress()
	return nil
}

//------------------------------------------------------------------------------

// Create new index from SIFT dataset
func Create(h *hnsw.HNSW[vector.VF32], threads int, dataset string) error {
	w := h.Pipe(threads)

	data := fmt.Sprintf("%s/%s_base.fvecs", dataset, filepath.Base(dataset))
	err := scanner(data,
		func(v vector.VF32) error {
			w <- v
			return nil
		},
	)
	if err != nil {
		return err
	}

	time.Sleep(5 * time.Second)
	return nil
}

//------------------------------------------------------------------------------

func Query(h *hnsw.HNSW[vector.VF32], dataset string) error {
	query := fmt.Sprintf("%s/%s_query.fvecs", dataset, filepath.Base(dataset))

	truth := fmt.Sprintf("%s/%s_groundtruth.ivecs", dataset, filepath.Base(dataset))
	fd, err := os.Open(truth)
	if err != nil {
		return err
	}
	defer fd.Close()

	tv := fvecs.NewDecoder[uint32](fd)

	k := 5
	success := 0
	failure := 0
	err = scanner(query,
		func(v vector.VF32) error {
			ids, err := tv.Read()
			if err != nil {
				return err
			}

			seq := h.Search(v, k, 100)
			i := 0
			s := 0

			for s < len(seq) && i < len(ids) {
				if ids[i] == uint32(seq[s].Key-1) {
					i++
					s++
				} else {
					i++
				}
			}

			if s < k || i > k {
				fmt.Printf("FAIL: drift %2d, seq %2d\n", i-s, s)
				failure++
			} else {
				success++
			}

			return nil
		},
	)
	if err != nil {
		return err
	}

	rate := float64(failure) / float64(success+failure)
	fmt.Printf("==> %d success, %d failures (%2.3f)\n", success, failure, rate)

	return nil
}
