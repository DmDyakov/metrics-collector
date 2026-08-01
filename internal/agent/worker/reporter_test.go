package worker

import (
	"context"
	"errors"
	"testing"
	"time"

	"metrics-collector/internal/agent/worker/mocks"

	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"
	"go.uber.org/zap"
)

func TestReporter_Run(t *testing.T) {
	logger := zap.NewNop()

	t.Run("sends metrics on tick", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		mockClient := mocks.NewMockClient(ctrl)
		mockStore := mocks.NewMockReporterStore(ctrl)

		batch := map[string]float64{"cpu": 42.5}
		mockStore.EXPECT().GetMetricsSnapshot().Return(batch).AnyTimes()
		mockClient.EXPECT().SendMetrics(gomock.Any(), batch).Return(nil).AnyTimes()

		reporter := NewReporter(mockStore, mockClient, logger, 1, 100)

		ctx, cancel := context.WithCancel(context.Background())

		go func() {
			time.Sleep(250 * time.Millisecond)
			cancel()
		}()

		err := reporter.Run(ctx)
		assert.NoError(t, err)
	})

	t.Run("worker handles send error", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		mockClient := mocks.NewMockClient(ctrl)
		mockStore := mocks.NewMockReporterStore(ctrl)

		batch := map[string]float64{"cpu": 42.5}
		mockStore.EXPECT().GetMetricsSnapshot().Return(batch).AnyTimes()
		mockClient.EXPECT().SendMetrics(gomock.Any(), batch).Return(errors.New("send failed")).AnyTimes()

		reporter := NewReporter(mockStore, mockClient, logger, 1, 100)

		ctx, cancel := context.WithCancel(context.Background())

		go func() {
			time.Sleep(250 * time.Millisecond)
			cancel()
		}()

		err := reporter.Run(ctx)
		assert.NoError(t, err)
	})

	t.Run("scheduler stops on context cancel", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		mockClient := mocks.NewMockClient(ctrl)
		mockStore := mocks.NewMockReporterStore(ctrl)

		mockStore.EXPECT().GetMetricsSnapshot().Return(map[string]float64{}).AnyTimes()
		mockClient.EXPECT().SendMetrics(gomock.Any(), gomock.Any()).Return(nil).AnyTimes()

		reporter := NewReporter(mockStore, mockClient, logger, 2, 100)

		ctx, cancel := context.WithCancel(context.Background())

		go func() {
			time.Sleep(250 * time.Millisecond)
			cancel()
		}()

		err := reporter.Run(ctx)
		assert.NoError(t, err)
	})

	t.Run("multiple workers process jobs", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		mockClient := mocks.NewMockClient(ctrl)
		mockStore := mocks.NewMockReporterStore(ctrl)

		batch := map[string]float64{"cpu": 42.5}
		mockStore.EXPECT().GetMetricsSnapshot().Return(batch).AnyTimes()
		mockClient.EXPECT().SendMetrics(gomock.Any(), batch).Return(nil).AnyTimes()

		reporter := NewReporter(mockStore, mockClient, logger, 2, 100)

		ctx, cancel := context.WithCancel(context.Background())

		go func() {
			time.Sleep(250 * time.Millisecond)
			cancel()
		}()

		err := reporter.Run(ctx)
		assert.NoError(t, err)
	})
}
