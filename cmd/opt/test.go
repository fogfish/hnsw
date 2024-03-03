//
// Copyright (C) 2024 Dmitry Kolesnikov
//
// This file may be modified and distributed under the terms
// of the MIT license.  See the LICENSE file for details.
// https://github.com/fogfish/hnsw
//

package opt

import (
	"github.com/fogfish/hnsw/cmd/try"
	"github.com/fogfish/hnsw/vector"
	"github.com/spf13/cobra"
)

func init() {
	rootCmd.AddCommand(testCmd)
	testCmd.Flags().StringVarP(&testDataset, "dataset", "d", "", "name of the dataset from http://corpus-texmex.irisa.fr")
	testCmd.Flags().StringVarP(&testSuite, "suite", "s", "siftsmall", "name of the dataset from http://corpus-texmex.irisa.fr")
}

var (
	testDataset string
	testSuite   string
)

var testCmd = &cobra.Command{
	Use:   "test",
	Short: "test the algorithm against dataset",
	Long: `
'hnsw draw' tests algorithms against datasets for approximate
nearest neighbor search available at http://corpus-texmex.irisa.fr.

It is required to obtain the dataset(s) into local environment:

	curl ftp://ftp.irisa.fr/local/texmex/corpus/siftsmall.tar.gz -o siftsmall.tar.gz

`,
	SilenceUsage: true,
	RunE:         test,
}

func test(cmd *cobra.Command, args []string) error {
	h := try.New(128)

	if err := vector.Read(h, testDataset); err != nil {
		return err
	}

	return try.Test(h, testSuite)
}
