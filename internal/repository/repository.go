package repository

import (
	"metrics-collector/internal/config"
	models "metrics-collector/internal/model"
)

type Repository struct {
	persistentStorage *FileStorage
	inMemoryStorage   *MemStorage
	storeInterval     int
	restore           bool
}

func NewRepository(cfg *config.ServerConfig) (*Repository, error) {
	r := &Repository{
		persistentStorage: NewFileStorage(cfg.FileStoragePath),
		inMemoryStorage:   NewMemStorage(),
		storeInterval:     cfg.StoreInterval,
		restore:           cfg.Restore,
	}

	if r.restore {
		r.RestoreMetrics()
	}
	return r, nil
}

func (r *Repository) RestoreMetrics() error {
	metrics, err := r.persistentStorage.LoadAllMetricsFrom()
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
	updated := r.inMemoryStorage.UpdateMetric(metric)

	if r.storeInterval == 0 {
		if err := r.persistentStorage.SaveSingleMetricTo(updated); err != nil {
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

func (r *Repository) BackupMetrics() error {
	metricsMap := r.inMemoryStorage.GetAllMetrics()

	metrics := make([]models.Metrics, 0, len(metricsMap))
	for _, metric := range metricsMap {
		metrics = append(metrics, metric)
	}

	if err := r.persistentStorage.ReplaceAllMetrics(metrics); err != nil {
		return err
	}

	return nil
}
