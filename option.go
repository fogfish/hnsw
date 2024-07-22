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
	"time"

	"math/rand"
)

// HNSW data structure configuration
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

// HNSW data structure configuration option
type Option func(*Config)

func With(opts ...Option) Option {
	return func(c *Config) {
		for _, opt := range opts {
			opt(c)
		}
	}
}

// Construction Efficiency Factor (efConstruction)
//
// The parameter controls the number of candidates evaluated during the graph
// construction phase. Higher values of efConstruction result in a more accurate
// and densely connected graph, leading to better search performance but at
// the cost of longer construction times.
//
// Typical values range from 100 to 500 (default 200).
func WithEfConstruction(ef int) Option {
	return func(c *Config) {
		c.efConstruction = ef
	}
}

// Maximum number of connections per node (M)
//
// This parameter controls the maximum number of neighbors each node can have.
// Higher values of M increase the connectivity and robustness of the graph,
// potentially improving search accuracy but at the cost of increased memory
// usage and longer construction times. Small M gives better result for low
// dimensional data. Big M is better for high dimensional data.
//
// Typical values range from 5 to 48 (default 16).
func WithM(m int) Option {
	return func(c *Config) {
		c.mLayerN = m
	}
}

// Maximum number of connections per node at level 0 (M0)
//
// This parameter controls the maximum number of neighbors each node can have
// at level 0.
//
// Typical value M0 >> M (default M * 2)
func WithM0(m int) Option {
	return func(c *Config) {
		c.mLayer0 = m
	}
}

func WithDefaultM0() Option {
	return func(c *Config) {
		c.mLayer0 = c.mLayerN * 2
	}
}

// Maximum Level Factor
//
// The maximum level a node can be assigned to in this hierarchical structure.
// The level of each node is determined probabilistically based on an
// exponential distribution. The maximum level factor controls the level
// distribution:
//
// P(level) = ⌊-log2(unif(0..1))∙mL⌋
//
// Typical value 1/log2(M) but this option helps to adjust it.
func WithL(l float64) Option {
	return func(c *Config) {
		c.mL = l
	}
}

func WithDefaultL() Option {
	return func(c *Config) {
		c.mL = 1 / math.Log2(1.0*float64(c.mLayerN))
	}
}

// Random Source
//
// Uniform random source for seeding exponential distribution.
func WithRandomSource(random rand.Source) Option {
	return func(c *Config) {
		c.random = random
	}
}

// Default options
func WithDefault() Option {
	return With(
		WithEfConstruction(200),
		WithM(16),
		WithDefaultM0(),
		WithDefaultL(),
		WithRandomSource(rand.NewSource(time.Now().UnixNano())),
	)
}
