package repository

import (
	"encoding/json"
	"errors"
	"io"
	models "metrics-collector/internal/model"
	"os"
)

type FileStorage struct {
	file string
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
