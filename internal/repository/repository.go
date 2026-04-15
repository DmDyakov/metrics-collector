package repository

import (
	"context"
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
	pgs, err := newPostgresStorage(cfg.DatabaseDSN)
	if err != nil {
		logger.Warn("PostgreSQL is unavailable", zap.Error(err))
		pgs = nil
	}

	var fls *FileStorage
	if pgs == nil {
		fls, err = newFileStorage(cfg.FileStoragePath)
		if err != nil {
			logger.Warn("File is unavailable", zap.Error(err))
			fls = nil
		}
	}

	r := &Repository{
		inMemoryStorage: newMemStorage(),
		fileStorage:     fls,
		//TODO: заменить на pgs
		postgresStorage: nil,
		storeInterval:   cfg.StoreInterval,
		restore:         cfg.Restore,
		logger:          logger,
	}

	if r.fileStorage == nil && r.postgresStorage == nil {
		return r, nil
	}

	if r.restore {
		r.restoreMetrics()
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
	return nil
}

// --- Metrics CRUD -------------------------------------------------

func (r *Repository) UpdateMetric(metric models.Metrics) (*models.Metrics, error) {
	updated := r.inMemoryStorage.UpdateMetricByArgs(metric)

	if r.storeInterval != 0 {
		return updated, nil
	}

	if r.postgresStorage != nil {
		if err := r.postgresStorage.saveSingleMetricTo(updated); err != nil {
			return nil, err
		}
	}

	if r.fileStorage != nil {
		if err := r.fileStorage.saveSingleMetricTo(updated); err != nil {
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
	metricsMap := r.inMemoryStorage.GetAllMetrics()

	metrics := make([]models.Metrics, 0, len(metricsMap))
	for _, metric := range metricsMap {
		metrics = append(metrics, metric)
	}

	if r.postgresStorage != nil {
		// TODO: Заменить на запрос к бд
		return nil
	}

	if err := r.fileStorage.replaceAllMetrics(metrics); err != nil {
		return err
	}

	return nil
}

func (r *Repository) restoreMetrics() error {
	metrics, err := r.loadAllMetricsFromStorage()
	if err != nil {
		return err
	}

	metricsMap := make(map[string]models.Metrics, len(metrics))
	for _, metric := range metrics {
		metricsMap[metric.ID] = metric
	}

	r.inMemoryStorage.replaceMetrics(metricsMap)

	return nil
}

func (r *Repository) loadAllMetricsFromStorage() ([]models.Metrics, error) {
	if r.postgresStorage != nil {
		// TODO: Заменить на запрос к бд
		return []models.Metrics{}, nil
	}

	if r.fileStorage != nil {
		return r.fileStorage.loadAllMetricsFrom()
	}

	return []models.Metrics{}, nil
}
