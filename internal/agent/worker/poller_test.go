package worker

import (
	"context"
	"testing"
	"time"

	"metrics-collector/internal/agent/worker/mocks"

	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"
	"go.uber.org/zap"
)

func TestPoller_PollMemStats(t *testing.T) {
	t.Run("updates store with mem stats", func(t *testing.T) {
		poller, mockStore := setupPollerTest(t)

		mockStore.EXPECT().
			UpdateMetrics(gomock.Any()).
			Times(1)

		poller.pollMemStats(1)
	})
}

func TestPoller_PollVirtualMemoryInfo(t *testing.T) {
	t.Run("updates store with memory info", func(t *testing.T) {
		poller, mockStore := setupPollerTest(t)

		mockStore.EXPECT().
			UpdateMetrics(gomock.Any()).
			Times(1)

		poller.pollVirtualMemoryInfo()
	})
}

func TestPoller_PollCPUPercentsInfo(t *testing.T) {
	t.Run("updates store with cpu percents", func(t *testing.T) {
		poller, mockStore := setupPollerTest(t)

		mockStore.EXPECT().
			UpdateMetrics(gomock.Any()).
			Times(1)

		poller.pollCPUPercentsInfo()
	})
}

func TestPoller_Run(t *testing.T) {
	t.Run("stops on context cancel", func(t *testing.T) {
		poller, mockStore := setupPollerTest(t)

		mockStore.EXPECT().
			UpdateMetrics(gomock.Any()).
			AnyTimes()

		// Используем канал для синхронизации
		ctx, cancel := context.WithCancel(context.Background())

		go func() {
			time.Sleep(500 * time.Millisecond)
			cancel()
		}()

		err := poller.Run(ctx)
		assert.ErrorIs(t, err, context.Canceled)
	})
}

func setupPollerTest(t *testing.T) (*Poller, *mocks.MockPollerStore) {
	t.Helper()

	ctrl := gomock.NewController(t)
	t.Cleanup(ctrl.Finish)

	mockStore := mocks.NewMockPollerStore(ctrl)
	logger := zap.NewNop()
	poller := NewPoller(mockStore, logger, 1)

	return poller, mockStore
}
