package middleware

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"go.uber.org/zap"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
)

func TestTrustedSubnetInterceptor_EmptyCIDR(t *testing.T) {
	interceptor := TrustedSubnetInterceptor("", zap.NewNop())

	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return "ok", nil
	}

	resp, err := interceptor(context.Background(), nil, &grpc.UnaryServerInfo{FullMethod: "/test"}, handler)
	assert.NoError(t, err)
	assert.Equal(t, "ok", resp)
}

func TestTrustedSubnetInterceptor_MissingMetadata(t *testing.T) {
	interceptor := TrustedSubnetInterceptor("192.168.1.0/24", zap.NewNop())

	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return "ok", nil
	}

	resp, err := interceptor(context.Background(), nil, &grpc.UnaryServerInfo{FullMethod: "/test"}, handler)
	assert.Nil(t, resp)
	assert.Error(t, err)

	st, _ := status.FromError(err)
	assert.Equal(t, codes.PermissionDenied, st.Code())
}

func TestTrustedSubnetInterceptor_ValidIP(t *testing.T) {
	interceptor := TrustedSubnetInterceptor("192.168.1.0/24", zap.NewNop())

	md := metadata.Pairs("x-real-ip", "192.168.1.5")
	ctx := metadata.NewIncomingContext(context.Background(), md)

	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return "ok", nil
	}

	resp, err := interceptor(ctx, nil, &grpc.UnaryServerInfo{FullMethod: "/test"}, handler)
	assert.NoError(t, err)
	assert.Equal(t, "ok", resp)
}

func TestTrustedSubnetInterceptor_InvalidIP(t *testing.T) {
	interceptor := TrustedSubnetInterceptor("192.168.1.0/24", zap.NewNop())

	md := metadata.Pairs("x-real-ip", "10.0.0.1")
	ctx := metadata.NewIncomingContext(context.Background(), md)

	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return "ok", nil
	}

	resp, err := interceptor(ctx, nil, &grpc.UnaryServerInfo{FullMethod: "/test"}, handler)
	assert.Nil(t, resp)

	st, _ := status.FromError(err)
	assert.Equal(t, codes.PermissionDenied, st.Code())
}

func TestTrustedSubnetInterceptor_InvalidCIDR(t *testing.T) {
	interceptor := TrustedSubnetInterceptor("invalid", zap.NewNop())

	md := metadata.Pairs("x-real-ip", "192.168.1.5")
	ctx := metadata.NewIncomingContext(context.Background(), md)

	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return "ok", nil
	}

	resp, err := interceptor(ctx, nil, &grpc.UnaryServerInfo{FullMethod: "/test"}, handler)
	assert.Nil(t, resp)

	st, _ := status.FromError(err)
	assert.Equal(t, codes.Internal, st.Code())
}
