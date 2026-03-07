package service

import "testing"

type mockRepository struct {
	counterCalls []struct {
		Name  string
		Value int64
	}
	gaugeCalls []struct {
		Name  string
		Value float64
	}
}

func (m *mockRepository) UpdateCounter(name string, value int64) {
	m.counterCalls = append(m.counterCalls, struct {
		Name  string
		Value int64
	}{name, value})
}

func (m *mockRepository) UpdateGauge(name string, value float64) {
	m.gaugeCalls = append(m.gaugeCalls, struct {
		Name  string
		Value float64
	}{name, value})
}

func TestMetricsService_Update(t *testing.T) {

	tests := []struct {
		name           string
		metricType     string
		metricName     string
		metricValueRaw string
		wantErr        error
		checkMock      func(*testing.T, *mockRepository)
	}{
		{
			name:           "valid counter metric",
			metricType:     "counter",
			metricName:     "test_counter",
			metricValueRaw: "123",
			wantErr:        nil,
			checkMock: func(t *testing.T, m *mockRepository) {
				if len(m.counterCalls) != 1 {
					t.Errorf("expected 1 counter call, got %d", len(m.counterCalls))
				}
				if len(m.gaugeCalls) != 0 {
					t.Errorf("expected 0 gauge calls, got %d", len(m.gaugeCalls))
				}

				if len(m.counterCalls) > 0 {
					if m.counterCalls[0].Name != "test_counter" {
						t.Errorf("counter name = %q, want %q", m.counterCalls[0].Name, "test_counter")
					}
					if m.counterCalls[0].Value != 123 {
						t.Errorf("counter value = %d, want %d", m.counterCalls[0].Value, 123)
					}
				}
			},
		},
		{
			name:           "unknown metric type",
			metricType:     "test",
			metricName:     "test_counter",
			metricValueRaw: "123",
			wantErr:        ErrUnknownMetricType,
			checkMock: func(t *testing.T, m *mockRepository) {
				if len(m.gaugeCalls) != 0 {
					t.Errorf("expected 0 gauge calls, got %d", len(m.gaugeCalls))
				}
				if len(m.counterCalls) != 0 {
					t.Errorf("expected 0 counter calls, got %d", len(m.counterCalls))
				}
			},
		},
		{
			name:           "invalid counter value",
			metricType:     "counter",
			metricName:     "test_counter",
			metricValueRaw: "123.5",
			wantErr:        ErrInvalidCounterValue,
			checkMock: func(t *testing.T, m *mockRepository) {
				if len(m.gaugeCalls) != 0 {
					t.Errorf("expected 0 gauge calls, got %d", len(m.gaugeCalls))
				}
				if len(m.counterCalls) != 0 {
					t.Errorf("expected 0 counter calls, got %d", len(m.counterCalls))
				}
			},
		},
		{
			name:           "invalid gauge value",
			metricType:     "gauge",
			metricName:     "test_gauge",
			metricValueRaw: "abc",
			wantErr:        ErrInvalidGaugeValue,
			checkMock: func(t *testing.T, m *mockRepository) {
				if len(m.gaugeCalls) != 0 {
					t.Errorf("expected 0 gauge calls, got %d", len(m.gaugeCalls))
				}
				if len(m.counterCalls) != 0 {
					t.Errorf("expected 0 counter calls, got %d", len(m.counterCalls))
				}
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockRepo := &mockRepository{}
			svc := NewMetricsService(mockRepo)
			err := svc.Update(tt.metricType, tt.metricName, tt.metricValueRaw)
			if err != tt.wantErr {
				t.Errorf("Update() error = %v, wantErr %v", err, tt.wantErr)
				t.Logf("Received: type=%s, name=%s, value=%s",
					tt.metricType, tt.metricName, tt.metricValueRaw)
			}

			if tt.checkMock != nil {
				tt.checkMock(t, mockRepo)
			}
		})
	}
}
