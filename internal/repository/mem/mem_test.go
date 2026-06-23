package mem

import (
	models "metrics-collector/internal/model"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestMemStorage_UpdateMetric(t *testing.T) {
	ms := NewMemStorage()

	testValue := 42.5
	testMetric := models.Metrics{
		ID:    "test_name",
		MType: "test_type",
		Value: &testValue,
	}

	ms.SaveMetric(testMetric)

	assert.Len(t, ms.metrics, 1)
	assert.Equal(t, testMetric, ms.metrics["test_name"])
}

func TestMemStorage_GetMetricByName(t *testing.T) {
	t.Run("found", func(t *testing.T) {
		ms := NewMemStorage()
		v := 42.5
		ms.SaveMetric(models.Metrics{ID: "test", MType: "gauge", Value: &v})

		m, ok := ms.GetMetricByName("test")
		assert.True(t, ok)
		assert.Equal(t, 42.5, *m.Value)
	})

	t.Run("not found", func(t *testing.T) {
		ms := NewMemStorage()
		_, ok := ms.GetMetricByName("missing")
		assert.False(t, ok)
	})
}

func TestMemStorage_SaveBatch(t *testing.T) {
	t.Run("saves multiple metrics", func(t *testing.T) {
		ms := NewMemStorage()
		v1 := 1.0
		v2 := 2.0
		batch := []models.Metrics{
			{ID: "m1", MType: "gauge", Value: &v1},
			{ID: "m2", MType: "gauge", Value: &v2},
		}

		count := ms.SaveBatch(batch)
		assert.Equal(t, 2, *count)
		assert.Len(t, ms.GetAll(), 2)
	})
}

func TestMemStorage_GetAll(t *testing.T) {
	t.Run("empty store returns empty map", func(t *testing.T) {
		ms := NewMemStorage()
		result := ms.GetAll()
		assert.Empty(t, result)
	})
}
