package grpc

import (
	"fmt"
	"metrics-collector/internal/domain/metrics"
	pb "metrics-collector/internal/proto"
	models "metrics-collector/internal/server/model"
)

// fromProto конвертирует proto в модель
func fromProto(m *pb.Metric) (models.Metrics, error) {
	var mType metrics.Type
	switch m.GetType() {
	case pb.Metric_GAUGE:
		mType = metrics.Gauge
	case pb.Metric_COUNTER:
		mType = metrics.Counter
	default:
		return models.Metrics{}, fmt.Errorf("unsupported metric type: %v", m.GetType())
	}

	metric := models.Metrics{
		ID:    m.GetId(),
		MType: mType,
	}

	if m.GetType() == pb.Metric_COUNTER {
		delta := m.GetDelta()
		metric.Delta = &delta
	}

	if m.GetType() == pb.Metric_GAUGE {
		value := m.GetValue()
		metric.Value = &value
	}

	return metric, nil
}
