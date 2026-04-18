package repository

import (
	"context"
	"errors"
	"fmt"
	"metrics-collector/internal/config"
	models "metrics-collector/internal/model"
	"time"

	"go.uber.org/zap"
)

type Repository struct {
	fileStorage     *FileStorage
	inMemoryStorage *MemStorage
	postgresStorage *PostgresStorage
	storeInterval   int
	restore         bool
	logger          *zap.Logger
}

func NewRepository(cfg *config.ServerConfig, logger *zap.Logger) (*Repository, error) {
	r := &Repository{
		inMemoryStorage: newMemStorage(),
		fileStorage:     nil,
		postgresStorage: nil,
		storeInterval:   cfg.StoreInterval,
		restore:         cfg.Restore,
		logger:          logger,
	}

	if cfg.DatabaseDSN == "" {
		logger.Info("Database DSN not provided, skipping PostgreSQL")
	} else {
		logger.Info("Attempting to connect to database...")
		pgs, err := newPostgresStorage(cfg.DatabaseDSN)
		if err != nil {
			return nil, fmt.Errorf("postgres connection failed: %w", err)
		}
		r.postgresStorage = pgs
	}

	if r.postgresStorage == nil {
		logger.Info("PostgreSQL unavailable, falling back to file storage")

		if cfg.FileStoragePath == "" {
			logger.Info("File storage path not set, using in-memory storage only")
		} else {
			fls := newFileStorage(cfg.FileStoragePath)
			r.fileStorage = fls
		}
	}

	if r.fileStorage == nil && r.postgresStorage == nil {
		return r, nil
	}

	if r.restore {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		r.restoreMetrics(ctx)
	}

	if cfg.StoreInterval > 0 {
		go r.startBackupWorker()
	} else {
		r.logger.Info("Backup worker disabled (store_interval = 0)")
	}

	return r, nil
}

// --- Health Check -------------------------------------------------

func (r *Repository) Ping(ctx context.Context) error {
	if r.postgresStorage != nil {
		return r.postgresStorage.db.PingContext(ctx)
	}

	return errors.New("PostgreSQL is unavailable")
}

// --- Metrics CRUD -------------------------------------------------

func (r *Repository) SaveMetric(ctx context.Context, metric models.Metrics) (*models.Metrics, error) {
	updated := r.inMemoryStorage.UpdateMetricByArgs(metric)

	if r.storeInterval != 0 {
		return updated, nil
	}

	if r.postgresStorage != nil {
		if err := r.postgresStorage.saveMetric(ctx, updated); err != nil {
			return nil, err
		}
	}

	if r.fileStorage != nil {
		if err := r.fileStorage.saveMetric(updated); err != nil {
			return nil, err
		}
	}

	return updated, nil

}

func (r *Repository) GetAllMetrics() map[string]models.Metrics {
	return r.inMemoryStorage.GetAllMetrics()
}

func (r *Repository) GetMetric(metricName string) (*models.Metrics, bool) {
	return r.inMemoryStorage.GetMetric(metricName)
}

func (r *Repository) SaveMetricsBatch(ctx context.Context, metrics []models.Metrics) (*int, error) {
	count := r.inMemoryStorage.saveMetricsBatch(metrics)

	if r.storeInterval != 0 {
		return count, nil
	}

	if r.postgresStorage != nil {
		c, err := r.postgresStorage.saveMetricsBatch(ctx, metrics)
		if err != nil {
			return nil, err
		}

		count = c
	}

	if r.fileStorage != nil {
		c, err := r.fileStorage.saveMetricsBatch(metrics)
		if err != nil {
			return nil, err
		}

		count = c
	}

	return count, nil
}

// --- Background Backup & Restore ----------------------------------

func (r *Repository) startBackupWorker() {
	ticker := time.NewTicker(time.Duration(r.storeInterval) * time.Second)
	defer ticker.Stop()

	r.logger.Info("Backup worker started",
		zap.Int("interval_seconds", r.storeInterval),
	)

	for range ticker.C {
		if err := r.backupMetrics(); err != nil {
			r.logger.Error("backup failed", zap.Error(err))
		} else {
			r.logger.Debug("backup completed successfully")
		}
	}
}

func (r *Repository) backupMetrics() error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	metricsMap := r.inMemoryStorage.GetAllMetrics()
	metrics := make([]models.Metrics, 0, len(metricsMap))
	for _, metric := range metricsMap {
		metrics = append(metrics, metric)
	}

	if r.postgresStorage != nil {
		_, err := r.postgresStorage.saveMetricsBatch(ctx, metrics)
		if err != nil {
			return err
		}
	}

	if r.fileStorage != nil {
		_, err := r.fileStorage.saveMetricsBatch(metrics)
		if err != nil {
			return err
		}
	}

	return nil
}

func (r *Repository) restoreMetrics(ctx context.Context) error {
	metrics, err := r.loadAllMetricsFromStorage(ctx)
	if err != nil {
		return err
	}

	r.inMemoryStorage.saveMetricsBatch(metrics)

	return nil
}

func (r *Repository) loadAllMetricsFromStorage(ctx context.Context) ([]models.Metrics, error) {
	metrics := []models.Metrics{}

	if r.postgresStorage != nil {
		ms, err := r.postgresStorage.loadAllMetrics(ctx)
		if err != nil {
			return nil, err
		}
		metrics = ms
	}

	if r.fileStorage != nil && len(metrics) == 0 {
		return r.fileStorage.loadAllMetrics()
	}

	return metrics, nil
}
