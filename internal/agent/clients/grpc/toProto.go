package grpc

import (
	pb "metrics-collector/internal/proto"
)

const pollCount = "PollCount"

// toProto конвертирует map[string]float64 в слайс pb.Metric для gRPC.
func toProto(metrics map[string]float64) []*pb.Metric {
	result := make([]*pb.Metric, 0, len(metrics))

	for k, v := range metrics {
		if k == pollCount {
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
