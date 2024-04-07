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
	"path/filepath"
	"time"

	"github.com/fogfish/hnsw"
	kv "github.com/fogfish/hnsw/vector"
	"github.com/kshard/fvecs"
	"github.com/kshard/vector"
)

// New HNSW Index for given vector's dimension
func New(vs int) *hnsw.HNSW[kv.VF32] {
	return hnsw.New[kv.VF32](
		kv.Surface(vector.Euclidean()),
		kv.Zero(vs),
		hnsw.WithEfConstruction(400),
		hnsw.WithM(16),
	)
}

// Insert dataset
func Insert(h *hnsw.HNSW[kv.VF32], threads int, dataset string) error {
	fmt.Printf("==> reading %s\n", dataset)

	f, err := os.Open(dataset)
	if err != nil {
		return err
	}
	defer f.Close()

	t := time.Now()
	c := uint32(0)

	progress := func() {
		os.Stderr.WriteString(
			fmt.Sprintf("==> read %9d vectors in %s (%d ns/op)\n", c, time.Since(t), time.Since(t).Nanoseconds()/int64(c)),
		)
	}

	d := fvecs.NewDecoder[float32](f)
	w := h.Pipe(threads)

	for {
		vec, err := d.Read()
		switch {
		case err == nil:
			c++

			w <- kv.VF32{Key: c, Vector: vec}
		case errors.Is(err, io.EOF):
			progress()
			return nil
		default:
			return err
		}

		if c%10000 == 0 {
			progress()
		}
	}
}

// Query index comparing with ground truth
func Query(h *hnsw.HNSW[kv.VF32], k int, query []float32, truth []uint32) (int, float64) {
	result := h.Search(kv.VF32{Vector: query}, k, 100)

	errors := 0
	weight := 0.0
	for i, vector := range result {
		if truth[i] != uint32(vector.Key-1) {
			errors++
			weight += float64(k) / float64(i+1)
		}
	}

	if errors > 0 {
		fmt.Printf("FAIL: %2d, %.2f (%.2f %%)\n", errors, weight, 100.0*float32(errors)/float32(len(result)))
	}
	return errors, weight
}

func Test(h *hnsw.HNSW[kv.VF32], dataset string) error {
	fmt.Printf("==> testing dataset %s\n", dataset)

	qf, err := os.Open(fmt.Sprintf("%s/%s_query.fvecs", dataset, filepath.Base(dataset)))
	if err != nil {
		return err
	}

	defer qf.Close()

	tf, err := os.Open(fmt.Sprintf("%s/%s_groundtruth.ivecs", dataset, filepath.Base(dataset)))
	if err != nil {
		return err
	}

	defer tf.Close()

	query := fvecs.NewDecoder[float32](qf)
	truth := fvecs.NewDecoder[uint32](tf)

	t := time.Now()
	c := 0

	errors := 0
	weight := 0.0

	for {
		q, err := query.Read()
		if err != nil {
			break
		}

		t, err := truth.Read()
		if err != nil {
			break
		}

		e, w := Query(h, 10, q, t)
		errors += e
		weight += w

		c++
	}

	fmt.Printf("\n%d queries in %v (%d ns/op)\n", c, time.Since(t), int(time.Since(t).Nanoseconds())/c)

	fmt.Printf("\n%d failed %v (%v)\n", c*10, errors, weight)

	return nil
}
