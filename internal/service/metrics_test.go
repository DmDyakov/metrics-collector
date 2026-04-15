package service

import (
	"context"
	models "metrics-collector/internal/model"
	"metrics-collector/internal/service/mocks"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"
)

func TestService_UpdateMetricByJSON(t *testing.T) {
	ctx := context.Background()

	t.Run("positive: create new counter metric", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		mockRepo := mocks.NewMockRepository(ctrl)
		svc := NewMetricsService(mockRepo)

		inputDelta := int64(10)
		input := models.Metrics{
			ID:    "TestCount",
			MType: models.Counter,
			Delta: &inputDelta,
		}

		mockRepo.EXPECT().
			GetMetric("TestCount").
			Return(nil, false)

		mockRepo.EXPECT().
			UpdateMetric(gomock.Any(), input).
			Return(&input, nil)

		result, err := svc.UpdateMetricByJSON(ctx, input)

		require.NoError(t, err)
		require.NotNil(t, result)
		assert.Equal(t, *input.Delta, *result.Delta)
		assert.Equal(t, input.MType, result.MType)
		assert.Equal(t, input.ID, result.ID)
		assert.Nil(t, result.Value)
	})

	t.Run("positive: increment existing counter metric", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		mockRepo := mocks.NewMockRepository(ctrl)
		svc := NewMetricsService(mockRepo)

		inputDelta := int64(10)
		input := models.Metrics{
			ID:    "TestCount",
			MType: models.Counter,
			Delta: &inputDelta,
		}

		outputDelta := int64(3)
		existing := models.Metrics{
			ID:    "TestCount",
			MType: models.Counter,
			Delta: &outputDelta,
		}

		sumDelta := *input.Delta + *existing.Delta
		processed := models.Metrics{
			ID:    "TestCount",
			MType: models.Counter,
			Delta: &sumDelta,
		}

		mockRepo.EXPECT().
			GetMetric("TestCount").
			Return(&existing, true)

		mockRepo.EXPECT().
			UpdateMetric(gomock.Any(), processed).
			Return(&processed, nil)

		result, err := svc.UpdateMetricByJSON(ctx, input)
		require.NoError(t, err)
		require.NotNil(t, result)
		assert.Equal(t, *result.Delta, *input.Delta+*existing.Delta)
		assert.Equal(t, input.MType, result.MType)
		assert.Equal(t, input.ID, result.ID)
		assert.Nil(t, result.Value)
	})

	t.Run("positive: create or update new gauge metric", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		mockRepo := mocks.NewMockRepository(ctrl)
		svc := NewMetricsService(mockRepo)

		inputValue := float64(2.5)
		input := models.Metrics{
			ID:    "TestGauge",
			MType: models.Gauge,
			Value: &inputValue,
		}

		mockRepo.EXPECT().
			UpdateMetric(gomock.Any(), input).
			Return(&input, nil)

		result, err := svc.UpdateMetricByJSON(ctx, input)
		require.NoError(t, err)
		require.NotNil(t, result)
		assert.Equal(t, input.MType, result.MType)
		assert.Equal(t, input.ID, result.ID)
		assert.Equal(t, *result.Value, *input.Value)
		assert.Nil(t, result.Delta)
	})
}
