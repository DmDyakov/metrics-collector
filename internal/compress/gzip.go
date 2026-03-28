package compress

import (
	"bytes"
	"compress/gzip"
	"fmt"
)

type Gzip struct{}

func NewGzip() *Gzip {
	return &Gzip{}
}

func (gz *Gzip) Compress(data []byte) ([]byte, error) {
	var buf bytes.Buffer

	w, err := gzip.NewWriterLevel(&buf, gzip.BestSpeed)
	if err != nil {
		return nil, fmt.Errorf("gzip new writer level: %w", err)
	}

	_, err = w.Write(data)
	if err != nil {
		w.Close()
		return nil, fmt.Errorf("gzip write: %w", err)
	}

	err = w.Close()
	if err != nil {
		return nil, fmt.Errorf("gzip close: %w", err)
	}

	return buf.Bytes(), nil
}

func (gz *Gzip) Decompress(data []byte) ([]byte, error) {
	r, err := gzip.NewReader(bytes.NewReader(data))
	if err != nil {
		return nil, err
	}
	defer r.Close()

	var b bytes.Buffer
	_, err = b.ReadFrom(r)
	if err != nil {
		return nil, fmt.Errorf("failed decompress data: %v", err)
	}

	return b.Bytes(), nil
}
