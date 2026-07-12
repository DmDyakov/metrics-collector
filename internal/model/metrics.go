package models

const (
	Counter = "counter"
	Gauge   = "gauge"
)

const (
	PollCount   = "PoolCount"
	RandomValue = "RandomValue"
)

// generate:reset
type Metrics struct {
	ID    string   `json:"id"`
	MType string   `json:"type"`
	Delta *int64   `json:"delta,omitempty"`
	Value *float64 `json:"value,omitempty"`
	Hash  string   `json:"hash,omitempty"`
}
