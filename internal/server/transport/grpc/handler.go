package grpc

import (
	"context"

	"metrics-collector/internal/domain/metrics"
	pb "metrics-collector/internal/proto"
	models "metrics-collector/internal/server/model"

	"go.uber.org/zap"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

//go:generate mockgen -destination=mocks/mock_metrics_service.go -package=mocks metrics-collector/internal/server/transport/grpc MetricsService

// MetricsService — интерфейс бизнес-логики.
type MetricsService interface {
	UpdateMetrics(ctx context.Context, batch []models.Metrics) (*int, error)
}

// Handler реализует pb.MetricsServer.
type Handler struct {
	pb.UnimplementedMetricsServer
	service MetricsService
	logger  *zap.Logger
}

// NewHandler создаёт обработчик.
func NewHandler(service MetricsService, logger *zap.Logger) *Handler {
	return &Handler{service: service, logger: logger}
}

// UpdateMetrics — реализация gRPC-метода.
func (h *Handler) UpdateMetrics(ctx context.Context, req *pb.UpdateMetricsRequest) (*pb.UpdateMetricsResponse, error) {
	batch := make([]models.Metrics, 0, len(req.GetMetrics()))
	for _, m := range req.GetMetrics() {
		metric := models.Metrics{
			ID:    m.GetId(),
			MType: metrics.Type(m.GetType().String()),
		}
		if m.GetType() == pb.Metric_COUNTER {
			delta := m.GetDelta()
			metric.Delta = &delta
		} else {
			value := m.GetValue()
			metric.Value = &value
		}
		batch = append(batch, metric)
	}

	_, err := h.service.UpdateMetrics(ctx, batch)
	if err != nil {
		h.logger.Error("failed to update metrics", zap.Error(err))
		return nil, status.Errorf(codes.Internal, "failed to update metrics: %v", err)
	}

	return pb.UpdateMetricsResponse_builder{}.Build(), nil
}
