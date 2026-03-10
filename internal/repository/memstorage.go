package repository

type MemStorage struct {
	gauges   map[string]float64
	counters map[string]int64
}

func (m MemStorage) GetAllMetricsRaw() (gauges map[string]float64, counters map[string]int64) {
	return m.gauges, m.counters
}

func (m *MemStorage) GetGaugeMetricValue(metricName string) (float64, bool) {
	value, ok := m.gauges[metricName]
	return value, ok
}

func (m *MemStorage) GetCountMetricValue(metricName string) (int64, bool) {
	value, ok := m.counters[metricName]
	return value, ok
}

func (m *MemStorage) UpdateCounterMetric(metricName string, v int64) {
	m.counters[metricName] += v
}

func (m *MemStorage) UpdateGaugeMetric(metricName string, v float64) {
	m.gauges[metricName] = v
}

func NewMemStorage() *MemStorage {
	return &MemStorage{
		gauges:   make(map[string]float64),
		counters: make(map[string]int64),
	}
}
