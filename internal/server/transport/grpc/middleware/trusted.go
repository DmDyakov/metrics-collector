// Package middleware содержит gRPC-интерцепторы.
package middleware

import (
	"context"
	"net"

	"go.uber.org/zap"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
)

// TrustedSubnetInterceptor проверяет, что IP агента принадлежит доверенной подсети.
func TrustedSubnetInterceptor(cidr string, logger *zap.Logger) grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {
		if cidr == "" {
			return handler(ctx, req)
		}

		md, ok := metadata.FromIncomingContext(ctx)
		if !ok {
			logger.Warn("missing metadata in gRPC request")
			return nil, status.Error(codes.PermissionDenied, "missing metadata")
		}

		ips := md.Get("x-real-ip")
		if len(ips) == 0 {
			logger.Warn("missing x-real-ip in gRPC metadata")
			return nil, status.Error(codes.PermissionDenied, "missing x-real-ip")
		}

		clientIP := net.ParseIP(ips[0])
		if clientIP == nil {
			logger.Warn("invalid IP in x-real-ip", zap.String("ip", ips[0]))
			return nil, status.Error(codes.PermissionDenied, "invalid x-real-ip")
		}

		_, subnet, err := net.ParseCIDR(cidr)
		if err != nil {
			logger.Error("invalid trusted subnet", zap.String("subnet", cidr), zap.Error(err))
			return nil, status.Error(codes.Internal, "invalid trusted subnet configuration")
		}

		if !subnet.Contains(clientIP) {
			logger.Warn("IP not in trusted subnet",
				zap.String("client_ip", clientIP.String()),
				zap.String("trusted_subnet", cidr),
			)
			return nil, status.Error(codes.PermissionDenied, "ip not in trusted subnet")
		}

		return handler(ctx, req)
	}
}
