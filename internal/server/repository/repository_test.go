package repository

import (
	"context"
	"errors"
	"testing"

	"metrics-collector/internal/config"
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

	repo := &Repository{
		mem:      mockMem,
		logger:   zap.NewNop(),
		mode:     ModeMemOnly,
		syncMode: SyncDeferred,
	}

	metric := models.Metrics{ID: "test", MType: "gauge", Value: newFloat64(42.5)}
	saved, err := repo.SaveMetric(context.Background(), metric)
	require.NoError(t, err)
	assert.Equal(t, "test", saved.ID)
}

func TestSaveMetric_ImmediateFile(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockMem := mocks.NewMockMemStorage(ctrl)
	mockFile := mocks.NewMockFileStorage(ctrl)

	mockMem.EXPECT().SaveMetric(gomock.Any()).Times(1)
	mockFile.EXPECT().SaveMetric(gomock.Any()).Return(nil).Times(1)

	repo := &Repository{
		mem:      mockMem,
		file:     mockFile,
		logger:   zap.NewNop(),
		mode:     ModeFile,
		syncMode: SyncImmediate,
	}

	metric := models.Metrics{ID: "test", MType: "gauge", Value: newFloat64(42.5)}
	saved, err := repo.SaveMetric(context.Background(), metric)
	require.NoError(t, err)
	assert.Equal(t, "test", saved.ID)
}

func TestSaveMetric_FileError(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockMem := mocks.NewMockMemStorage(ctrl)
	mockFile := mocks.NewMockFileStorage(ctrl)

	mockMem.EXPECT().SaveMetric(gomock.Any()).Times(1)
	mockFile.EXPECT().SaveMetric(gomock.Any()).Return(errors.New("disk full")).Times(1)

	repo := &Repository{
		mem:      mockMem,
		file:     mockFile,
		logger:   zap.NewNop(),
		mode:     ModeFile,
		syncMode: SyncImmediate,
	}

	metric := models.Metrics{ID: "test", MType: "gauge", Value: newFloat64(42.5)}
	_, err := repo.SaveMetric(context.Background(), metric)
	assert.Error(t, err)
}

func TestSaveMetricsBatch_DeferredFile(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockMem := mocks.NewMockMemStorage(ctrl)
	mockMem.EXPECT().SaveBatch(gomock.Any()).Return(newInt(2)).Times(1)

	repo := &Repository{
		mem:      mockMem,
		file:     mocks.NewMockFileStorage(ctrl),
		logger:   zap.NewNop(),
		mode:     ModeFile,
		syncMode: SyncDeferred,
	}

	metrics := []models.Metrics{
		{ID: "m1", MType: "gauge", Value: newFloat64(1.0)},
		{ID: "m2", MType: "counter", Delta: newInt64(2)},
	}

	count, err := repo.SaveMetricsBatch(context.Background(), metrics)
	require.NoError(t, err)
	assert.Equal(t, 2, *count)
}

func TestSaveMetricsBatch_ImmediateFile(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockMem := mocks.NewMockMemStorage(ctrl)
	mockFile := mocks.NewMockFileStorage(ctrl)

	mockMem.EXPECT().SaveBatch(gomock.Any()).Return(newInt(2)).Times(1)
	mockFile.EXPECT().SaveBatch(gomock.Any()).Return(newInt(2), nil).Times(1)

	repo := &Repository{
		mem:      mockMem,
		file:     mockFile,
		logger:   zap.NewNop(),
		mode:     ModeFile,
		syncMode: SyncImmediate,
	}

	metrics := []models.Metrics{
		{ID: "m1", MType: "gauge", Value: newFloat64(1.0)},
		{ID: "m2", MType: "counter", Delta: newInt64(2)},
	}

	count, err := repo.SaveMetricsBatch(context.Background(), metrics)
	require.NoError(t, err)
	assert.Equal(t, 2, *count)
}

func TestSaveMetricsBatch_Postgres(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockMem := mocks.NewMockMemStorage(ctrl)
	mockPg := mocks.NewMockPostgresStorage(ctrl)

	mockMem.EXPECT().SaveBatch(gomock.Any()).Return(newInt(3)).Times(1)
	mockPg.EXPECT().SaveBatch(gomock.Any(), gomock.Any()).Return(newInt(3), nil).Times(1)

	repo := &Repository{
		mem:      mockMem,
		pg:       mockPg,
		logger:   zap.NewNop(),
		mode:     ModePostgres,
		syncMode: SyncImmediate,
	}

	metrics := []models.Metrics{
		{ID: "m1", MType: "gauge", Value: newFloat64(1.0)},
	}

	count, err := repo.SaveMetricsBatch(context.Background(), metrics)
	require.NoError(t, err)
	assert.Equal(t, 3, *count)
}

func TestGetAllMetrics(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	expected := map[string]models.Metrics{
		"cpu": {ID: "cpu", MType: "gauge", Value: newFloat64(80.5)},
	}

	mockMem := mocks.NewMockMemStorage(ctrl)
	mockMem.EXPECT().GetAll().Return(expected).Times(1)

	repo := &Repository{
		mem:    mockMem,
		logger: zap.NewNop(),
	}

	all := repo.GetAllMetrics()
	assert.Len(t, all, 1)
	assert.Equal(t, 80.5, *all["cpu"].Value)
}

func TestGetMetric(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	expected := &models.Metrics{ID: "cpu", MType: "gauge", Value: newFloat64(80.5)}

	mockMem := mocks.NewMockMemStorage(ctrl)
	mockMem.EXPECT().GetMetricByName("cpu").Return(expected, true).Times(1)

	repo := &Repository{
		mem:    mockMem,
		logger: zap.NewNop(),
	}

	metric, ok := repo.GetMetric("cpu")
	require.True(t, ok)
	assert.Equal(t, 80.5, *metric.Value)
}

