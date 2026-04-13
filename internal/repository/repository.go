package repository

import (
	"metrics-collector/internal/config"
	models "metrics-collector/internal/model"

	"go.uber.org/zap"
)

type Repository struct {
	persistentStorage *FileStorage
	inMemoryStorage   *MemStorage
	storeInterval     int
	restore           bool
	logger            *zap.Logger
}

func NewRepository(cfg *config.ServerConfig, logger *zap.Logger) (*Repository, error) {
	r := &Repository{
		persistentStorage: newFileStorage(cfg.FileStoragePath),
		inMemoryStorage:   NewMemStorage(),
		storeInterval:     cfg.StoreInterval,
		restore:           cfg.Restore,
		logger:            logger,
	}

	if r.restore {
		r.RestoreMetrics()
	}

	if cfg.StoreInterval > 0 {
		go r.startBackupWorker()
	} else {
		r.logger.Info("Backup worker disabled (store_interval = 0)")
	}

	return r, nil
}

func (r *Repository) RestoreMetrics() error {
	metrics, err := r.persistentStorage.loadAllMetricsFrom()
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

func (r *Repository) UpdateMetric(metric models.Metrics) (*models.Metrics, error) {
	updated := r.inMemoryStorage.UpdateMetricByArgs(metric)

	if r.storeInterval == 0 {
		if err := r.persistentStorage.saveSingleMetricTo(updated); err != nil {
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

func (r *Repository) backupMetrics() error {
	metricsMap := r.inMemoryStorage.GetAllMetrics()

	metrics := make([]models.Metrics, 0, len(metricsMap))
	for _, metric := range metricsMap {
		metrics = append(metrics, metric)
	}

	if err := r.persistentStorage.replaceAllMetrics(metrics); err != nil {
		return err
	}

	return nil
}
