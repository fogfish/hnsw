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

	"github.com/akrylysov/pogreb"
	"github.com/fogfish/hnsw"
	"github.com/fogfish/hnsw/cmd/hnsw/glove"
	"github.com/fogfish/hnsw/cmd/hnsw/viz"
	kv "github.com/fogfish/hnsw/vector"
	"github.com/kshard/atom"
	"github.com/spf13/cobra"
)

func init() {
	rootCmd.AddCommand(gloveCmd)

	gloveCmd.AddCommand(gloveCreateCmd)
	gloveCreateCmd.Flags().StringVarP(&gloveDataset, "glove", "g", "", "path to GLoVe dataset")
	gloveCreateCmd.Flags().IntVarP(&hnswEfConn, "efficiency-factor", "e", 200, "Construction Efficiency Factor")
	gloveCreateCmd.Flags().IntVarP(&hnswM, "max-connections", "m", 16, "Maximum Connections")
	gloveCreateCmd.Flags().IntVarP(&hnswM0, "max-connections-0", "0", 32, "Maximum Connections at Layer 0")

	gloveCmd.AddCommand(gloveQueryCmd)
	gloveQueryCmd.Flags().StringVarP(&gloveDataset, "glove", "g", "", "path to GLoVe dataset")

	gloveCmd.AddCommand(gloveDrawCmd)
	gloveDrawCmd.Flags().StringVar(&gloveDrawHTML, "html", "hnsw.html", "visualized dataset")
	gloveDrawCmd.Flags().IntVarP(&gloveDrawLevel, "level", "l", 0, "level to visualize")
	gloveDrawCmd.Flags().BoolVar(&gloveDraw3D, "3d", false, "draw 3D visualization")
}

var gloveCmd = &cobra.Command{
	Use:   "glove",
	Short: "evaluate HNSW using GLoVe-like datasets",
	Long: `
The command-line utility support the creation and optimization of HNSW indices,
using GloVe-like datasets. You can "create" an HNSW index tailored to your dataset,
evaluate the index quality by "query" the same indexed words, and "draw"
the graph layers to gain insights into the structure and connectivity of your data.

Obtain dataset copy for evaluation from https://nlp.stanford.edu/projects/glove/,
or use any textual representation following the format:

	word -0.37604 0.24116 ... -0.26098 -0.0079604
`,
	SilenceUsage: true,
	Run:          root,
}

var (
	gloveDataset   string
	gloveDrawHTML  string
	gloveDrawLevel int
	gloveDraw3D    bool
)

//------------------------------------------------------------------------------

var gloveCreateCmd = &cobra.Command{
	Use:   "create",
	Short: "create HNSW index for using GLoVe-like datasets.",
	Long: `
Create persistent HNSW index using GLoVe-like dataset.

Obtain dataset copy for evaluation from https://nlp.stanford.edu/projects/glove/,
or use any textual representation following the format:

	word -0.37604 0.24116 ... -0.26098 -0.0079604
`,
	SilenceUsage: true,
	RunE:         gloveCreate,
}

func gloveCreate(cmd *cobra.Command, args []string) (err error) {
	if gloveDataset == "" {
		return errors.New("undefined GLoVe dataset")
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

	atoms := atom.New(atom.NewPermanentMap(db))

	h := glove.New(hnswM, hnswM0, hnswEfConn, sysSeed)
	if err := glove.Create(atoms, h, sysThreads, gloveDataset); err != nil {
		return err
	}

	if err := h.Write(db); err != nil {
		return err
	}

	os.Stderr.WriteString(fmt.Sprintf("==> created %s\n", hnswDataset))
	os.Stderr.WriteString(fmt.Sprintf("   %s\n", h))

	return nil
}

//------------------------------------------------------------------------------

var gloveQueryCmd = &cobra.Command{
	Use:   "query",
	Short: "evaluate the index quality by 'query' the same indexed words",
	Long: `
Evaluate the index quality by "query" the same indexed words. It considers
query successful if result include it.
`,
	SilenceUsage: true,
	RunE:         gloveQuery,
}

func gloveQuery(cmd *cobra.Command, args []string) (err error) {
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

	atoms := atom.New(atom.NewPermanentMap(db))

	h := glove.New(hnswM, hnswM0, hnswEfConn, sysSeed)
	if err := h.Read(db); err != nil {
		return err
	}

	os.Stderr.WriteString(fmt.Sprintf("==> reading %s\n", hnswDataset))
	os.Stderr.WriteString(fmt.Sprintf("   %s\n", h))

	if err := glove.Query(atoms, h, gloveDataset); err != nil {
		return err
	}

	return nil
}

//------------------------------------------------------------------------------

var gloveDrawCmd = &cobra.Command{
	Use:   "draw",
	Short: "visualize HNSW index.",
	Long: `
Visualize existing HNSW index using GloVe-like dataset.
`,
	SilenceUsage: true,
	RunE:         gloveDraw,
}

func gloveDraw(cmd *cobra.Command, args []string) (err error) {
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

	atoms := atom.New(atom.NewPermanentMap(db))

	h := glove.New(hnswM, hnswM0, hnswEfConn, sysSeed)
	if err := h.Read(db); err != nil {
		return err
	}

	os.Stderr.WriteString(fmt.Sprintf("==> drawing %s\n", hnswDataset))
	os.Stderr.WriteString(fmt.Sprintf("   %s\n", h))

	if gloveDrawLevel > h.Level() {
		return fmt.Errorf("level must be in range of [0, %d]", h.Level())
	}

	if gloveDraw3D {
		if err := drawGloveLevel3D(atoms, h, gloveDrawLevel); err != nil {
			return err
		}
	} else {
		if err := drawGloveLevel(atoms, h, gloveDrawLevel); err != nil {
			return err
		}
	}

	return nil
}

func drawGloveLevel(atoms *atom.Pool, h *hnsw.HNSW[kv.VF32], level int) error {
	upScaleD := func(d float32) float32 { return d }
	vec2text := func(v kv.VF32) string { return atoms.String(v.Key) }

	nodes, links := viz.Visualize(level, h, upScaleD, vec2text)
	if len(nodes) == 0 || len(links) == 0 {
		return fmt.Errorf("level is empty")
	}

	f, err := os.Create(gloveDrawHTML)
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

func drawGloveLevel3D(atoms *atom.Pool, h *hnsw.HNSW[kv.VF32], level int) error {
	vec2text := func(v kv.VF32) string { return atoms.String(v.Key) }
	nodes := viz.Visualize3D(level, h, vec2text)

	if len(nodes) == 0 {
		return fmt.Errorf("level is empty")
	}

	f, err := os.Create(gloveDrawHTML)
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