func TestGetMetric_NotFound(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockMem := mocks.NewMockMemStorage(ctrl)
	mockMem.EXPECT().GetMetricByName("nonexistent").Return(nil, false).Times(1)

	repo := &Repository{
		mem:    mockMem,
		logger: zap.NewNop(),
	}

	_, ok := repo.GetMetric("nonexistent")
	assert.False(t, ok)
}

func TestPing_MemOnly(t *testing.T) {
	repo := &Repository{
		mode:   ModeMemOnly,
		logger: zap.NewNop(),
	}

	err := repo.Ping(context.Background())
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "ping unavailable")
}

func TestPing_Postgres(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockPg := mocks.NewMockPostgresStorage(ctrl)
	mockPg.EXPECT().Ping(gomock.Any()).Return(nil).Times(1)

	repo := &Repository{
		pg:     mockPg,
		mode:   ModePostgres,
		logger: zap.NewNop(),
	}

	err := repo.Ping(context.Background())
	assert.NoError(t, err)
}

func TestBackupMetrics_Postgres(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	metricsMap := map[string]models.Metrics{
		"cpu": {ID: "cpu", MType: "gauge", Value: newFloat64(80.5)},
	}

	mockMem := mocks.NewMockMemStorage(ctrl)
	mockPg := mocks.NewMockPostgresStorage(ctrl)

	mockMem.EXPECT().GetAll().Return(metricsMap).Times(1)
	mockPg.EXPECT().SaveBatch(gomock.Any(), gomock.Any()).Return(nil, nil).Times(1)

	repo := &Repository{
		mem:    mockMem,
		pg:     mockPg,
		mode:   ModePostgres,
		logger: zap.NewNop(),
	}

	err := repo.BackupMetrics(context.Background())
	assert.NoError(t, err)
}

func TestBackupMetrics_File(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	metricsMap := map[string]models.Metrics{
		"cpu": {ID: "cpu", MType: "gauge", Value: newFloat64(80.5)},
	}

	mockMem := mocks.NewMockMemStorage(ctrl)
	mockFile := mocks.NewMockFileStorage(ctrl)

	mockMem.EXPECT().GetAll().Return(metricsMap).Times(1)
	mockFile.EXPECT().SaveBatch(gomock.Any()).Return(nil, nil).Times(1)

	repo := &Repository{
		mem:    mockMem,
		file:   mockFile,
		mode:   ModeFile,
		logger: zap.NewNop(),
	}

	err := repo.BackupMetrics(context.Background())
	assert.NoError(t, err)
}

func TestRestoreMetrics_Postgres(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	metrics := []models.Metrics{
		{ID: "cpu", MType: "gauge", Value: newFloat64(80.5)},
	}

	mockMem := mocks.NewMockMemStorage(ctrl)
	mockPg := mocks.NewMockPostgresStorage(ctrl)

	mockPg.EXPECT().GetAll(gomock.Any()).Return(metrics, nil).Times(1)
	mockMem.EXPECT().SaveBatch(metrics).Return(nil).Times(1)

	repo := &Repository{
		mem:    mockMem,
		pg:     mockPg,
		mode:   ModePostgres,
		logger: zap.NewNop(),
	}

	err := repo.RestoreMetrics(context.Background())
	assert.NoError(t, err)
}

func TestRestoreMetrics_File(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	metrics := []models.Metrics{
		{ID: "cpu", MType: "gauge", Value: newFloat64(80.5)},
	}

	mockMem := mocks.NewMockMemStorage(ctrl)
	mockFile := mocks.NewMockFileStorage(ctrl)

	mockFile.EXPECT().GetAll().Return(metrics, nil).Times(1)
	mockMem.EXPECT().SaveBatch(metrics).Return(nil).Times(1)

	repo := &Repository{
		mem:    mockMem,
		file:   mockFile,
		mode:   ModeFile,
		logger: zap.NewNop(),
	}

	err := repo.RestoreMetrics(context.Background())
	assert.NoError(t, err)
}

func TestNewRepository_MemOnly(t *testing.T) {
	cfg := &config.ServerConfig{
		DatabaseDSN:     "",
		FileStoragePath: "",
		StoreInterval:   300,
	}

	repo, err := NewRepository(cfg, zap.NewNop())
	require.NoError(t, err)
	assert.NotNil(t, repo)
	assert.Equal(t, ModeMemOnly, repo.mode)
	assert.Equal(t, SyncDeferred, repo.syncMode)
}

func TestNewRepository_FileStorage(t *testing.T) {
	tmpDir := t.TempDir()
	filePath := tmpDir + "/metrics.json"

	cfg := &config.ServerConfig{
		DatabaseDSN:     "",
		FileStoragePath: filePath,
		StoreInterval:   0,
	}

	repo, err := NewRepository(cfg, zap.NewNop())
	require.NoError(t, err)
	assert.Equal(t, ModeFile, repo.mode)
	assert.Equal(t, SyncImmediate, repo.syncMode)
}

func TestNewRepository_PostgresDSN_Invalid(t *testing.T) {
	cfg := &config.ServerConfig{
		DatabaseDSN:     "postgres://invalid:5432/db",
		FileStoragePath: "",
	}

	_, err := NewRepository(cfg, zap.NewNop())
	assert.Error(t, err)
}

func newFloat64(v float64) *float64 {
	return &v
}

func newInt64(v int64) *int64 {
	return &v
}

func newInt(v int) *int {
	return &v
}
