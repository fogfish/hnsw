//
// Copyright (C) 2024 Dmitry Kolesnikov
//
// This file may be modified and distributed under the terms
// of the MIT license.  See the LICENSE file for details.
// https://github.com/fogfish/hnsw
//

package viz

import (
	"math"

	"github.com/fogfish/hnsw"
	"github.com/fogfish/hnsw/cmd/hnsw/num"
	kv "github.com/fogfish/hnsw/vector"
	"github.com/go-echarts/go-echarts/v2/charts"
	"github.com/go-echarts/go-echarts/v2/components"
	"github.com/go-echarts/go-echarts/v2/opts"
)

var (
	schemaColor = []string{"#1F4B99", "#3A77A3", "#60A2B3", "#95CACB", "#D8EEEC", "#FDE7D7", "#F0B790", "#D98955", "#BD5C29", "#9E2B0E", "#DCE0E5"}
	schemaLabel = []string{"0.05", "0.1", "0.15", "0.2", "0.25", "0.3", "0.35", "0.4", "0.45", "0.5", "high"}

	schemaScatter3D = []string{"#1F4B99", "#2B5E9C", "#38709E", "#48819F", "#5B92A1", "#71A3A2", "#8AB3A2", "#A7C3A2", "#C7D1A1", "#EBDDA0", "#FCD993", "#F5C57D", "#EDB269", "#E49F57", "#DA8C46", "#CF7937", "#C4662A", "#B8541E", "#AB4015", "#9E2B0E"}
)

func Categories() []*opts.GraphCategory {
	kinds := []*opts.GraphCategory{}

	for i, c := range schemaLabel {
		kinds = append(kinds,
			&opts.GraphCategory{
				Name: c,
				ItemStyle: &opts.ItemStyle{
					Color: schemaColor[i],
				},
			},
		)
	}

	return kinds
}

//------------------------------------------------------------------------------

// Visualize layer of HNSW as graph
func Visualize(level int, h *hnsw.HNSW[kv.VF32], upScaleD func(float32) float32, stringify func(kv.VF32) string) ([]opts.GraphNode, []opts.GraphLink) {
	xy := num.Text2D(num.ToMatrix(level, h))
	minmax := num.MinMax(xy)
	spanX := minmax.At(0, 1) - minmax.At(0, 0)
	spanY := minmax.At(1, 1) - minmax.At(1, 0)

	upScaleX := func(row int) float32 {
		return float32(((xy.At(row, 0) - minmax.At(0, 0)) * 1024.0) / spanX)
	}

	upScaleY := func(row int) float32 {
		return float32(((xy.At(row, 1) - minmax.At(1, 0)) * 1024.0) / spanY)
	}

	nodes := []opts.GraphNode{}
	links := []opts.GraphLink{}

	row := 0

	h.ForAll(level, func(rank int, vector kv.VF32, vertex []kv.VF32) error {
		weight := float32(0.0)

		for _, dst := range vertex {
			d := upScaleD(h.Distance(vector, dst))
			weight += d

			links = append(links,
				opts.GraphLink{
					Source:    stringify(vector),
					Target:    stringify(dst),
					Value:     d,
					LineStyle: &opts.LineStyle{Color: colorOf(d)},
				},
			)
		}

		weight = weight / float32(len(vertex))

		nodes = append(nodes,
			opts.GraphNode{
				Name:       stringify(vector),
				X:          upScaleX(row),
				Y:          upScaleY(row),
				Category:   categoryOf(weight),
				SymbolSize: []int{4, 4},
			},
		)
		row++

		return nil
	})

	return nodes, links
}

//------------------------------------------------------------------------------

// Visualize layer of HNSW as scattered points
func Visualize3D(level int, h *hnsw.HNSW[kv.VF32], stringify func(kv.VF32) string) []opts.Chart3DData {
	xy := num.Text3D(num.ToMatrix(level, h))
	minmax := num.MinMax(xy)
	spanX := minmax.At(0, 1) - minmax.At(0, 0)
	spanY := minmax.At(1, 1) - minmax.At(1, 0)
	spanZ := minmax.At(2, 1) - minmax.At(2, 0)

	upScaleX := func(row int) float32 {
		return float32(((xy.At(row, 0)-minmax.At(0, 0))*200.0)/spanX - 100.0)
	}

	upScaleY := func(row int) float32 {
		return float32(((xy.At(row, 1)-minmax.At(1, 0))*200.0)/spanY - 100.0)
	}

	upScaleZ := func(row int) float32 {
		return float32(((xy.At(row, 2)-minmax.At(2, 0))*200.0)/spanZ - 100.0)
	}

	row := 0
	nodes := []opts.Chart3DData{}
	h.ForAll(level, func(rank int, vector kv.VF32, vertex []kv.VF32) error {
		nodes = append(nodes, opts.Chart3DData{
			Name: stringify(vector),
			Value: []any{
				upScaleX(row),
				upScaleY(row),
				upScaleZ(row),
			},
		})

		row++
		return nil
	})

	return nodes
}

//------------------------------------------------------------------------------

func NewGraph(nodes []opts.GraphNode, links []opts.GraphLink) *components.Page {
	graph := charts.NewGraph()

	graph.AddSeries("graph", nodes, links).
		SetGlobalOptions(
			charts.WithInitializationOpts(opts.Initialization{
				Width:  "100%",
				Height: "100%",
			}),
		).
		SetSeriesOptions(
			charts.WithGraphChartOpts(opts.GraphChart{
				Layout:             "force",
				Roam:               opts.Bool(true),
				FocusNodeAdjacency: opts.Bool(true),
				Force: &opts.GraphForce{
					Repulsion: 100.0,
				},
				Categories: Categories(),
			}),
			charts.WithEmphasisOpts(opts.Emphasis{
				Label: &opts.Label{
					Show:     opts.Bool(true),
					Color:    "black",
					Position: "left",
				},
			}),
			charts.WithLineStyleOpts(opts.LineStyle{
				Curveness: 0.3,
				Color:     "source",
			}),
		)

	page := components.NewPage()
	page.SetLayout(components.PageFullLayout)
	page.AddCharts(graph)

	return page
}

//------------------------------------------------------------------------------

func NewScatter3D(points []opts.Chart3DData) *components.Page {

	graph := charts.NewScatter3D()
	graph.AddSeries("graph", points)
	graph.SetGlobalOptions(
		charts.WithInitializationOpts(opts.Initialization{
			Width:  "100%",
			Height: "100%",
		}),

		charts.WithVisualMapOpts(
			opts.VisualMap{
				Calculable: opts.Bool(true),
				Min:        -100,
				Max:        100,
				InRange:    &opts.VisualMapInRange{Color: schemaScatter3D},
			},
		),
	)

	page := components.NewPage()
	page.SetLayout(components.PageFullLayout)
	page.AddCharts(graph)

	return page
}

//------------------------------------------------------------------------------

// func labelOf(atoms *atom.Pool, vector kv.VF32) string {
// 	label := atoms.String(vector.Key)

// 	if label == "" {
// 		return "ø"
// 	}

// 	return label
// }

func colorOf(distance float32) string {
	// 0.5 distance corresponds to orthogonal vectors
	if distance > 0.5 {
		return "#DCE0E5"
	}

	norm := distance * float32(len(schemaColor)-2) / 0.5
	return schemaColor[int(math.Trunc(float64(norm)))]
}

func categoryOf(distance float32) int {
	if distance > 0.5 {
		return len(schemaLabel) - 1
	}

	norm := distance * float32(len(schemaColor)-2) / 0.5
	return int(math.Trunc(float64(norm)))
}
