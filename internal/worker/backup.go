package worker

import (
	"time"

	"go.uber.org/zap"
)

type BackupWorker struct {
	storeInterval int
	repo          Repository
	logger        *zap.SugaredLogger
}

type Repository interface {
	BackupMetrics() error
}

func NewBackupWorker(storeInterval int, repo Repository, logger *zap.SugaredLogger) *BackupWorker {
	return &BackupWorker{
		storeInterval: storeInterval,
		repo:          repo,
		logger:        logger,
	}
}

func (bw *BackupWorker) Start() {
	if bw.storeInterval <= 0 {
		return
	}

	go func() {
		for {
			time.Sleep(time.Duration(bw.storeInterval) * time.Second)
			if err := bw.repo.BackupMetrics(); err != nil {
				bw.logger.Errorw("backup worker err", "error", err)
			}
		}
	}()
}
