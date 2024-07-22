//
// Copyright (C) 2024 Dmitry Kolesnikov
//
// This file may be modified and distributed under the terms
// of the MIT license.  See the LICENSE file for details.
// https://github.com/fogfish/hnsw
//

package opt

import (
	"errors"
	"fmt"
	"io"
	"os"
	"strconv"

	"github.com/akrylysov/pogreb"
	"github.com/fogfish/hnsw"
	"github.com/fogfish/hnsw/cmd/hnsw/num"
	"github.com/fogfish/hnsw/cmd/hnsw/sift"
	"github.com/fogfish/hnsw/cmd/hnsw/viz"
	kv "github.com/fogfish/hnsw/vector"
	"github.com/spf13/cobra"
)

func init() {
	rootCmd.AddCommand(siftCmd)

	siftCmd.AddCommand(siftCreateCmd)
	siftCreateCmd.Flags().StringVarP(&siftDataset, "sift", "s", "", "path to SIFT dataset")
	siftCreateCmd.Flags().IntVarP(&hnswEfConn, "efficiency-factor", "e", 200, "Construction Efficiency Factor")
	siftCreateCmd.Flags().IntVarP(&hnswM, "max-connections", "m", 16, "Maximum Connections")
	siftCreateCmd.Flags().IntVarP(&hnswM0, "max-connections-0", "0", 32, "Maximum Connections at Layer 0")

	siftCmd.AddCommand(siftQueryCmd)
	siftQueryCmd.Flags().StringVarP(&siftDataset, "sift", "s", "", "path to SIFT dataset")

	siftCmd.AddCommand(siftDrawCmd)
	siftDrawCmd.Flags().StringVar(&siftDrawHTML, "html", "hnsw.html", "visualized dataset")
	siftDrawCmd.Flags().IntVarP(&siftDrawLevel, "level", "l", 0, "level to visualize")
	siftDrawCmd.Flags().BoolVar(&siftDraw3D, "3d", false, "draw 3D visualization")
}

var siftCmd = &cobra.Command{
	Use:   "sift",
	Short: "evaluate HNSW algorithm using SIFT dataset.",
	Long: `
Evaluate HNSW algorithm using SIFT dataset. Obtain dataset copy for evaluation
from http://corpus-texmex.irisa.fr.
`,
	SilenceUsage: true,
	Run:          root,
}

var (
	siftDataset   string
	siftDrawHTML  string
	siftDrawLevel int
	siftDraw3D    bool
)

//------------------------------------------------------------------------------

var siftCreateCmd = &cobra.Command{
	Use:   "create",
	Short: "create HNSW index for SIFT dataset.",
	Long: `
Create persistent HNSW index using SIFT dataset. Obtain dataset copy for
evaluation from http://corpus-texmex.irisa.fr.
`,
	SilenceUsage: true,
	RunE:         siftCreate,
}

func siftCreate(cmd *cobra.Command, args []string) (err error) {
	if siftDataset == "" {
		return errors.New("undefined SIFT dataset")
	}

	if hnswDataset == "" {
		return errors.New("undefined output HNSW index")
	}

	db, err := pogreb.Open(hnswDataset, nil)
	if err != nil {
		panic(err)
	}
	defer func() {
		err = db.Close()
	}()

	h := sift.New(hnswM, hnswM0, hnswEfConn, sysSeed)
	if err := sift.Create(h, sysThreads, siftDataset); err != nil {
		return err
	}

	if err := h.Write(db); err != nil {
		return err
	}

	os.Stderr.WriteString(fmt.Sprintf("==> created %s\n", siftDataset))
	os.Stderr.WriteString(fmt.Sprintf("   %s\n", h))

	return nil
}

//------------------------------------------------------------------------------

var siftQueryCmd = &cobra.Command{
	Use:   "query",
	Short: "query HNSW index for SIFT dataset.",
	Long: `
Query existing HNSW index using SIFT dataset. It uses ground truth dataset for
validation.
`,
	SilenceUsage: true,
	RunE:         siftQuery,
}

func siftQuery(cmd *cobra.Command, args []string) (err error) {
	if hnswDataset == "" {
		return errors.New("undefined output HNSW index")
	}

	db, err := pogreb.Open(hnswDataset, nil)
	if err != nil {
		panic(err)
	}
	defer func() {
		err = db.Close()
	}()

	h := sift.New(hnswM, hnswM0, hnswEfConn, sysSeed)
	if err := h.Read(db); err != nil {
		return err
	}

	if err := sift.Query(h, siftDataset); err != nil {
		return err
	}

	return nil
}

//------------------------------------------------------------------------------

var siftDrawCmd = &cobra.Command{
	Use:   "draw",
	Short: "visualize HNSW index for SIFT dataset.",
	Long: `
Visualize existing HNSW index using SIFT dataset.
`,
	SilenceUsage: true,
	RunE:         siftDraw,
}

func siftDraw(cmd *cobra.Command, args []string) (err error) {
	if hnswDataset == "" {
		return errors.New("undefined output HNSW index")
	}

	db, err := pogreb.Open(hnswDataset, nil)
	if err != nil {
		panic(err)
	}
	defer func() {
		err = db.Close()
	}()

	h := sift.New(hnswM, hnswM0, hnswEfConn, sysSeed)
	if err := h.Read(db); err != nil {
		return err
	}

	os.Stderr.WriteString(
		fmt.Sprintf("==> draw %s\n", hnswDataset),
	)

	if gloveDrawLevel > h.Level() {
		return fmt.Errorf("level must be in range of [0, %d]", h.Level())
	}

	if siftDraw3D {
		if err := drawSiftLevel3D(siftDrawLevel, h); err != nil {
			return err
		}
	} else {
		if err := drawSiftLevel(siftDrawLevel, h); err != nil {
			return err
		}
	}

	return nil
}

func drawSiftLevel(level int, h *hnsw.HNSW[kv.VF32]) error {
	min, max := num.MinMaxDistance(level, h)
	upScaleD := func(d float32) float32 { return ((d - min) * 0.6) / (max - min) }
	vec2text := func(v kv.VF32) string { return strconv.Itoa(int(v.Key)) }

	nodes, links := viz.Visualize(level, h, upScaleD, vec2text)
	if len(nodes) == 0 || len(links) == 0 {
		return fmt.Errorf("level is empty")
	}

	f, err := os.Create(siftDrawHTML)
	if err != nil {
		return err
	}
	defer f.Close()

	page := viz.NewGraph(nodes, links)

	if err := page.Render(io.MultiWriter(f)); err != nil {
		return err
	}

	return nil
}

func drawSiftLevel3D(level int, h *hnsw.HNSW[kv.VF32]) error {
	vec2text := func(v kv.VF32) string { return strconv.Itoa(int(v.Key)) }
	nodes := viz.Visualize3D(level, h, vec2text)

	if len(nodes) == 0 {
		return fmt.Errorf("level is empty")
	}

	f, err := os.Create(siftDrawHTML)
	if err != nil {
		return err
	}
	defer f.Close()

	page := viz.NewScatter3D(nodes)

	if err := page.Render(io.MultiWriter(f)); err != nil {
		return err
	}

	return nil
}
