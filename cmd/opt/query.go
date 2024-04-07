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
	"time"

	"github.com/akrylysov/pogreb"
	"github.com/fogfish/hnsw/cmd/try"
	kv "github.com/fogfish/hnsw/vector"
	"github.com/kshard/fvecs"
	"github.com/spf13/cobra"
)

func init() {
	rootCmd.AddCommand(queryCmd)
	queryCmd.Flags().StringVarP(&queryDataset, "dataset", "d", "", "path to hnsw index")
	queryCmd.Flags().IntVarP(&queryVecSize, "vector", "v", 128, "vector size")
	queryCmd.Flags().StringVarP(&queryQuery, "query", "q", "", ".fvecs")
	queryCmd.Flags().StringVarP(&queryText, "text", "t", "", ".bvecs")
}

var (
	queryDataset string
	queryVecSize int
	queryQuery   string
	queryText    string
)

var queryCmd = &cobra.Command{
	Use:   "query",
	Short: "query test dataset",
	Long: `
`,
	SilenceUsage: true,
	RunE:         query,
}

func query(cmd *cobra.Command, args []string) error {
	db, err := pogreb.Open(queryDataset, nil)
	if err != nil {
		panic(err)
	}

	h := try.New(queryVecSize)

	if err := h.Read(db); err != nil {
		return err
	}

	text, err := readText()
	if err != nil {
		return err
	}

	fv, err := os.Open(queryQuery)
	if err != nil {
		return err
	}
	defer fv.Close()

	//
	t := time.Now()
	c := 1
	fr := fvecs.NewDecoder[float32](fv)

	for {
		q, err := fr.Read()
		switch {
		case err == nil:
			os.Stdout.WriteString("\n---\n")

			search := kv.VF32{Vector: q}
			result := h.Search(search, 5, 100)
			for _, v := range result {
				d := h.Distance(search, v)
				os.Stdout.WriteString(
					fmt.Sprintf("%f\n%s\n", d, text[v.Key]),
				)
			}

		case errors.Is(err, io.EOF):
			os.Stderr.WriteString(
				fmt.Sprintf("==> query %9d vectors in %s (%d ns/op)\n", c, time.Since(t), int(time.Since(t).Nanoseconds())/c),
			)
			db.Close()
			return nil
		default:
			return err
		}

		c++

		if c%1000 == 0 {
			os.Stderr.WriteString(
				fmt.Sprintf("==> query %9d vectors in %s (%d ns/op)\n", c, time.Since(t), int(time.Since(t).Nanoseconds())/c),
			)
		}
	}
}

func readText() (map[uint32]string, error) {
	bv, err := os.Open(queryText)
	if err != nil {
		return nil, err
	}
	defer bv.Close()

	id := uint32(0)
	text := map[uint32]string{}
	br := fvecs.NewDecoder[byte](bv)

	for {
		t, err := br.Read()

		switch {
		case err == nil:
			id++
			text[id] = string(t)
		case errors.Is(err, io.EOF):
			return text, nil
		default:
			return nil, err
		}

	}
}
