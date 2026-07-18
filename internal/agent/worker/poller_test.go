package worker

import (
	"testing"

	"metrics-collector/internal/agent/worker/mocks"

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

func setupPollerTest(t *testing.T) (*Poller, *mocks.MockPollerStore) {
	t.Helper()

	ctrl := gomock.NewController(t)
	t.Cleanup(ctrl.Finish)

	mockStore := mocks.NewMockPollerStore(ctrl)
	logger := zap.NewNop()
	poller := NewPoller(mockStore, logger, 1)

	return poller, mockStore
}
