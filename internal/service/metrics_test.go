package service

import (
	"testing"

	models "metrics-collector/internal/model"
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

func (m *mockRepository) GetMetric(name string) (*models.Metrics, bool) {
	metric, ok := m.data[name]
	return &metric, ok
}

func (m *mockRepository) UpdateMetric(metric models.Metrics) *models.Metrics {
	m.data[metric.ID] = metric
	return &metric
}

func TestService_UpdateMetricV2(t *testing.T) {
	gaugeVal := 10.5
	counterVal := int64(5)

	t.Run("update gauge", func(t *testing.T) {
		repo := newMockRepository()
		svc := NewMetricsService(repo)

		metric := models.Metrics{
			ID:    "Alloc",
			MType: models.Gauge,
			Value: &gaugeVal,
		}

		result, err := svc.UpdateMetricV2(metric)

		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if result == nil {
			t.Fatal("expected result, got nil")
		}

		saved, ok := repo.data["Alloc"]
		if !ok {
			t.Fatal("metric not saved")
		}

		if saved.Value == nil {
			t.Fatal("value is nil")
		}

		if *saved.Value != gaugeVal {
			t.Errorf("expected %f, got %f", gaugeVal, *saved.Value)
		}
	})

	t.Run("update counter", func(t *testing.T) {
		repo := newMockRepository()
		svc := NewMetricsService(repo)

		metric := models.Metrics{
			ID:    "PollCount",
			MType: models.Counter,
			Delta: &counterVal,
		}

		result, err := svc.UpdateMetricV2(metric)

		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if result == nil {
			t.Fatal("expected result, got nil")
		}

		saved, ok := repo.data["PollCount"]
		if !ok {
			t.Fatal("metric not saved")
		}

		if saved.Delta == nil {
			t.Fatal("delta is nil")
		}

		if *saved.Delta != counterVal {
			t.Errorf("expected %d, got %d", counterVal, *saved.Delta)
		}
	})

	t.Run("increment counter", func(t *testing.T) {
		repo := newMockRepository()

		initial := int64(5)
		repo.UpdateMetric(models.Metrics{
			ID:    "PollCount",
			MType: models.Counter,
			Delta: &initial,
		})

		svc := NewMetricsService(repo)

		increment := int64(3)
		metric := models.Metrics{
			ID:    "PollCount",
			MType: models.Counter,
			Delta: &increment,
		}

		result, err := svc.UpdateMetricV2(metric)

		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if result == nil {
			t.Fatal("expected result, got nil")
		}

		saved, ok := repo.data["PollCount"]
		if !ok {
			t.Fatal("metric not saved")
		}

		if saved.Delta == nil {
			t.Fatal("delta is nil")
		}

		expected := int64(8) // 5 + 3
		if *saved.Delta != expected {
			t.Errorf("expected %d, got %d", expected, *saved.Delta)
		}
	})
}
