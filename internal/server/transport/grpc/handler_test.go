package grpc

import (
	"context"
	"errors"
	"testing"

	pb "metrics-collector/internal/proto"
	models "metrics-collector/internal/server/model"
	"metrics-collector/internal/server/transport/grpc/mocks"

	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"
	"go.uber.org/zap"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func TestHandler_UpdateMetrics_Success(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockSvc := mocks.NewMockMetricsService(ctrl)
	mockSvc.EXPECT().
		UpdateMetrics(gomock.Any(), gomock.Any()).
		Return(intPtr(3), nil)

	h := NewHandler(mockSvc, zap.NewNop())

	req := pb.UpdateMetricsRequest_builder{
		Metrics: []*pb.Metric{
			pb.Metric_builder{Id: "cpu", Type: pb.Metric_GAUGE, Value: 42.5}.Build(),
			pb.Metric_builder{Id: "hits", Type: pb.Metric_COUNTER, Delta: 10}.Build(),
		},
	}.Build()

	resp, err := h.UpdateMetrics(context.Background(), req)

	assert.NoError(t, err)
	assert.NotNil(t, resp)
}

func TestHandler_UpdateMetrics_ServiceError(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockSvc := mocks.NewMockMetricsService(ctrl)
	mockSvc.EXPECT().
		UpdateMetrics(gomock.Any(), gomock.Any()).
		Return(nil, errors.New("db error"))

	h := NewHandler(mockSvc, zap.NewNop())

	req := pb.UpdateMetricsRequest_builder{
		Metrics: []*pb.Metric{
			pb.Metric_builder{Id: "cpu", Type: pb.Metric_GAUGE, Value: 42.5}.Build(),
		},
	}.Build()

	resp, err := h.UpdateMetrics(context.Background(), req)

	assert.Nil(t, resp)
	assert.Error(t, err)

	st, ok := status.FromError(err)
	assert.True(t, ok)
	assert.Equal(t, codes.Internal, st.Code())
}

func TestHandler_UpdateMetrics_CounterConversion(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockSvc := mocks.NewMockMetricsService(ctrl)
	mockSvc.EXPECT().
		UpdateMetrics(gomock.Any(), gomock.Any()).
		DoAndReturn(func(_ context.Context, batch []models.Metrics) (*int, error) {
			assert.Equal(t, "hits", batch[0].ID)
			assert.Equal(t, int64(10), *batch[0].Delta)
			return intPtr(1), nil
		})

	h := NewHandler(mockSvc, zap.NewNop())

	req := pb.UpdateMetricsRequest_builder{
		Metrics: []*pb.Metric{
			pb.Metric_builder{Id: "hits", Type: pb.Metric_COUNTER, Delta: 10}.Build(),
		},
	}.Build()

	_, err := h.UpdateMetrics(context.Background(), req)
	assert.NoError(t, err)
}

func intPtr(i int) *int {
	return &i
}
