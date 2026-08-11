package worker

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"
	"go.uber.org/zap"

	models "metrics-collector/internal/server/model"
	"metrics-collector/internal/server/worker/mocks"
)

func TestBackupWorker_Run(t *testing.T) {
	t.Run("restore on start", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		mem := mocks.NewMockMemStorage(ctrl)
		persistent := mocks.NewMockPersistentStorage(ctrl)

		metrics := []models.Metrics{{ID: "test", MType: "gauge"}}

		persistent.EXPECT().GetAll(gomock.Any()).Return(metrics, nil)
		mem.EXPECT().SaveBatch(metrics).Return(nil)

		bw := NewBackupWorker(true, 0, 5*time.Second, mem, persistent, zap.NewNop())

		ctx, cancel := context.WithCancel(context.Background())
		cancel()

		err := bw.Run(ctx)
		assert.NoError(t, err)
	})

	t.Run("restore error returned", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		mem := mocks.NewMockMemStorage(ctrl)
		persistent := mocks.NewMockPersistentStorage(ctrl)

		expectedErr := errors.New("restore failed")
		persistent.EXPECT().GetAll(gomock.Any()).Return(nil, expectedErr)

		bw := NewBackupWorker(true, 0, 5*time.Second, mem, persistent, zap.NewNop())

		ctx := context.Background()
		err := bw.Run(ctx)
		assert.ErrorIs(t, err, expectedErr)
	})

	t.Run("backup on ticker", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		mem := mocks.NewMockMemStorage(ctrl)
		persistent := mocks.NewMockPersistentStorage(ctrl)

		mem.EXPECT().GetAll().Return(map[string]models.Metrics{
			"test": {ID: "test", MType: "gauge"},
		}).MinTimes(1)
		persistent.EXPECT().SaveBatch(gomock.Any(), gomock.Any()).Return(nil, nil).MinTimes(1)

		bw := NewBackupWorker(false, 1, 5*time.Second, mem, persistent, zap.NewNop())

		ctx, cancel := context.WithCancel(context.Background())

		go func() {
			time.Sleep(1500 * time.Millisecond)
			cancel()
		}()

		err := bw.Run(ctx)
		assert.NoError(t, err)
	})

	t.Run("disabled when interval is 0", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		mem := mocks.NewMockMemStorage(ctrl)
		persistent := mocks.NewMockPersistentStorage(ctrl)

		mem.EXPECT().GetAll().Times(0)
		persistent.EXPECT().SaveBatch(gomock.Any(), gomock.Any()).Times(0)

		bw := NewBackupWorker(false, 0, 5*time.Second, mem, persistent, zap.NewNop())

		ctx, cancel := context.WithTimeout(context.Background(), 50*time.Millisecond)
		defer cancel()

		err := bw.Run(ctx)
		assert.NoError(t, err)
	})
}

func TestBackupWorker_New(t *testing.T) {
	t.Run("creates worker with correct fields", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		mem := mocks.NewMockMemStorage(ctrl)
		persistent := mocks.NewMockPersistentStorage(ctrl)
		timeout := 5 * time.Second

		bw := NewBackupWorker(true, 10, timeout, mem, persistent, zap.NewNop())

		assert.NotNil(t, bw)
		assert.True(t, bw.restore)
		assert.Equal(t, 10, bw.interval)
		assert.Equal(t, timeout, bw.timeout)
	})
}
