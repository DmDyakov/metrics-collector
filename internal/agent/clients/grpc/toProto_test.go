package grpc

import (
	pb "metrics-collector/internal/proto"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestToProto(t *testing.T) {
	input := map[string]float64{
		"PollCount": 10,
		"cpu_usage": 42.5,
		"mem_usage": 68.0,
	}

	result := toProto(input)

	assert.Len(t, result, 3)

	assert.Equal(t, "PollCount", result[0].GetId())
	assert.Equal(t, pb.Metric_COUNTER, result[0].GetType())
	assert.Equal(t, int64(10), result[0].GetDelta())

	assert.Equal(t, "cpu_usage", result[1].GetId())
	assert.Equal(t, pb.Metric_GAUGE, result[1].GetType())
	assert.Equal(t, 42.5, result[1].GetValue())
}
