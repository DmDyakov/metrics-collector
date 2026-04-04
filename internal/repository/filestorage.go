package repository

import (
	"encoding/json"
	"errors"
	"io"
	models "metrics-collector/internal/model"
	"os"
	"time"

	"go.uber.org/zap"
)

type FileStorage struct {
	file string
}

func (r *Repository) startBackupWorker() {
	go func() {
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
	}()
}

func newFileStorage(file string) *FileStorage {
	return &FileStorage{
		file: file,
	}
}

func (f *FileStorage) saveSingleMetricTo(metric *models.Metrics) error {
	metrics, err := f.loadAllMetricsFrom()
	if err != nil {
		return err
	}
	updated := false

	for i, m := range metrics {
		if m.ID == metric.ID {
			metrics[i] = *metric
			updated = true
			break
		}
	}

	if !updated {
		metrics = append(metrics, *metric)
	}

	return f.replaceAllMetrics(metrics)
}

func (f *FileStorage) replaceAllMetrics(metrics []models.Metrics) error {
	data, err := json.MarshalIndent(metrics, "", "  ")
	if err != nil {
		return err
	}

	return os.WriteFile(f.file, data, 0644)
}

func (f *FileStorage) loadAllMetricsFrom() ([]models.Metrics, error) {
	file, err := os.OpenFile(f.file, os.O_RDONLY|os.O_CREATE, 0666)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	var metrics []models.Metrics

	decoder := json.NewDecoder(file)
	if err := decoder.Decode(&metrics); err != nil {
		if errors.Is(err, io.EOF) {
			return []models.Metrics{}, nil
		}
		return nil, err
	}

	return metrics, nil
}
