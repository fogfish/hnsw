//
// Copyright (C) 2024 Dmitry Kolesnikov
//
// This file may be modified and distributed under the terms
// of the MIT license.  See the LICENSE file for details.
// https://github.com/fogfish/hnsw
//

package pq

import (
	"container/heap"
)

type Ord[T any] interface{ Compare(T, T) int }

type Queue[T any] struct {
	heap *heaps[T]
}

func New[T any](ord Ord[T], seq ...T) Queue[T] {
	mm := [400]T{}
	pq := Queue[T]{
		heap: &heaps[T]{
			ord: ord,
			mem: mm[0:0:400], // make([]T, 0),
		},
	}

	for _, x := range seq {
		pq.Enq(x)
	}

	return pq
}

func (q Queue[T]) Len() int {
	return len(q.heap.mem)
}

func (q Queue[T]) Head() T {
	return q.heap.mem[0]
}

func (q Queue[T]) Enq(v T) {
	// avoid unnecessary casting of T
	q.heap.mem = append(q.heap.mem, v)
	heap.Push(q.heap, 1)
}

func (q Queue[T]) Deq() T {
	// avoid unnecessary casting of T
	heap.Pop(q.heap)

	mem, n := q.heap.maybeShrink()
	item := (*mem)[n-1]
	q.heap.mem = (*mem)[0 : n-1]
	return item
}

//
//
//

const shrinkMinCap = 1000
const shrinkNewSizeFactor = 2
const shrinkCapLenFactorCondition = 4

type heaps[T any] struct {
	ord Ord[T]
	mem []T
}

func (h *heaps[T]) Len() int {
	return len(h.mem)
}

func (h *heaps[T]) Less(i, j int) bool {
	return h.ord.Compare(h.mem[i], h.mem[j]) == -1
}

func (h *heaps[T]) Swap(i, j int) {
	h.mem[i], h.mem[j] = h.mem[j], h.mem[i]
}

func (h *heaps[T]) Push(x any) {}

func (h *heaps[T]) Pop() any {
	return nil
}

func (h *heaps[T]) maybeShrink() (*[]T, int) {
	l, c := len(h.mem), cap(h.mem)
	if cap(h.mem) > shrinkMinCap && c/l > shrinkCapLenFactorCondition {
		mem := make([]T, shrinkNewSizeFactor*l)
		copy(mem, h.mem)
		return &mem, l
	}
	return &h.mem, l
}
