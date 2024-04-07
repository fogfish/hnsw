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
	"runtime"

	"github.com/akrylysov/pogreb"
	"github.com/fogfish/hnsw/cmd/try"
	"github.com/spf13/cobra"
)

func init() {
	rootCmd.AddCommand(createCmd)
	createCmd.Flags().StringVarP(&createDataset, "dataset", "d", "", "path to input *.fvecs files")
	createCmd.Flags().StringVarP(&createOutput, "output", "o", "test", "output")
	createCmd.Flags().IntVarP(&createVecSize, "vector", "v", 128, "vector size")
}

var (
	createDataset string
	createOutput  string
	createVecSize int
)

var createCmd = &cobra.Command{
	Use:   "create",
	Short: "create the dataset",
	Long: `
Creates the dataset from *.fvecs, making it reusable for other tests
`,
	SilenceUsage: true,
	RunE:         create,
}

func create(cmd *cobra.Command, args []string) error {
	db, err := pogreb.Open(createOutput, nil)
	if err != nil {
		panic(err)
	}

	h := try.New(createVecSize)
	if err := try.Insert(h, runtime.NumCPU(), createDataset); err != nil {
		return err
	}

	fmt.Printf("==> writing %s\n", createOutput)
	if err := h.Write(db); err != nil {
		return err
	}

	if err := db.Sync(); err != nil {
		return err
	}

	if err := db.Close(); err != nil {
		return err
	}

	return nil
}
