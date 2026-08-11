// Package file реализует файловое хранилище метрик.
package file

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	models "metrics-collector/internal/server/model"
	"os"
)

type FileStorage struct {
	file string
}

func NewFileStorage(file string) *FileStorage {
	return &FileStorage{
		file: file,
	}
}

func (f *FileStorage) SaveMetric(ctx context.Context, metric models.Metrics) error {
	metrics, err := f.GetAll(ctx)
	if err != nil {
		return err
	}
	updated := false

	for i, m := range metrics {
		if m.ID == metric.ID {
			metrics[i] = metric
			updated = true
			break
		}
	}

	if !updated {
		metrics = append(metrics, metric)
	}

	_, err = f.SaveBatch(ctx, metrics)
	if err != nil {
		return err
	}

	return nil
}

func (f *FileStorage) SaveBatch(ctx context.Context, metrics []models.Metrics) (*int, error) {
	data, err := json.MarshalIndent(metrics, "", "  ")
	if err != nil {
		return nil, err
	}

	if err := os.WriteFile(f.file, data, 0644); err != nil {
		return nil, err
	}

	savedCount := len(metrics)

	return &savedCount, nil
}

func (f *FileStorage) GetAll(ctx context.Context) ([]models.Metrics, error) {
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

// Ping всегда возвращает nil — файловому хранилищу не нужен ping.
func (f *FileStorage) Ping(ctx context.Context) error { return nil }

// Close всегда возвращает nil — файловое хранилище не держит открытых дескрипторов.
func (f *FileStorage) Close() error { return nil }
