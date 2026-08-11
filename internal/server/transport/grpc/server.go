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

type Server struct {
	*grpc.Server
	addr   string
	logger *zap.Logger
}

// NewServer создаёт gRPC-сервер и регистрирует обработчик
func NewServer(addr string, cidr string, handler *Handler, logger *zap.Logger) (*Server, error) {

	trustedSubnet, err := middleware.ParseCIDR(cidr)
	if err != nil {
		return nil, fmt.Errorf("gRPC server failed to parse cidr: %w", err)
	}

	s := grpc.NewServer(
		grpc.UnaryInterceptor(middleware.TrustedSubnetInterceptor(trustedSubnet, logger)),
	)
	pb.RegisterMetricsServer(s, handler)

	return &Server{
		Server: s,
		addr:   addr,
		logger: logger,
	}, nil
}

// Run запускает сервер.
func (s *Server) Run(ctx context.Context) error {
	if s == nil {
		return nil
	}

	listener, err := net.Listen("tcp", s.addr)
	if err != nil {
		return fmt.Errorf("gRPC server failed to listen: %w", err)
	}

	s.logger.Info("gRPC server started", zap.String("addr", s.addr))

	go func() {
		<-ctx.Done()
		s.logger.Info("Shutting down gRPC server...")
		s.GracefulStop()
	}()

	if err := s.Serve(listener); err != nil && err != grpc.ErrServerStopped {
		return fmt.Errorf("gRPC server error: %w", err)
	}

	s.logger.Info("gRPC server stopped gracefully")
	return nil
}
