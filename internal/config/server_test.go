package config

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewServerConfig_Defaults(t *testing.T) {
	cfg, err := NewServerConfig([]string{})

	require.NoError(t, err)
	assert.Equal(t, "localhost:8080", cfg.HTTPAddress)
	assert.Equal(t, 20, cfg.StoreInterval)
	assert.Equal(t, 5*time.Second, cfg.RequestTimeout)
	assert.Equal(t, 10*time.Second, cfg.ShutdownTimeout)
}

func TestNewServerConfig_StoreIntervalNegative(t *testing.T) {
	_, err := NewServerConfig([]string{"-i", "-1"})

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "store interval must be non-negative")
}

func TestNewServerConfig_GRPCAddress(t *testing.T) {
	cfg, err := NewServerConfig([]string{"-grpc-addr", ":9090"})

	require.NoError(t, err)
	assert.Equal(t, ":9090", cfg.GRPCAddress)
}

func TestNewServerConfig_TrustedSubnet(t *testing.T) {
	cfg, err := NewServerConfig([]string{"-t", "192.168.1.0/24"})

	require.NoError(t, err)
	assert.Equal(t, "192.168.1.0/24", cfg.TrustedSubnet)
}
