// Package grpc реализует gRPC-клиент для отправки метрик на сервер.
package grpc

import (
	"context"
	"fmt"
	"time"

	"metrics-collector/internal/config"
	pb "metrics-collector/internal/proto"

	"go.uber.org/zap"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/metadata"
)

//go:generate mockgen -destination=mocks/mock_metrics_client.go -package=mocks metrics-collector/internal/proto MetricsClient

// Client — gRPC-клиент для отправки метрик.
type Client struct {
	conn   *grpc.ClientConn
	client pb.MetricsClient
	cfg    *config.AgentConfig
	logger *zap.Logger
}

// New создаёт gRPC-клиент.
func New(cfg *config.AgentConfig, logger *zap.Logger) (*Client, error) {
	addr := cfg.GRPCAddress
	if addr == "" {
		return nil, fmt.Errorf("gRPC address is empty")
	}

	conn, err := grpc.NewClient(addr, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		return nil, fmt.Errorf("failed to create grpc connection: %w", err)
	}

	return &Client{
		conn:   conn,
		client: pb.NewMetricsClient(conn),
		cfg:    cfg,
		logger: logger,
	}, nil
}

// Close закрывает gRPC соединение.
func (c *Client) Close() error {
	if c.conn != nil {
		return c.conn.Close()
	}
	return nil
}

// SendMetrics отправляет метрики на сервер по gRPC.
func (c *Client) SendMetrics(ctx context.Context, batch map[string]float64) error {
	if len(batch) == 0 {
		c.logger.Warn("No metrics for gRPC send")
		return nil
	}

	c.logger.Info("Sending metrics via gRPC")
	start := time.Now()

	metrics := toProto(batch)

	req := pb.UpdateMetricsRequest_builder{
		Metrics: metrics,
	}.Build()

	md := metadata.Pairs("x-real-ip", c.cfg.AgentIP)
	ctx = metadata.NewOutgoingContext(ctx, md)

	_, err := c.client.UpdateMetrics(ctx, req)
	if err != nil {
		c.logger.Error("gRPC UpdateMetrics failed", zap.Error(err))
		return err
	}

	c.logger.Info("gRPC request sent",
		zap.Int("metrics_count", len(metrics)),
		zap.Duration("duration", time.Since(start)),
	)

	return nil
}
