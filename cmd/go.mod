module github.com/fogfish/hnsw/cmd

go 1.21.3

require (
	github.com/fogfish/hnsw v0.0.0-00010101000000-000000000000
	github.com/go-echarts/go-echarts/v2 v2.3.3
	github.com/spf13/cobra v1.8.0
)

require github.com/fogfish/golem/pure v0.10.1 // indirect

require (
	github.com/inconshreveable/mousetrap v1.1.0 // indirect
	github.com/spf13/pflag v1.0.5 // indirect
	github.com/willf/bitset v1.1.11
)

replace github.com/fogfish/hnsw => ../
