package worker

import (
	"context"
	"time"

	"go.uber.org/zap"
)

type BackupRepository interface {
	RestoreMetrics(ctx context.Context) error
	BackupMetrics(ctx context.Context) error
}

// BackupWorker periodically saves in-memory metrics to persistent storage.
type BackupWorker struct {
	restore       bool
	storeInterval int
	repo          BackupRepository
	logger        *zap.Logger
}

// NewBackupWorker creates a new BackupWorker.
func NewBackupWorker(
	restore bool,
	storeInterval int,
	repo BackupRepository,
	logger *zap.Logger,
) *BackupWorker {
	return &BackupWorker{
		restore:       restore,
		storeInterval: storeInterval,
		repo:          repo,
		logger:        logger,
	}
}

// Run starts periodic backup of metrics.
func (b *BackupWorker) Run(ctx context.Context) error {
	if b.restore {
		if err := b.repo.RestoreMetrics(ctx); err != nil {
			return err
		}
	}

	if b.storeInterval <= 0 {
		b.logger.Info("Backup worker disabled (store_interval = 0)")
		return nil
	}

	ticker := time.NewTicker(time.Duration(b.storeInterval) * time.Second)
	defer ticker.Stop()

	b.logger.Info("Backup worker started", zap.Int("interval_seconds", b.storeInterval))

	for {
		select {
		case <-ctx.Done():
			b.logger.Info("Backup worker stopped")
			return ctx.Err()
		case <-ticker.C:
			if err := b.repo.BackupMetrics(ctx); err != nil {
				b.logger.Error("backup failed", zap.Error(err))
			} else {
				b.logger.Debug("backup completed successfully")
			}
		}
	}
}
