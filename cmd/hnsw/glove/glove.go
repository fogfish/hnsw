//
// Copyright (C) 2024 Dmitry Kolesnikov
//
// This file may be modified and distributed under the terms
// of the MIT license.  See the LICENSE file for details.
// https://github.com/fogfish/hnsw
//

// Package glove implements testing of HNSW implementation using GLoVe dataset:
//   - https://nlp.stanford.edu/projects/glove/
package glove

import (
	"fmt"
	"math"
	"math/rand"
	"os"
	"time"

	"github.com/fogfish/hnsw"
	"github.com/fogfish/hnsw/vector"
	"github.com/kshard/atom"
	surface "github.com/kshard/vector"
)

// New HNSW Index for given vector's dimension
func New(m, m0, efC int) *hnsw.HNSW[vector.VF32] {
	return hnsw.New(
		vector.SurfaceVF32(surface.Cosine()),
		hnsw.WithEfConstruction(efC),
		hnsw.WithM(m),
		hnsw.WithM0(m0),
		hnsw.WithRandomSource(rand.NewSource(0x123456789)),
	)
}

func scanner(atoms *atom.Pool, dataset string, f func(string, vector.VF32) error) error {
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

	scanner := NewDecoder(fd)
	for scanner.Scan() {
		txt, vec := scanner.Text()
		key, err := atoms.Atom(txt)
		if err != nil {
			return err
		}

		c++
		if err := f(txt, vector.VF32{Key: key, Vec: vec}); err != nil {
			return err
		}

		if c%10000 == 0 {
			progress()
		}
	}
	if err := scanner.Err(); err != nil {
		return err
	}

	progress()
	return nil
}

//------------------------------------------------------------------------------

// Create new index from GLoVe dataset
func Create(atoms *atom.Pool, h *hnsw.HNSW[vector.VF32], threads int, dataset string) error {
	w := h.Pipe(threads)

	err := scanner(atoms, dataset,
		func(txt string, v vector.VF32) error {
			w <- v
			return nil
		},
	)
	if err != nil {
		return err
	}

	close(w)
	return nil
}

//------------------------------------------------------------------------------

// Query dataset, absence of zero distance implies error
func Query(atoms *atom.Pool, h *hnsw.HNSW[vector.VF32], dataset string) error {
	c := 0

	err := scanner(atoms, dataset,
		func(txt string, v vector.VF32) error {
			seq := h.Search(v, 5, 100)

			os.Stdout.WriteString(fmt.Sprintf("\n==> %s\n", txt))
			for i, x := range seq {
				d := h.Distance(v, x)
				os.Stdout.WriteString(fmt.Sprintf("  %16s : %2.5f\n", atoms.String(x.Key), d))

				if i == 0 && math.Abs(float64(d)) > 1e-5 {
					c++
				}
			}

			return nil
		},
	)
	if err != nil {
		return err
	}

	os.Stderr.WriteString(fmt.Sprintf("==> failed %d times\n", c))

	return nil
}
