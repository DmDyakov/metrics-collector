package grpc

import (
	"context"
	"testing"

	"metrics-collector/internal/agent/clients/grpc/mocks"
	"metrics-collector/internal/config"
	pb "metrics-collector/internal/proto"

	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"
	"go.uber.org/zap"
	"google.golang.org/grpc/metadata"
)

func TestNew_EmptyAddress(t *testing.T) {
	cfg := &config.AgentConfig{GRPCAddress: ""}
	_, err := New(cfg, zap.NewNop())
	assert.Error(t, err)
}

func TestSendMetrics_EmptyBatch(t *testing.T) {
	cfg := &config.AgentConfig{AgentIP: "192.168.1.1"}
	c := &Client{cfg: cfg, logger: zap.NewNop()}

	err := c.SendMetrics(context.Background(), map[string]float64{})
	assert.NoError(t, err)
}

func TestSendMetrics_Success(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockClient := mocks.NewMockMetricsClient(ctrl)
	mockClient.EXPECT().
		UpdateMetrics(gomock.Any(), gomock.Any(), gomock.Any()).
		Return(&pb.UpdateMetricsResponse{}, nil)

	c := &Client{
		client: mockClient,
		cfg:    &config.AgentConfig{AgentIP: "192.168.1.1"},
		logger: zap.NewNop(),
	}

	batch := map[string]float64{"cpu_usage": 42.5}
	err := c.SendMetrics(context.Background(), batch)
	assert.NoError(t, err)
}

func TestSendMetrics_ServerError(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockClient := mocks.NewMockMetricsClient(ctrl)
	mockClient.EXPECT().
		UpdateMetrics(gomock.Any(), gomock.Any(), gomock.Any()).
		Return(nil, assert.AnError)

	c := &Client{
		client: mockClient,
		cfg:    &config.AgentConfig{AgentIP: "192.168.1.1"},
		logger: zap.NewNop(),
	}

	err := c.SendMetrics(context.Background(), map[string]float64{"cpu_usage": 42.5})
	assert.Error(t, err)
}

func TestSendMetrics_XRealIP(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	expectedIP := "10.0.0.1"

	mockClient := mocks.NewMockMetricsClient(ctrl)
	mockClient.EXPECT().
		UpdateMetrics(gomock.Any(), gomock.Any(), gomock.Any()).
		DoAndReturn(func(ctx context.Context, _ *pb.UpdateMetricsRequest, _ ...interface{}) (*pb.UpdateMetricsResponse, error) {
			md, ok := metadata.FromOutgoingContext(ctx)
			assert.True(t, ok)
			ips := md.Get("x-real-ip")
			assert.Equal(t, expectedIP, ips[0])
			return &pb.UpdateMetricsResponse{}, nil
		})

	c := &Client{
		client: mockClient,
		cfg:    &config.AgentConfig{AgentIP: expectedIP},
		logger: zap.NewNop(),
	}

	err := c.SendMetrics(context.Background(), map[string]float64{"cpu_usage": 42.5})
	assert.NoError(t, err)
}
