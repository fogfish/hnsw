module github.com/fogfish/hnsw/cmd

go 1.22.0

require (
	github.com/bits-and-blooms/bitset v1.13.0
	github.com/fogfish/hnsw v0.0.0-00010101000000-000000000000
	github.com/go-echarts/go-echarts/v2 v2.3.3
	github.com/kshard/fvecs v0.0.1
	github.com/kshard/vector v0.0.2
	github.com/spf13/cobra v1.8.0
)

require github.com/fogfish/golem/pure v0.10.1 // indirect

require (
	github.com/chewxy/math32 v1.10.1 // indirect
	github.com/inconshreveable/mousetrap v1.1.0 // indirect
	github.com/spf13/pflag v1.0.5 // indirect
	golang.org/x/sys v0.17.0 // indirect
)

replace github.com/fogfish/hnsw => ../
