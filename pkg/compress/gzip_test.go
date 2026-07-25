package compress

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestGzip_CompressDecompress(t *testing.T) {
	t.Run("compress and decompress", func(t *testing.T) {
		gz := NewGzip()
		original := []byte(`{"id":"test","type":"gauge","value":42.5}`)

		compressed, err := gz.Compress(original)
		require.NoError(t, err)
		assert.NotEmpty(t, compressed)

		decompressed, err := gz.Decompress(compressed)
		require.NoError(t, err)
		assert.Equal(t, original, decompressed)
	})
}
