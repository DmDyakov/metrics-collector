package worker

import "context"

//go:generate mockgen -destination=mocks/mock_store.go -package=mocks . Store
type Store interface {
	UpdateMetrics(metrics map[string]float64)
	GetMetricsSnapshot() map[string]float64
}

//go:generate mockgen -destination=mocks/mock_client.go -package=mocks . Client
type Client interface {
	SendMetrics(ctx context.Context, metrics map[string]float64) error
}
