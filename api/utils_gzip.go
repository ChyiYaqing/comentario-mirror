package main

import (
	"bytes"
	"compress/gzip"
)

func gzipStatic(b []byte) ([]byte, error) {
	var buf bytes.Buffer
	g := gzip.NewWriter(&buf)
	if _, err := g.Write(b); err != nil {
		err = g.Close()
		return []byte{}, err
	}

	if err := g.Close(); err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}
