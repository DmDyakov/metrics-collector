package service

import (
	"context"
	"errors"
	"metrics-collector/internal/domain/metrics"
	"metrics-collector/internal/server/errs"
	models "metrics-collector/internal/server/model"
	"metrics-collector/internal/server/service/mocks"
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
			MType: metrics.Counter,
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
			MType: metrics.Counter,
			Delta: &inputDelta,
		}

		outputDelta := int64(3)
		existing := models.Metrics{
			ID:    "TestCount",
			MType: metrics.Counter,
			Delta: &outputDelta,
		}

		sumDelta := *input.Delta + *existing.Delta
		processed := models.Metrics{
			ID:    "TestCount",
			MType: metrics.Counter,
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
			MType: metrics.Gauge,
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
	t.Run("negative: empty metric ID", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		mockRepo := mocks.NewMockMetricsRepository(ctrl)
		svc := NewMetricsService(mockRepo)

		input := models.Metrics{ID: "", MType: metrics.Gauge}

		_, err := svc.UpdateMetric(ctx, input)
		require.Error(t, err)
		assert.ErrorIs(t, err, errs.ErrInvalidRequest)
	})

	t.Run("negative: unknown metric type", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		mockRepo := mocks.NewMockMetricsRepository(ctrl)
		svc := NewMetricsService(mockRepo)

		input := models.Metrics{ID: "test", MType: "invalid"}

		_, err := svc.UpdateMetric(ctx, input)
		require.Error(t, err)
		assert.ErrorIs(t, err, errs.ErrInvalidRequest)
		assert.ErrorIs(t, err, errs.ErrUnknownMetricType)
	})

	t.Run("negative: gauge without value", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		mockRepo := mocks.NewMockMetricsRepository(ctrl)
		svc := NewMetricsService(mockRepo)

		input := models.Metrics{ID: "test", MType: metrics.Gauge, Value: nil}

		_, err := svc.UpdateMetric(ctx, input)
		require.Error(t, err)
		assert.ErrorIs(t, err, errs.ErrInvalidRequest)
		assert.ErrorIs(t, err, errs.ErrMetricValueForGaugeRequired)
	})

	t.Run("negative: counter without delta", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		mockRepo := mocks.NewMockMetricsRepository(ctrl)
		svc := NewMetricsService(mockRepo)

		input := models.Metrics{ID: "test", MType: metrics.Counter, Delta: nil}

		_, err := svc.UpdateMetric(ctx, input)
		require.Error(t, err)
		assert.ErrorIs(t, err, errs.ErrInvalidRequest)
		assert.ErrorIs(t, err, errs.ErrMetricDeltaForCountRequired)
	})

	t.Run("negative: save metric fails", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		mockRepo := mocks.NewMockMetricsRepository(ctrl)
		svc := NewMetricsService(mockRepo)

		delta := int64(10)
		input := models.Metrics{ID: "test", MType: metrics.Counter, Delta: &delta}

		mockRepo.EXPECT().
			GetMetric("test").
			Return(nil, false)

		mockRepo.EXPECT().
			SaveMetric(gomock.Any(), input).
			Return(nil, errors.New("db error"))

		_, err := svc.UpdateMetric(ctx, input)
		require.Error(t, err)
		assert.ErrorIs(t, err, errs.ErrInvalidResponse)
	})
}

func TestService_GetMetric(t *testing.T) {
	t.Run("positive: get gauge metric", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		mockRepo := mocks.NewMockMetricsRepository(ctrl)
		svc := NewMetricsService(mockRepo)

		value := 42.5
		existing := &models.Metrics{ID: "test", MType: metrics.Gauge, Value: &value}

		mockRepo.EXPECT().GetMetric("test").Return(existing, true)

		input := models.Metrics{ID: "test", MType: metrics.Gauge}
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

		input := models.Metrics{ID: "missing", MType: metrics.Gauge}
		_, err := svc.GetMetric(input)

		var notFound *errs.MetricNotFoundError
		require.ErrorAs(t, err, &notFound)
	})

	t.Run("negative: empty metric name", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		mockRepo := mocks.NewMockMetricsRepository(ctrl)
		svc := NewMetricsService(mockRepo)

		_, err := svc.GetMetric(models.Metrics{ID: "", MType: metrics.Gauge})
		require.Error(t, err)
		assert.ErrorIs(t, err, errs.ErrInvalidRequest)
	})

	t.Run("negative: type mismatch", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		mockRepo := mocks.NewMockMetricsRepository(ctrl)
		svc := NewMetricsService(mockRepo)

		value := float64(42.5)
		existing := &models.Metrics{ID: "test", MType: metrics.Gauge, Value: &value}

		mockRepo.EXPECT().GetMetric("test").Return(existing, true)

		input := models.Metrics{ID: "test", MType: metrics.Counter}
		_, err := svc.GetMetric(input)
		require.Error(t, err)
		assert.ErrorIs(t, err, errs.ErrInvalidResponse)
		assert.ErrorIs(t, err, errs.ErrMetricTypeMismatch)
	})

	t.Run("negative: invalid metric type in request", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		mockRepo := mocks.NewMockMetricsRepository(ctrl)
		svc := NewMetricsService(mockRepo)

		_, err := svc.GetMetric(models.Metrics{ID: "test", MType: "invalid"})
		require.Error(t, err)
		assert.ErrorIs(t, err, errs.ErrInvalidRequest)
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
			"m1": {ID: "m1", MType: metrics.Gauge, Value: &v1},
			"m2": {ID: "m2", MType: metrics.Gauge, Value: &v2},
		})

		result, err := svc.GetAllMetrics()
		require.NoError(t, err)
		assert.Len(t, result, 2)
	})
}

func TestService_UpdateMetrics(t *testing.T) {
	ctx := context.Background()

	t.Run("positive: batch update with duplicates", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		mockRepo := mocks.NewMockMetricsRepository(ctrl)
		svc := NewMetricsService(mockRepo)

		delta1 := int64(10)
		delta2 := int64(20)
		batch := []models.Metrics{
			{ID: "c1", MType: metrics.Counter, Delta: &delta1},
			{ID: "c1", MType: metrics.Counter, Delta: &delta2},
		}

		mockRepo.EXPECT().
			GetMetric("c1").
			Return(nil, false)

		count := 1
		mockRepo.EXPECT().
			SaveMetricsBatch(ctx, gomock.Any()).
			DoAndReturn(func(_ context.Context, metrics []models.Metrics) (*int, error) {
				require.Len(t, metrics, 1)
				assert.Equal(t, int64(30), *metrics[0].Delta)
				return &count, nil
			})

		result, err := svc.UpdateMetrics(ctx, batch)
		require.NoError(t, err)
		assert.Equal(t, 1, *result)
	})

	t.Run("negative: invalid metric in batch", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		mockRepo := mocks.NewMockMetricsRepository(ctrl)
		svc := NewMetricsService(mockRepo)

		delta := int64(10)
		batch := []models.Metrics{
			{ID: "ok", MType: metrics.Counter, Delta: &delta},
			{ID: "bad", MType: "invalid", Delta: &delta},
		}

		_, err := svc.UpdateMetrics(ctx, batch)
		require.Error(t, err)
		assert.ErrorIs(t, err, errs.ErrInvalidRequest)
	})
}
