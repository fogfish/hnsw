//
// Copyright (C) 2024 Dmitry Kolesnikov
//
// This file may be modified and distributed under the terms
// of the MIT license.  See the LICENSE file for details.
// https://github.com/fogfish/hnsw
//

package pq_test

import (
	"math/rand"
	"testing"

	"github.com/fogfish/hnsw/internal/pq"
	"github.com/fogfish/it/v2"
)

type E struct {
	weight int
	value  int
}

type ordE string

func (ordE) Compare(a, b E) int {
	if a.weight < b.weight {
		return -1
	}

	if a.weight > b.weight {
		return 1
	}

	return 0
}

const SIZE = 500000

func TestPQ(t *testing.T) {
	vl := 0
	pq := pq.New(ordE(""))
	mw := 0

	for i := 0; i < SIZE; i++ {
		w := rand.Intn(20)
		if w > mw {
			mw = w
		}
		pq.Enq(E{weight: w, value: i})
		vl += i
	}

	priority := 0

	for i := 0; i < SIZE; i++ {
		e := pq.Deq()

		it.Then(t).ShouldNot(
			it.Less(e.weight, priority),
		)

		priority = e.weight
		vl -= e.value
	}

	it.Then(t).Should(
		it.Equal(vl, 0),
	)
}
