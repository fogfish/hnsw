//
// Copyright (C) 2024 Dmitry Kolesnikov
//
// This file may be modified and distributed under the terms
// of the MIT license.  See the LICENSE file for details.
// https://github.com/fogfish/hnsw
//

package hnsw

import (
	"fmt"
	"math"
	"math/rand"
	"strings"
	"sync"
	"time"

	"github.com/kshard/vector"
)

// Config of the HNSW
type Config struct {
	// size of the dynamic candidate list efConstruction.
	efConstruction int

	// Number of established connections from each node.
	mLayerN int
	mLayer0 int

	// Normalization factor for level generation
	mL float64

	//
	random rand.Source
}

// Config Options
type Option func(*Config)

func With(opts ...Option) Option {
	return func(c *Config) {
		for _, opt := range opts {
			opt(c)
		}
	}
}

// Configure size of dynamic candidate list
func WithEfConstruction(ef int) Option {
	return func(c *Config) {
		c.efConstruction = ef
	}
}

// Configures number of established connections from each node.
// M in [5 to 48] is recommended range. Small M gives better result for low
// dimensional data. Big M is better for high dimensional data.
// High M impacts on memory consumption
func WithM(m int) Option {
	return func(c *Config) {
		c.mLayerN = m
		c.mLayer0 = m * 2
		c.mL = 1 / math.Log(1.0*float64(m))
	}
}

// Configure Random Source
func WithRandomSource(random rand.Source) Option {
	return func(c *Config) {
		c.random = random
	}
}

// HNSW data type
type HNSW[Vector any] struct {
	sync.RWMutex
	config  Config
	surface vector.Surface[Vector]

	heap  []Node[Vector]
	head  Pointer
	level int
}

// Creates new instance of data structure
func New[Vector any](
	surface vector.Surface[Vector],
	zero Vector,
	opts ...Option,
) *HNSW[Vector] {
	config := Config{}
	def := With(
		WithEfConstruction(100),
		WithM(16),
		WithRandomSource(rand.NewSource(time.Now().UnixNano())),
	)

	def(&config)
	for _, opt := range opts {
		opt(&config)
	}

	hnsw := &HNSW[Vector]{
		config:  config,
		surface: surface,
	}

	node := Node[Vector]{
		Vector:      zero,
		Connections: make([][]Pointer, hnsw.config.mLayer0+1),
	}

	hnsw.level = len(node.Connections)
	hnsw.heap = []Node[Vector]{node}
	hnsw.head = 0

	return hnsw
}

func (h *HNSW[Vector]) Level() int { return h.level }

//
//
//

// func (h *HNSW[Vector]) Head() Pointer                  { return h.head }
// func (h *HNSW[Vector]) Node(addr Pointer) Node[Vector] { return h.heap[addr] }

func (h *HNSW[Vector]) Dump() {
	sb := strings.Builder{}

	for lvl := h.level - 1; lvl >= 0; lvl-- {
		sb.WriteString(fmt.Sprintf("\n\n==> %v\n", lvl))

		h.FMap(lvl, func(level int, vector Vector, vertex []Vector) error {

			sb.WriteString(fmt.Sprintf("%v | ", vector))
			for _, e := range vertex {
				sb.WriteString(fmt.Sprintf("%v ", e))
			}
			sb.WriteString("\n")

			return nil
		})

		// 		visited := map[Pointer]struct{}{}

		// 		sb.WriteString(fmt.Sprintf("\n\n==> %v\n", lvl))
		// 		h.dump(&sb, lvl, visited, h.head)
	}

	fmt.Println(sb.String())
}

// func (h *HNSW[Vector]) dump(sb *strings.Builder, level int, visited map[Pointer]struct{}, addr Pointer) {
// 	if _, has := visited[addr]; has {
// 		return
// 	}

// 	visited[addr] = struct{}{}

// 	sb.WriteString(fmt.Sprintf("%v | ", h.heap[addr].Vector))
// 	for _, e := range h.heap[addr].Connections[level] {
// 		sb.WriteString(fmt.Sprintf("%v ", h.heap[e].Vector))
// 	}
// 	sb.WriteString("\n")

// 	for _, e := range h.heap[addr].Connections[level] {
// 		h.dump(sb, level, visited, e)
// 	}
// }
