//
// Copyright (C) 2024 Dmitry Kolesnikov
//
// This file may be modified and distributed under the terms
// of the MIT license.  See the LICENSE file for details.
// https://github.com/fogfish/hnsw
//

package hnsw

import (
	"math"
	"math/rand"
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

// Slots to coordinate concurrent I/O
const heapRWSlots = 1024

// HNSW data type
type HNSW[Vector any] struct {
	rwCore sync.RWMutex
	rwHeap [heapRWSlots]sync.RWMutex

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

// Create data structure FromNodes
func FromNodes[Vector any](
	surface vector.Surface[Vector],
	nodes Nodes[Vector],
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

	hnsw.level = nodes.Rank
	hnsw.heap = nodes.Heap
	hnsw.head = nodes.Head

	return hnsw
}

func (h *HNSW[Vector]) Nodes() Nodes[Vector] {
	return Nodes[Vector]{
		Rank: h.level,
		Head: h.head,
		Heap: h.heap,
	}
}

func (h *HNSW[Vector]) Level() int { return h.level }

func (h *HNSW[Vector]) Distance(a, b Vector) float32 {
	return h.surface.Distance(a, b)
}
