package worker

import (
	"context"
	"testing"

	"metrics-collector/internal/agent/worker/mocks"

	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"
	"go.uber.org/zap"
)

func TestReporter_WorkerStopsOnContextCancel(t *testing.T) {
	t.Run("worker exits when context cancelled", func(t *testing.T) {
		reporter, _, _ := setupReporterTest(t)

		ctx, cancel := context.WithCancel(context.Background())
		cancel()

		err := reporter.worker(ctx, 1)
		assert.NoError(t, err)
	})
}

func TestReporter_WorkerStopsOnClosedChannel(t *testing.T) {
	t.Run("worker exits when jobs channel closed", func(t *testing.T) {
		reporter, _, _ := setupReporterTest(t)

		close(reporter.jobs)

		err := reporter.worker(context.Background(), 1)
		assert.NoError(t, err)
	})
}

func setupReporterTest(t *testing.T) (*Reporter, *mocks.MockReporterStore, *mocks.MockClient) {
	t.Helper()

	ctrl := gomock.NewController(t)
	t.Cleanup(ctrl.Finish)

	mockStore := mocks.NewMockReporterStore(ctrl)
	mockClient := mocks.NewMockClient(ctrl)
	logger := zap.NewNop()
	reporter := NewReporter(mockStore, mockClient, logger, 2, 1)

	return reporter, mockStore, mockClient
}
