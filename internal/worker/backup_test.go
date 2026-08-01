package worker

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"go.uber.org/zap"
)

type mockBackupRepo struct {
	restoreCalls int
	backupCalls  int
	restoreErr   error
	backupErr    error
}

func (m *mockBackupRepo) RestoreMetrics(ctx context.Context) error {
	m.restoreCalls++
	return m.restoreErr
}

func (m *mockBackupRepo) BackupMetrics(ctx context.Context) error {
	m.backupCalls++
	return m.backupErr
}

func TestBackupWorker_Run(t *testing.T) {
	t.Run("restore on start", func(t *testing.T) {
		repo := &mockBackupRepo{}
		bw := NewBackupWorker(true, 0, repo, zap.NewNop())

		ctx, cancel := context.WithCancel(context.Background())
		cancel()

		err := bw.Run(ctx)
		assert.NoError(t, err)
		assert.Equal(t, 1, repo.restoreCalls)
	})

	t.Run("restore error returned", func(t *testing.T) {
		expectedErr := errors.New("restore failed")
		repo := &mockBackupRepo{restoreErr: expectedErr}
		bw := NewBackupWorker(true, 0, repo, zap.NewNop())

		ctx := context.Background()
		err := bw.Run(ctx)
		assert.ErrorIs(t, err, expectedErr)
	})

	t.Run("backup on ticker", func(t *testing.T) {
		repo := &mockBackupRepo{}
		bw := NewBackupWorker(false, 1, repo, zap.NewNop())

		ctx, cancel := context.WithCancel(context.Background())

		go func() {
			time.Sleep(1500 * time.Millisecond)
			cancel()
		}()

		err := bw.Run(ctx)
		assert.ErrorIs(t, err, context.Canceled)
		assert.GreaterOrEqual(t, repo.backupCalls, 1)
	})

	t.Run("backup error is logged, not returned", func(t *testing.T) {
		repo := &mockBackupRepo{backupErr: errors.New("backup failed")}
		bw := NewBackupWorker(false, 1, repo, zap.NewNop())

		ctx, cancel := context.WithCancel(context.Background())

		go func() {
			time.Sleep(1500 * time.Millisecond)
			cancel()
		}()

		err := bw.Run(ctx)
		assert.ErrorIs(t, err, context.Canceled)
		assert.GreaterOrEqual(t, repo.backupCalls, 1)
	})

	t.Run("disabled when interval is 0", func(t *testing.T) {
		repo := &mockBackupRepo{}
		bw := NewBackupWorker(false, 0, repo, zap.NewNop())

		ctx, cancel := context.WithTimeout(context.Background(), 50*time.Millisecond)
		defer cancel()

		err := bw.Run(ctx)
		assert.NoError(t, err)
		assert.Equal(t, 0, repo.backupCalls)
	})
}

func TestBackupWorker_New(t *testing.T) {
	t.Run("creates worker with correct fields", func(t *testing.T) {
		repo := &mockBackupRepo{}
		bw := NewBackupWorker(true, 10, repo, zap.NewNop())
		assert.NotNil(t, bw)
		assert.True(t, bw.restore)
		assert.Equal(t, 10, bw.storeInterval)
	})
}
