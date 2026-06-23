package worker

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"go.uber.org/zap"
)

type mockBackupRepo struct {
	restoreErr error
	backupErr  error
}

func (m *mockBackupRepo) RestoreMetrics(ctx context.Context) error { return m.restoreErr }
func (m *mockBackupRepo) BackupMetrics(ctx context.Context) error  { return m.backupErr }

func TestBackupWorker_Run(t *testing.T) {
	t.Run("restore on start", func(t *testing.T) {
		repo := &mockBackupRepo{}
		bw := NewBackupWorker(true, 0, repo, zap.NewNop())

		ctx, cancel := context.WithCancel(context.Background())
		cancel()

		err := bw.Run(ctx)
		assert.NoError(t, err)
	})

	t.Run("disabled when interval is 0", func(t *testing.T) {
		repo := &mockBackupRepo{}
		bw := NewBackupWorker(false, 0, repo, zap.NewNop())

		ctx, cancel := context.WithTimeout(context.Background(), 50*time.Millisecond)
		defer cancel()

		err := bw.Run(ctx)
		assert.NoError(t, err)
	})
}

func TestBackupWorker_Run_Disabled(t *testing.T) {
	t.Run("disabled when interval is 0", func(t *testing.T) {
		repo := &mockBackupRepo{}
		bw := NewBackupWorker(false, 0, repo, zap.NewNop())

		ctx, cancel := context.WithTimeout(context.Background(), 50*time.Millisecond)
		defer cancel()

		err := bw.Run(ctx)
		assert.NoError(t, err)
	})
}

func TestBackupWorker_New(t *testing.T) {
	t.Run("creates worker with correct fields", func(t *testing.T) {
		repo := &mockBackupRepo{}
		bw := NewBackupWorker(true, 10, repo, zap.NewNop())
		assert.NotNil(t, bw)
		assert.True(t, bw.restore)
		assert.Equal(t, 10, bw.storeInterval)
		assert.NotNil(t, bw.repo)
		assert.NotNil(t, bw.logger)
	})
}
