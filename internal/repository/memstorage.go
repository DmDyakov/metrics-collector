package repository

type MemStorage struct {
	gauges   map[string]float64
	counters map[string]int64
}

func (m *MemStorage) UpdateCounter(name string, v int64) {
	m.counters[name] += v
}

func (m *MemStorage) UpdateGauge(name string, v float64) {
	m.gauges[name] = v
}

func NewMemStorage() *MemStorage {
	return &MemStorage{
		gauges:   make(map[string]float64),
		counters: make(map[string]int64),
	}
}
