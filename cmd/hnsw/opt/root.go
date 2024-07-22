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
	"os"
	"strings"

	"github.com/spf13/cobra"
)

// Execute is entry point for cobra cli application
func Execute() {
	if err := rootCmd.Execute(); err != nil {
		e := err.Error()
		fmt.Println(strings.ToUpper(e[:1]) + e[1:])
		os.Exit(1)
	}
}

func init() {
	rootCmd.PersistentFlags().IntVarP(&sysThreads, "threads", "t", 1, "number of threads")
	rootCmd.PersistentFlags().Int64Var(&sysSeed, "seed", 0, "seed for random generator (default current timestamp)")
	rootCmd.PersistentFlags().StringVarP(&hnswDataset, "index", "i", "", "path to HNSW persistent index")
}

var rootCmd = &cobra.Command{
	Use:   "hnsw",
	Short: "Command-line utility for  for hyper-optimizing the parameters of your HNSW graphs.",
	Long: `
The command-line utility designed for hyper-optimizing the parameters of your
HNSW graphs. This tool allows you to efficiently explore various configurations,
such as M, M0, efConstruction, and others, to find the optimal settings for
your specific dataset and use case.

The command line utility is tailored for working with
* GLoVe https://nlp.stanford.edu/projects/glove/
* SIFT http://corpus-texmex.irisa.fr
* your own dataset in the textual (GLoVe format).
	`,
	Run: root,
}

func root(cmd *cobra.Command, args []string) {
	cmd.Help()
}

var (
	hnswEfConn  int
	hnswM       int
	hnswM0      int
	hnswDataset string
	sysSeed     int64
	sysThreads  int
)
