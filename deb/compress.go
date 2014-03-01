package deb

import (
	"archive/tar"
	"compress/bzip2"
	"compress/gzip"
	"io"
)

func NewTarBz2Reader(r io.Reader) *tar.Reader {
	br := bzip2.NewReader(r)
	return tar.NewReader(br)
}

func NewTarGzReader(r io.Reader) (*tar.Reader, error) {
	gr, err := gzip.NewReader(r)
	if err != nil {
		return nil, err
	}

	return tar.NewReader(gr), nil
}
