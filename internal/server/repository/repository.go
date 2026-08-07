// Package repository определяет интерфейсы для хранения метрик.
package repository

import (
	"context"
	"fmt"
	"metrics-collector/internal/config"
	models "metrics-collector/internal/server/model"
	"metrics-collector/internal/server/repository/file"
	"metrics-collector/internal/server/repository/mem"
	"metrics-collector/internal/server/repository/postgres"
	"time"

	"go.uber.org/zap"
)

//go:generate mockgen -destination=mocks/mock_mem_storage.go -package=mocks metrics-collector/internal/repository MemStorage
type MemStorage interface {
	SaveMetric(metric models.Metrics)
	SaveBatch(metrics []models.Metrics) *int
	GetAll() map[string]models.Metrics
	GetMetricByName(key string) (*models.Metrics, bool)
}

//go:generate mockgen -destination=mocks/mock_postgres_storage.go -package=mocks metrics-collector/internal/repository PostgresStorage
type PostgresStorage interface {
	Ping(ctx context.Context) error
	SaveMetric(ctx context.Context, m models.Metrics) error
	SaveBatch(ctx context.Context, metrics []models.Metrics) (*int, error)
	GetAll(ctx context.Context) ([]models.Metrics, error)
}

//go:generate mockgen -destination=mocks/mock_file_storage.go -package=mocks metrics-collector/internal/repository FileStorage
type FileStorage interface {
	SaveMetric(metric models.Metrics) error
	SaveBatch(metrics []models.Metrics) (*int, error)
	GetAll() ([]models.Metrics, error)
}

type Mode string

const (
	ModeFile     Mode = "file"
	ModePostgres Mode = "postgres"
	ModeMemOnly  Mode = "mem"
)

type SyncMode string

const (
	SyncImmediate SyncMode = "immediate"
	SyncDeferred  SyncMode = "deferred"
)

// Repository manages metric storage across memory, file, and database.
type Repository struct {
	file     FileStorage
	mem      MemStorage
	pg       PostgresStorage
	logger   *zap.Logger
	mode     Mode
	syncMode SyncMode
}

// NewRepository creates a Repository based on server configuration.
func NewRepository(cfg *config.ServerConfig, logger *zap.Logger) (*Repository, error) {

	r := &Repository{
		logger:   logger,
		mode:     ModeMemOnly,
		syncMode: SyncDeferred,
	}

	r.mem = mem.NewMemStorage()

	if cfg.DatabaseDSN == "" {
		logger.Info("Database DSN not provided, skipping PostgreSQL")
	} else {
		logger.Info("Attempting to connect to database...")
		pgs, err := postgres.NewPostgresStorage(cfg.DatabaseDSN, logger)
		if err != nil {
			return nil, fmt.Errorf("postgres connection failed: %w", err)
		}
		r.pg = pgs
		r.mode = ModePostgres
		logger.Info("Storage mode: postgres storage")
	}

	if r.mode != ModePostgres {
		logger.Info("PostgreSQL unavailable, falling back to file storage")

		if cfg.FileStoragePath == "" {
			logger.Info("File storage path not set, skipping file storage")
		} else {
			fls := file.NewFileStorage(cfg.FileStoragePath)
			r.file = fls
			r.mode = ModeFile
			logger.Info("Storage mode: file storage")
		}
	}

	switch r.mode {
	case ModePostgres, ModeFile:
		if cfg.StoreInterval == 0 {
			r.syncMode = SyncImmediate
			logger.Info("Sync mode: immediate")
		} else {
			logger.Info("Sync mode: deferred")
		}
	case ModeMemOnly:
		logger.Info("Storage mode: memory only")
	}

	return r, nil
}

// --- Health Check -------------------------------------------------

// Ping checks the database connection.
func (r *Repository) Ping(ctx context.Context) error {
	switch r.mode {
	case ModePostgres:
		return r.pg.Ping(ctx)
	default:
		return fmt.Errorf("ping unavailable")
	}
}

// --- Metrics CRUD -------------------------------------------------

// SaveMetric saves a metric to memory and syncs it to persistent storage.
func (r *Repository) SaveMetric(ctx context.Context, metric models.Metrics) (*models.Metrics, error) {
	r.mem.SaveMetric(metric)

	if r.syncMode == SyncDeferred {
		return &metric, nil
	}

	switch r.mode {
	case ModePostgres:
		if err := r.pg.SaveMetric(ctx, metric); err != nil {
			return nil, err
		}
	case ModeFile:
		if err := r.file.SaveMetric(metric); err != nil {
			return nil, err
		}
	}

	return &metric, nil
}

// GetAllMetrics returns all metrics from memory.
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

	if r.syncMode == SyncDeferred {
		return count, nil
	}

	switch r.mode {
	case ModePostgres:
		count, err := r.pg.SaveBatch(ctx, metrics)
		if err != nil {
			return nil, err
		}
		return count, nil
	case ModeFile:
		count, err := r.file.SaveBatch(metrics)
		if err != nil {
			return nil, err
		}
		return count, nil
	default:
		return count, nil
	}
}

// --- Backup -------------------------------------------------

// RestoreMetrics restores metrics from persistent storage into memory.
func (r *Repository) RestoreMetrics(ctx context.Context) error {
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	r.logger.Info("Restore metrics started...")
	start := time.Now()
	var metrics []models.Metrics
	switch r.mode {
	case ModePostgres:
		ms, err := r.pg.GetAll(ctx)
		if err != nil {
			return fmt.Errorf("restore metrics from postgres failed:%w", err)
		}
		metrics = ms
		r.logger.Info("Metrics restored",
			zap.Int("from_postgres", len(ms)),
			zap.Duration("took", time.Since(start)),
		)
	case ModeFile:
		ms, err := r.file.GetAll()
		if err != nil {
			return fmt.Errorf("restore metrics from file failed:%w", err)
		}
		metrics = ms
		r.logger.Info("Metrics restored",
			zap.Int("from_file", len(ms)),
			zap.Duration("took", time.Since(start)),
		)
	default:
		return fmt.Errorf("restore metrics failed: unsupported storage mode")
	}

	r.mem.SaveBatch(metrics)
	r.logger.Debug("restore metrics completed successfully")
	return nil
}

// BackupMetrics saves all in-memory metrics to persistent storage.
func (r *Repository) BackupMetrics(ctx context.Context) error {
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	metricsMap := r.mem.GetAll()
	metrics := make([]models.Metrics, 0, len(metricsMap))
	for _, metric := range metricsMap {
		metrics = append(metrics, metric)
	}

	switch r.mode {
	case ModePostgres:
		_, err := r.pg.SaveBatch(ctx, metrics)
		if err != nil {
			return err
		}
	case ModeFile:
		_, err := r.file.SaveBatch(metrics)
		if err != nil {
			return err
		}
	default:
		return fmt.Errorf("backup metrics failed: unsupported storage mode")
	}

	return nil
}
