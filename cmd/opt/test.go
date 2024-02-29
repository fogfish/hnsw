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
	"github.com/fogfish/hnsw/kv"
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

	// f := fmt.Sprintf("%s/%s_base.fvecs", testDataset, filepath.Base(testDataset))

	// if err := try.Insert(h, 8, f); err != nil {
	// 	return err
	// }

	// if err := kv.Write(h, "test"); err != nil {
	// 	panic(err)
	// }

	if err := kv.Read(h, "test"); err != nil {
		panic(err)
	}

	// h.Dump()

	// w, _ := os.Create("test.ivecs")
	// e := fvecs.NewEncoder[uint32](w)

	// h.Encode(e)

	// h.FMap(3, func(level int, vector try.Node, vertex []try.Node) error {
	// 	fmt.Printf("ID: %d => %d\n", vector.ID, len(vertex))

	// 	return nil
	// })

	// return nil
	return try.Test(h, testDataset)
}
