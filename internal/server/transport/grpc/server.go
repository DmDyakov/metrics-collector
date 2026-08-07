// Package grpc реализует gRPC-сервер для приёма метрик.
package grpc

import (
	"context"
	"fmt"
	"net"

	pb "metrics-collector/internal/proto"
	"metrics-collector/internal/server/transport/grpc/middleware"

	"go.uber.org/zap"
	"google.golang.org/grpc"
)

// New создаёт gRPC-сервер.
func New(addr string, trustedSubnet string, handler *Handler, logger *zap.Logger) (*grpc.Server, error) {
	listener, err := net.Listen("tcp", addr)
	if err != nil {
		return nil, fmt.Errorf("failed to listen on %s: %w", addr, err)
	}

	s := grpc.NewServer(
		grpc.UnaryInterceptor(middleware.TrustedSubnetInterceptor(trustedSubnet, logger)),
	)

	pb.RegisterMetricsServer(s, handler)

	go func() {
		if err := s.Serve(listener); err != nil {
			logger.Error("gRPC server failed", zap.Error(err))
		}
	}()

	logger.Info("gRPC server started", zap.String("addr", addr))

	return s, nil
}

// Run ждёт сигнала и глушит сервер.
func Run(ctx context.Context, s *grpc.Server, logger *zap.Logger) error {
	<-ctx.Done()
	logger.Info("Shutting down gRPC server...")
	s.GracefulStop()
	return nil
}
