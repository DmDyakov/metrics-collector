package file

import (
	"testing"

	models "metrics-collector/internal/server/model"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestFileStorage_SaveAndGetAll(t *testing.T) {
	t.Run("save and retrieve metrics", func(t *testing.T) {
		tmpFile := t.TempDir() + "/test_metrics.json"
		fs := NewFileStorage(tmpFile)

		v := 42.5
		err := fs.SaveMetric(models.Metrics{ID: "cpu", MType: "gauge", Value: &v})
		require.NoError(t, err)

		metrics, err := fs.GetAll()
		require.NoError(t, err)
		assert.Len(t, metrics, 1)
		assert.Equal(t, "cpu", metrics[0].ID)
	})

	t.Run("save batch", func(t *testing.T) {
		tmpFile := t.TempDir() + "/test_batch.json"
		fs := NewFileStorage(tmpFile)

		v1 := 1.0
		v2 := 2.0
		count, err := fs.SaveBatch([]models.Metrics{
			{ID: "m1", MType: "gauge", Value: &v1},
			{ID: "m2", MType: "gauge", Value: &v2},
		})
		require.NoError(t, err)
		assert.Equal(t, 2, *count)
	})
}
