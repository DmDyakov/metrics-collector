package repository

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

	ms.UpdateMetricByArgs(testMetric)

	assert.Len(t, ms.metrics, 1)
	assert.Equal(t, testMetric, ms.metrics["test_name"])
}
