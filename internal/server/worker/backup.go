// Package worker содержит фоновые задачи
package worker

import (
	"context"
	"fmt"
	models "metrics-collector/internal/server/model"
	"time"

	"go.uber.org/zap"
)

//go:generate mockgen -destination=mocks/mock_persistent_storage.go -package=mocks . PersistentStorage
type PersistentStorage interface {
	SaveBatch(ctx context.Context, metrics []models.Metrics) (*int, error)
	GetAll(ctx context.Context) ([]models.Metrics, error)
}

//go:generate mockgen -destination=mocks/mock_mem_storage.go -package=mocks . MemStorage
type MemStorage interface {
	SaveBatch(metrics []models.Metrics) *int
	GetAll() map[string]models.Metrics
}

// BackupWorker periodically saves in-memory metrics to persistent storage.
type BackupWorker struct {
	restore    bool
	interval   int
	timeout    time.Duration
	mem        MemStorage
	persistent PersistentStorage
	logger     *zap.Logger
}

// NewBackupWorker creates a new BackupWorker.
func NewBackupWorker(
	restore bool,
	interval int,
	timeout time.Duration,
	mem MemStorage,
	persistent PersistentStorage,
	logger *zap.Logger,
) *BackupWorker {
	return &BackupWorker{
		restore:    restore,
		interval:   interval,
		timeout:    timeout,
		mem:        mem,
		persistent: persistent,
		logger:     logger,
	}
}

// Run starts periodic backup of metrics.
func (b *BackupWorker) Run(ctx context.Context) error {
	if b.restore {
		restoreCtx, cancel := context.WithTimeout(ctx, b.timeout)
		defer cancel()

		if err := b.Restore(restoreCtx); err != nil {
			return err
		}
	}

	if b.interval <= 0 {
		b.logger.Info("Backup worker disabled (store_interval = 0)")
		return nil
	}

	ticker := time.NewTicker(time.Duration(b.interval) * time.Second)
	defer ticker.Stop()

	b.logger.Info("Backup worker started", zap.Int("interval_seconds", b.interval))

	for {
		select {
		case <-ctx.Done():
			b.logger.Info("Backup worker stopped")
			return nil
		case <-ticker.C:
			backupCtx, cancel := context.WithTimeout(context.Background(), b.timeout)

			if err := b.Backup(backupCtx); err != nil {
				b.logger.Error("backup failed", zap.Error(err))
			} else {
				b.logger.Debug("backup completed successfully")
			}
			cancel()
		}
	}
}

// Restore restores metrics from persistent storage into memory.
func (b *BackupWorker) Restore(ctx context.Context) error {
	if b.persistent == nil {
		return fmt.Errorf("restore unavailable: no persistent storage")
	}

	ctx, cancel := context.WithTimeout(ctx, b.timeout)
	defer cancel()

	b.logger.Info("Restoring metrics...")
	start := time.Now()

	metrics, err := b.persistent.GetAll(ctx)
	if err != nil {
		return fmt.Errorf("restore metrics failed: %w", err)
	}

	b.mem.SaveBatch(metrics)
	b.logger.Info("Metrics restored",
		zap.Int("count", len(metrics)),
		zap.Duration("took", time.Since(start)),
	)

	return nil
}

// Backup saves all in-memory metrics to persistent storage.
func (b *BackupWorker) Backup(ctx context.Context) error {
	if b.persistent == nil {
		return fmt.Errorf("backup unavailable: no persistent storage")
	}

	ctx, cancel := context.WithTimeout(ctx, b.timeout)
	defer cancel()

	metricsMap := b.mem.GetAll()
	metrics := make([]models.Metrics, 0, len(metricsMap))
	for _, m := range metricsMap {
		metrics = append(metrics, m)
	}

	_, err := b.persistent.SaveBatch(ctx, metrics)
	return err
}
