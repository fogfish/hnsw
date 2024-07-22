//
// Copyright (C) 2024 Dmitry Kolesnikov
//
// This file may be modified and distributed under the terms
// of the MIT license.  See the LICENSE file for details.
// https://github.com/fogfish/hnsw
//

package glove

import (
	"bufio"
	"io"
	"strconv"
	"strings"
)

type Decoder struct {
	scanner *bufio.Scanner
	txt     string
	vec     []float32
	err     error
}

func NewDecoder(r io.Reader) *Decoder {
	return &Decoder{
		scanner: bufio.NewScanner(r),
	}
}

func (dec *Decoder) Scan() bool {
	if !dec.scanner.Scan() {
		return false
	}

	dec.txt, dec.vec, dec.err = dec.parse(dec.scanner.Text())
	return dec.err == nil
}

func (dec *Decoder) parse(line string) (string, []float32, error) {
	seq := strings.Split(line, " ")

	vec := make([]float32, len(seq)-1)
	for i := 1; i < len(seq); i++ {
		v, err := strconv.ParseFloat(seq[i], 32)
		if err != nil {
			return "", nil, err
		}
		vec[i-1] = float32(v)
	}

	return seq[0], vec, nil
}

func (dec *Decoder) Err() error {
	if dec.err != nil {
		return dec.err
	}
	return dec.scanner.Err()
}

func (dec *Decoder) Text() (string, []float32) {
	return dec.txt, dec.vec
}
