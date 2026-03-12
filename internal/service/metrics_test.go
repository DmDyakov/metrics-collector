package service

import (
	models "metrics-collector/internal/model"
	"testing"
)

type mockRepository struct {
	data map[string]models.Metrics
}

func newMockRepository() *mockRepository {
	return &mockRepository{
		data: make(map[string]models.Metrics),
	}
}

func (m *mockRepository) GetAllMetrics() map[string]models.Metrics {
	return m.data
}

func (m *mockRepository) GetMetric(name string) (models.Metrics, bool) {
	metric, ok := m.data[name]
	return metric, ok
}

func (m *mockRepository) UpdateMetric(metric models.Metrics) {
	m.data[metric.ID] = metric
}

func TestUpdateMetric(t *testing.T) {
	tests := []struct {
		name      string
		mType     string
		mName     string
		mValue    string
		expectErr error
		expectKey string
	}{
		{
			name:      "valid gauge",
			mType:     models.Gauge,
			mName:     "Alloc",
			mValue:    "10.5",
			expectErr: nil,
			expectKey: "Alloc",
		},
		{
			name:      "valid counter",
			mType:     models.Counter,
			mName:     "PollCount",
			mValue:    "5",
			expectErr: nil,
			expectKey: "PollCount",
		},
		{
			name:      "invalid counter value",
			mType:     models.Counter,
			mName:     "PollCount",
			mValue:    "abc",
			expectErr: ErrInvalidCounterValue,
		},
		{
			name:      "invalid gauge value",
			mType:     models.Gauge,
			mName:     "Alloc",
			mValue:    "abc",
			expectErr: ErrInvalidGaugeValue,
		},
		{
			name:      "unknown type",
			mType:     "unknown",
			mName:     "Metric",
			mValue:    "1",
			expectErr: ErrUnknownMetricType,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			repo := newMockRepository()
			svc := NewMetricsService(repo)

			err := svc.UpdateMetric(tt.mType, tt.mName, tt.mValue)

			if err != tt.expectErr {
				t.Fatalf("expected error %v, got %v", tt.expectErr, err)
			}

			if tt.expectErr == nil {
				if _, ok := repo.data[tt.expectKey]; !ok {
					t.Fatalf("metric not saved in repo")
				}
			}

		})
	}
}
