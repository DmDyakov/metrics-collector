package grpc

import (
	"metrics-collector/internal/domain/metrics"
	pb "metrics-collector/internal/proto"
)

// toProto конвертирует map[string]float64 в слайс pb.Metric для gRPC.
func toProto(batch map[string]float64) []*pb.Metric {
	result := make([]*pb.Metric, 0, len(batch))

	for k, v := range batch {
		if k == metrics.PollCount {
			delta := int64(v)
			result = append(result, pb.Metric_builder{
				Id:    k,
				Type:  pb.Metric_COUNTER,
				Delta: delta,
			}.Build())
		} else {
			result = append(result, pb.Metric_builder{
				Id:    k,
				Type:  pb.Metric_GAUGE,
				Value: v,
			}.Build())
		}
	}

	return result
}
