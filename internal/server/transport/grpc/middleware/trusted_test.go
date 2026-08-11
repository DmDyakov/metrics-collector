package middleware

import (
	"context"
	"net"
	"testing"

	"github.com/stretchr/testify/assert"
	"go.uber.org/zap"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
)

func TestParseCIDR(t *testing.T) {
	t.Run("empty string returns nil", func(t *testing.T) {
		subnet, err := ParseCIDR("")
		assert.NoError(t, err)
		assert.Nil(t, subnet)
	})

	t.Run("valid CIDR", func(t *testing.T) {
		subnet, err := ParseCIDR("192.168.1.0/24")
		assert.NoError(t, err)
		assert.NotNil(t, subnet)
	})

	t.Run("invalid CIDR", func(t *testing.T) {
		_, err := ParseCIDR("invalid")
		assert.Error(t, err)
	})
}

func TestTrustedSubnetInterceptor_NilSubnet(t *testing.T) {
	interceptor := TrustedSubnetInterceptor(nil, zap.NewNop())

	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return "ok", nil
	}

	resp, err := interceptor(context.Background(), nil, &grpc.UnaryServerInfo{FullMethod: "/test"}, handler)
	assert.NoError(t, err)
	assert.Equal(t, "ok", resp)
}

func TestTrustedSubnetInterceptor_MissingMetadata(t *testing.T) {
	_, subnet, _ := net.ParseCIDR("192.168.1.0/24")
	interceptor := TrustedSubnetInterceptor(subnet, zap.NewNop())

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
	_, subnet, _ := net.ParseCIDR("192.168.1.0/24")
	interceptor := TrustedSubnetInterceptor(subnet, zap.NewNop())

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
	_, subnet, _ := net.ParseCIDR("192.168.1.0/24")
	interceptor := TrustedSubnetInterceptor(subnet, zap.NewNop())

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

func TestTrustedSubnetInterceptor_InvalidClientIP(t *testing.T) {
	_, subnet, _ := net.ParseCIDR("192.168.1.0/24")
	interceptor := TrustedSubnetInterceptor(subnet, zap.NewNop())

	md := metadata.Pairs("x-real-ip", "not-an-ip")
	ctx := metadata.NewIncomingContext(context.Background(), md)

	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return "ok", nil
	}

	resp, err := interceptor(ctx, nil, &grpc.UnaryServerInfo{FullMethod: "/test"}, handler)
	assert.Nil(t, resp)

	st, _ := status.FromError(err)
	assert.Equal(t, codes.PermissionDenied, st.Code())
}

func TestTrustedSubnetInterceptor_MissingXRealIP(t *testing.T) {
	_, subnet, _ := net.ParseCIDR("192.168.1.0/24")
	interceptor := TrustedSubnetInterceptor(subnet, zap.NewNop())

	md := metadata.Pairs("other-header", "value")
	ctx := metadata.NewIncomingContext(context.Background(), md)

	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return "ok", nil
	}

	resp, err := interceptor(ctx, nil, &grpc.UnaryServerInfo{FullMethod: "/test"}, handler)
	assert.Nil(t, resp)

	st, _ := status.FromError(err)
	assert.Equal(t, codes.PermissionDenied, st.Code())
}
