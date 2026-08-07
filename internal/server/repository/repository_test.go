package repository

import (
	"context"
	"errors"
	"testing"

	models "metrics-collector/internal/server/model"
	"metrics-collector/internal/server/repository/mocks"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"
	"go.uber.org/zap"
)

func TestSaveMetric_MemOnly(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockMem := mocks.NewMockMemStorage(ctrl)
	mockMem.EXPECT().SaveMetric(gomock.Any()).Times(1)

	repo := New(mockMem, nil, 300, zap.NewNop())

	metric := models.Metrics{ID: "test", MType: "gauge", Value: floatPtr(42.5)}
	saved, err := repo.SaveMetric(context.Background(), metric)
	require.NoError(t, err)
	assert.Equal(t, "test", saved.ID)
}

func TestSaveMetric_ImmediatePersistent(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockMem := mocks.NewMockMemStorage(ctrl)
	mockPersistent := mocks.NewMockPersistentStorage(ctrl)

	mockMem.EXPECT().SaveMetric(gomock.Any()).Times(1)
	mockPersistent.EXPECT().SaveMetric(gomock.Any(), gomock.Any()).Return(nil).Times(1)

	repo := New(mockMem, mockPersistent, 0, zap.NewNop())

	metric := models.Metrics{ID: "test", MType: "gauge", Value: floatPtr(42.5)}
	saved, err := repo.SaveMetric(context.Background(), metric)
	require.NoError(t, err)
	assert.Equal(t, "test", saved.ID)
}

func TestSaveMetric_PersistentError(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockMem := mocks.NewMockMemStorage(ctrl)
	mockPersistent := mocks.NewMockPersistentStorage(ctrl)

	mockMem.EXPECT().SaveMetric(gomock.Any()).Times(1)
	mockPersistent.EXPECT().SaveMetric(gomock.Any(), gomock.Any()).Return(errors.New("disk full")).Times(1)

	repo := New(mockMem, mockPersistent, 0, zap.NewNop())

	metric := models.Metrics{ID: "test", MType: "gauge", Value: floatPtr(42.5)}
	_, err := repo.SaveMetric(context.Background(), metric)
	assert.Error(t, err)
}

func TestSaveMetricsBatch_Deferred(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockMem := mocks.NewMockMemStorage(ctrl)
	mockMem.EXPECT().SaveBatch(gomock.Any()).Return(intPtr(2)).Times(1)

	repo := New(mockMem, mocks.NewMockPersistentStorage(ctrl), 300, zap.NewNop())

	metrics := []models.Metrics{
		{ID: "m1", MType: "gauge", Value: floatPtr(1.0)},
		{ID: "m2", MType: "counter", Delta: int64Ptr(2)},
	}

	count, err := repo.SaveMetricsBatch(context.Background(), metrics)
	require.NoError(t, err)
	assert.Equal(t, 2, *count)
}

func TestSaveMetricsBatch_Immediate(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockMem := mocks.NewMockMemStorage(ctrl)
	mockPersistent := mocks.NewMockPersistentStorage(ctrl)

	mockMem.EXPECT().SaveBatch(gomock.Any()).Return(intPtr(2)).Times(1)
	mockPersistent.EXPECT().SaveBatch(gomock.Any(), gomock.Any()).Return(intPtr(2), nil).Times(1)

	repo := New(mockMem, mockPersistent, 0, zap.NewNop())

	metrics := []models.Metrics{
		{ID: "m1", MType: "gauge", Value: floatPtr(1.0)},
	}

	count, err := repo.SaveMetricsBatch(context.Background(), metrics)
	require.NoError(t, err)
	assert.Equal(t, 2, *count)
}

func TestGetAllMetrics(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	expected := map[string]models.Metrics{
		"cpu": {ID: "cpu", MType: "gauge", Value: floatPtr(80.5)},
	}

	mockMem := mocks.NewMockMemStorage(ctrl)
	mockMem.EXPECT().GetAll().Return(expected).Times(1)

	repo := New(mockMem, nil, 300, zap.NewNop())

	all := repo.GetAllMetrics()
	assert.Len(t, all, 1)
	assert.Equal(t, 80.5, *all["cpu"].Value)
}

func TestGetMetric(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	expected := &models.Metrics{ID: "cpu", MType: "gauge", Value: floatPtr(80.5)}

	mockMem := mocks.NewMockMemStorage(ctrl)
	mockMem.EXPECT().GetMetricByName("cpu").Return(expected, true).Times(1)

	repo := New(mockMem, nil, 300, zap.NewNop())

	metric, ok := repo.GetMetric("cpu")
	require.True(t, ok)
	assert.Equal(t, 80.5, *metric.Value)
}

func TestGetMetric_NotFound(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockMem := mocks.NewMockMemStorage(ctrl)
	mockMem.EXPECT().GetMetricByName("nonexistent").Return(nil, false).Times(1)

	repo := New(mockMem, nil, 300, zap.NewNop())

	_, ok := repo.GetMetric("nonexistent")
	assert.False(t, ok)
}

func TestPing_NoPersistent(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	repo := New(mocks.NewMockMemStorage(ctrl), nil, 300, zap.NewNop())
	err := repo.Ping(context.Background())
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "ping unavailable")
}

func TestPing_Persistent(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockPersistent := mocks.NewMockPersistentStorage(ctrl)
	mockPersistent.EXPECT().Ping(gomock.Any()).Return(nil).Times(1)

	repo := New(mocks.NewMockMemStorage(ctrl), mockPersistent, 300, zap.NewNop())
	err := repo.Ping(context.Background())
	assert.NoError(t, err)
}

func TestBackupMetrics(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	metricsMap := map[string]models.Metrics{
		"cpu": {ID: "cpu", MType: "gauge", Value: floatPtr(80.5)},
	}

	mockMem := mocks.NewMockMemStorage(ctrl)
	mockPersistent := mocks.NewMockPersistentStorage(ctrl)

	mockMem.EXPECT().GetAll().Return(metricsMap).Times(1)
	mockPersistent.EXPECT().SaveBatch(gomock.Any(), gomock.Any()).Return(nil, nil).Times(1)

	repo := New(mockMem, mockPersistent, 300, zap.NewNop())
	err := repo.BackupMetrics(context.Background())
	assert.NoError(t, err)
}

func TestRestoreMetrics(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	metrics := []models.Metrics{
		{ID: "cpu", MType: "gauge", Value: floatPtr(80.5)},
	}

	mockMem := mocks.NewMockMemStorage(ctrl)
	mockPersistent := mocks.NewMockPersistentStorage(ctrl)

	mockPersistent.EXPECT().GetAll(gomock.Any()).Return(metrics, nil).Times(1)
	mockMem.EXPECT().SaveBatch(metrics).Return(nil).Times(1)

	repo := New(mockMem, mockPersistent, 300, zap.NewNop())
	err := repo.RestoreMetrics(context.Background())
	assert.NoError(t, err)
}

func floatPtr(v float64) *float64 { return &v }
func int64Ptr(v int64) *int64     { return &v }
func intPtr(v int) *int           { return &v }
