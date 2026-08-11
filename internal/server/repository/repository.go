// Package repository определяет интерфейсы для хранения метрик.
package repository

import (
	"context"
	"fmt"

	models "metrics-collector/internal/server/model"

	"go.uber.org/zap"
)

//go:generate mockgen -destination=mocks/mock_mem_storage.go -package=mocks metrics-collector/internal/server/repository MemStorage
type MemStorage interface {
	SaveMetric(metric models.Metrics)
	SaveBatch(metrics []models.Metrics) *int
	GetAll() map[string]models.Metrics
	GetMetricByName(key string) (*models.Metrics, bool)
}

//go:generate mockgen -destination=mocks/mock_persistent_storage.go -package=mocks metrics-collector/internal/server/repository PersistentStorage
type PersistentStorage interface {
	Ping(ctx context.Context) error
	SaveMetric(ctx context.Context, m models.Metrics) error
	SaveBatch(ctx context.Context, metrics []models.Metrics) (*int, error)
	GetAll(ctx context.Context) ([]models.Metrics, error)
	Close() error
}

type SyncMode string

const (
	SyncImmediate SyncMode = "immediate"
	SyncDeferred  SyncMode = "deferred"
)

// Repository управляет хранением метрик в памяти и персистентном хранилище.
type Repository struct {
	mem        MemStorage
	persistent PersistentStorage
	logger     *zap.Logger
	syncMode   SyncMode
}

// New создаёт Repository.
func New(mem MemStorage, persistent PersistentStorage, storeInterval int, logger *zap.Logger) *Repository {
	syncMode := SyncDeferred
	if storeInterval == 0 && persistent != nil {
		syncMode = SyncImmediate
	}

	return &Repository{
		mem:        mem,
		persistent: persistent,
		logger:     logger,
		syncMode:   syncMode,
	}
}

// Ping checks the database connection.
func (r *Repository) Ping(ctx context.Context) error {
	if r.persistent == nil {
		return fmt.Errorf("ping unavailable")
	}
	return r.persistent.Ping(ctx)
}

// SaveMetric saves a metric to memory and syncs it to persistent storage.
func (r *Repository) SaveMetric(ctx context.Context, metric models.Metrics) (*models.Metrics, error) {
	r.mem.SaveMetric(metric)

	if r.syncMode == SyncImmediate && r.persistent != nil {
		if err := r.persistent.SaveMetric(ctx, metric); err != nil {
			return nil, err
		}
	}

	return &metric, nil
}

func (r *Repository) GetAllMetrics() map[string]models.Metrics {
	return r.mem.GetAll()
}

// GetMetric returns a metric by name from memory.
func (r *Repository) GetMetric(metricName string) (*models.Metrics, bool) {
	return r.mem.GetMetricByName(metricName)
}

// SaveMetricsBatch batch-saves metrics to memory and syncs them to persistent storage.
func (r *Repository) SaveMetricsBatch(ctx context.Context, metrics []models.Metrics) (*int, error) {
	count := r.mem.SaveBatch(metrics)

	if r.syncMode == SyncImmediate && r.persistent != nil {
		count, err := r.persistent.SaveBatch(ctx, metrics)
		if err != nil {
			return nil, err
		}
		return count, nil
	}

	return count, nil
}
