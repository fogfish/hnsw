//
// Copyright (C) 2024 Dmitry Kolesnikov
//
// This file may be modified and distributed under the terms
// of the MIT license.  See the LICENSE file for details.
// https://github.com/fogfish/hnsw
//

package opt

import (
	"fmt"
	"io"
	"os"
	"strconv"

	"github.com/fogfish/hnsw"
	"github.com/fogfish/hnsw/cmd/try"
	"github.com/go-echarts/go-echarts/v2/charts"
	"github.com/go-echarts/go-echarts/v2/components"
	"github.com/go-echarts/go-echarts/v2/opts"
	"github.com/spf13/cobra"
	"github.com/willf/bitset"
)

func init() {
	rootCmd.AddCommand(drawCmd)
	drawCmd.Flags().StringVarP(&drawDataset, "dataset", "d", "siftsmall", "name of the dataset from http://corpus-texmex.irisa.fr")
	drawCmd.Flags().StringVarP(&drawOutput, "output", "o", ".", "directory to output rendered layers")
}

var (
	drawDataset string
	drawOutput  string
)

var drawCmd = &cobra.Command{
	Use:   "draw",
	Short: "draw the graph",
	Long: `
'hnsw graw' visualize graph the graph(s) build from datasets for approximate
nearest neighbor search available at http://corpus-texmex.irisa.fr.

It is required to obtain the dataset(s) into local environment:

	curl ftp://ftp.irisa.fr/local/texmex/corpus/siftsmall.tar.gz -o siftsmall.tar.gz

`,
	SilenceUsage: true,
	RunE:         draw,
}

func draw(cmd *cobra.Command, args []string) error {
	h := try.New()

	if err := try.Create(h, drawDataset); err != nil {
		return err
	}

	fmt.Printf("\n==> drawing %s\n", drawDataset)
	for level := h.Level(); level >= 0; level-- {
		fmt.Printf("==> draw level %3d\n", level)
		if err := drawLevel(h, level); err != nil {
			return err
		}
	}

	return nil
}

func drawLevel(h *hnsw.HNSW[try.Node], level int) error {
	nodes, links, kinds := cutLevel(h, level)
	if len(nodes) == 0 || len(links) == 0 {
		return nil
	}

	graph := charts.NewGraph()

	graph.AddSeries("graph", nodes, links).
		SetSeriesOptions(
			charts.WithGraphChartOpts(opts.GraphChart{
				Layout:             "force",
				Draggable:          true,
				Roam:               true,
				FocusNodeAdjacency: true,
				Force: &opts.GraphForce{
					Repulsion:  200.0, //800.0,
					Gravity:    0.05,  //0.01,
					EdgeLength: 60.0,
				},

				Categories: kinds,
			}),
			charts.WithEmphasisOpts(opts.Emphasis{
				Label: &opts.Label{
					Show:     true,
					Color:    "black",
					Position: "left",
				},
			}),
			charts.WithLineStyleOpts(opts.LineStyle{
				Curveness: 0.3,
			}),
		)

	page := components.NewPage()
	page.Height = "100%"
	page.Width = "100%"
	page.AddCharts(
		graph,
	)

	f, err := os.Create(fmt.Sprintf("%s/graph-%s-L%d.html", drawOutput, drawDataset, level))
	if err != nil {
		return err
	}
	return page.Render(io.MultiWriter(f))
}

func cutLevel(h *hnsw.HNSW[try.Node], level int) ([]opts.GraphNode, []opts.GraphLink, []*opts.GraphCategory) {
	var visited bitset.BitSet
	mrank := level
	nodes := []opts.GraphNode{}
	links := []opts.GraphLink{}
	kinds := []*opts.GraphCategory{}

	h.FMap(level, func(rank int, vector try.Node, vertex []try.Node) error {
		if visited.Test(uint(vector.ID)) {
			return nil
		}
		visited.Set(uint(vector.ID))

		if rank > mrank {
			mrank = rank
		}

		nodes = append(nodes,
			opts.GraphNode{
				Name:     strconv.Itoa(vector.ID),
				Category: rank - level - 1,
			},
		)

		for _, v := range vertex {
			links = append(links,
				opts.GraphLink{
					Source: strconv.Itoa(vector.ID),
					Target: strconv.Itoa(v.ID),
				},
			)
		}

		return nil
	})

	for i := level; i <= mrank; i++ {
		kinds = append(kinds,
			&opts.GraphCategory{
				Name: strconv.Itoa(i),
			},
		)
	}

	return nodes, links, kinds
}
