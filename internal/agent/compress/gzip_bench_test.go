package compress

import (
	"testing"
)

func BenchmarkCompress(b *testing.B) {
	gz := NewGzip()
	data := []byte(`{"id":"test","type":"gauge","value":42.5}`)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = gz.Compress(data)
	}
}

func BenchmarkDecompress(b *testing.B) {
	gz := NewGzip()
	data := []byte(`{"id":"test","type":"gauge","value":42.5}`)
	compressed, _ := gz.Compress(data)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = gz.Decompress(compressed)
	}
}
