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
	"io"
	"os"

	"github.com/kshard/fvecs"
	"github.com/spf13/cobra"
)

func init() {
	rootCmd.AddCommand(queryCmd)
	queryCmd.Flags().StringVarP(&queryVectors, "dataset", "d", "", ".fvecs")
	queryCmd.Flags().StringVarP(&queryText, "text", "t", "", ".bvecs")
	queryCmd.Flags().IntVarP(&queryVecSize, "vector", "v", 128, "vector size")
	queryCmd.Flags().StringVarP(&queryQuery, "query", "q", "", ".fvecs")

	// drawCmd.Flags().StringVarP(&drawOutput, "output", "o", ".", "directory to output rendered layers")
}

var (
	queryVectors string
	queryText    string
	queryVecSize int
	queryQuery   string

	// drawOutput  string
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
	// h := try.New(queryVecSize)
	// if err := try.Create(h, queryVectors); err != nil {
	// 	return err
	// }

	// text, err := readText()
	// if err != nil {
	// 	return err
	// }

	// fv, err := os.Open(queryQuery)
	// if err != nil {
	// 	return err
	// }
	// defer fv.Close()

	// //
	// t := time.Now()
	// c := 1
	// fr := fvecs.NewDecoder[float32](fv)
	// for {
	// 	q, err := fr.Read()
	// 	switch {
	// 	case err == nil:
	// 		os.Stdout.WriteString("\n---\n")

	// 		result := h.Search(try.Node{Vector: q}, 10, 100)
	// 		for _, v := range result {
	// 			d := vector.Cosine.Distance(q, v.Vector)
	// 			os.Stdout.WriteString(
	// 				fmt.Sprintf("%f\n%s\n", d, text[v.ID]),
	// 			)
	// 		}

	// 	case errors.Is(err, io.EOF):
	// 		os.Stderr.WriteString(
	// 			fmt.Sprintf("==> query %9d vectors in %s (%d ns/op)\n", c, time.Since(t), int(time.Since(t).Nanoseconds())/c),
	// 		)
	// 		return nil
	// 	default:
	// 		return err
	// 	}

	// 	c++

	// 	if c%1000 == 0 {
	// 		os.Stderr.WriteString(
	// 			fmt.Sprintf("==> query %9d vectors in %s (%d ns/op)\n", c, time.Since(t), int(time.Since(t).Nanoseconds())/c),
	// 		)
	// 	}
	// }
	return nil
}

func readText() (map[int]string, error) {
	bv, err := os.Open(queryText)
	if err != nil {
		return nil, err
	}
	defer bv.Close()

	id := 1
	text := map[int]string{}
	br := fvecs.NewDecoder[byte](bv)

	for {
		t, err := br.Read()

		switch {
		case err == nil:
			text[id] = string(t)
		case errors.Is(err, io.EOF):
			return text, nil
		default:
			return nil, err
		}

		id++
	}
}
