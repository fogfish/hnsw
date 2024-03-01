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

	"github.com/fogfish/hnsw/cmd/try"
	"github.com/fogfish/hnsw/vector"
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
	h := try.New(createVecSize)
	if err := try.Insert(h, runtime.NumCPU(), createDataset); err != nil {
		return err
	}

	fmt.Printf("==> writing %s\n", createOutput)
	if err := vector.Write(h, createOutput); err != nil {
		return err
	}

	return nil
}
