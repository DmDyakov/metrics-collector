package service

import (
	"context"
	"metrics-collector/internal/errs"
	models "metrics-collector/internal/model"
	"metrics-collector/internal/service/mocks"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"
)

func TestService_UpdateMetric(t *testing.T) {
	ctx := context.Background()

	t.Run("positive: create new counter metric", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		mockRepo := mocks.NewMockMetricsRepository(ctrl)
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
			SaveMetric(gomock.Any(), input).
			Return(&input, nil)

		result, err := svc.UpdateMetric(ctx, input)

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

		mockRepo := mocks.NewMockMetricsRepository(ctrl)
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
			SaveMetric(gomock.Any(), processed).
			Return(&processed, nil)

		result, err := svc.UpdateMetric(ctx, input)
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

		mockRepo := mocks.NewMockMetricsRepository(ctrl)
		svc := NewMetricsService(mockRepo)

		inputValue := float64(2.5)
		input := models.Metrics{
			ID:    "TestGauge",
			MType: models.Gauge,
			Value: &inputValue,
		}

		mockRepo.EXPECT().
			SaveMetric(gomock.Any(), input).
			Return(&input, nil)

		result, err := svc.UpdateMetric(ctx, input)
		require.NoError(t, err)
		require.NotNil(t, result)
		assert.Equal(t, input.MType, result.MType)
		assert.Equal(t, input.ID, result.ID)
		assert.Equal(t, *result.Value, *input.Value)
		assert.Nil(t, result.Delta)
	})
}

func TestService_GetMetric(t *testing.T) {
	t.Run("positive: get gauge metric", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		mockRepo := mocks.NewMockMetricsRepository(ctrl)
		svc := NewMetricsService(mockRepo)

		value := 42.5
		existing := &models.Metrics{ID: "test", MType: models.Gauge, Value: &value}

		mockRepo.EXPECT().GetMetric("test").Return(existing, true)

		input := models.Metrics{ID: "test", MType: models.Gauge}
		result, err := svc.GetMetric(input)

		require.NoError(t, err)
		assert.Equal(t, 42.5, *result.Value)
	})

	t.Run("negative: metric not found", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		mockRepo := mocks.NewMockMetricsRepository(ctrl)
		svc := NewMetricsService(mockRepo)

		mockRepo.EXPECT().GetMetric("missing").Return(nil, false)

		input := models.Metrics{ID: "missing", MType: models.Gauge}
		_, err := svc.GetMetric(input)

		var notFound *errs.MetricNotFoundError
		require.ErrorAs(t, err, &notFound)
	})
}

func TestService_GetAllMetrics(t *testing.T) {
	t.Run("positive: returns all metrics", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		mockRepo := mocks.NewMockMetricsRepository(ctrl)
		svc := NewMetricsService(mockRepo)

		v1 := 1.0
		v2 := 2.0
		mockRepo.EXPECT().GetAllMetrics().Return(map[string]models.Metrics{
			"m1": {ID: "m1", MType: models.Gauge, Value: &v1},
			"m2": {ID: "m2", MType: models.Gauge, Value: &v2},
		})

		result, err := svc.GetAllMetrics()
		require.NoError(t, err)
		assert.Len(t, result, 2)
	})
}
