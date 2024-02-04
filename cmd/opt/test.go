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

	"github.com/fogfish/hnsw/cmd/try"
	"github.com/spf13/cobra"
)

func init() {
	rootCmd.AddCommand(testCmd)
	testCmd.Flags().StringVarP(&testDataset, "dataset", "d", "siftsmall", "name of the dataset from http://corpus-texmex.irisa.fr")
}

var (
	testDataset string
)

var testCmd = &cobra.Command{
	Use:   "test",
	Short: "test the algorithm against dataset",
	Long: `
'hnsw graw' tests algorithms against datasets for approximate
nearest neighbor search available at http://corpus-texmex.irisa.fr.

It is required to obtain the dataset(s) into local environment:

	curl ftp://ftp.irisa.fr/local/texmex/corpus/siftsmall.tar.gz -o siftsmall.tar.gz

`,
	SilenceUsage: true,
	RunE:         test,
}

func test(cmd *cobra.Command, args []string) error {
	h := try.New()

	if err := try.Create(h, testDataset); err != nil {
		return err
	}

	fmt.Println()

	return try.Test(h, testDataset)
}
