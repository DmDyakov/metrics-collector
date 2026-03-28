package compress

import (
	"bytes"
	"compress/gzip"
	"errors"
	"fmt"
)

var (
	ErrCompress   = errors.New("compression failed")
	ErrDecompress = errors.New("decompression failed")
)

type Gzip struct{}

func NewGzip() *Gzip {
	return &Gzip{}
}

func (gz *Gzip) Compress(data []byte) ([]byte, error) {
	var buf bytes.Buffer

	w, err := gzip.NewWriterLevel(&buf, gzip.BestSpeed)
	if err != nil {
		return nil, fmt.Errorf("%w: %w", ErrCompress, err)
	}

	_, err = w.Write(data)
	if err != nil {
		w.Close()
		return nil, fmt.Errorf("%w: %w", ErrCompress, err)
	}

	err = w.Close()
	if err != nil {
		return nil, fmt.Errorf("%w: %w", ErrCompress, err)
	}

	return buf.Bytes(), nil
}

func (gz *Gzip) Decompress(data []byte) ([]byte, error) {
	r, err := gzip.NewReader(bytes.NewReader(data))
	if err != nil {
		return nil, fmt.Errorf("%w: %w", ErrDecompress, err)
	}
	defer r.Close()

	var b bytes.Buffer
	_, err = b.ReadFrom(r)
	if err != nil {
		return nil, fmt.Errorf("%w: %w", ErrDecompress, err)
	}

	return b.Bytes(), nil
}
