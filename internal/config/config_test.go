package config

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewAgentConfig(t *testing.T) {
	t.Run("valid args", func(t *testing.T) {
		cfg, err := NewAgentConfig([]string{
			"-a", "localhost:9090",
			"-p", "5",
			"-r", "15",
			"-l", "3",
		})
		require.NoError(t, err)
		assert.Equal(t, "localhost:9090", cfg.ServerBaseURL)
		assert.Equal(t, 5, cfg.PollInterval)
		assert.Equal(t, 15, cfg.ReportInterval)
		assert.Equal(t, 3, cfg.RateLimit)
	})

	t.Run("empty server url", func(t *testing.T) {
		_, err := NewAgentConfig([]string{"-a", ""})
		assert.Error(t, err)
	})
}

func TestNewServerConfig(t *testing.T) {
	t.Run("valid args", func(t *testing.T) {
		cfg, err := NewServerConfig([]string{
			"-a", "localhost:8080",
			"-i", "30",
		})
		require.NoError(t, err)
		assert.Equal(t, "localhost:8080", cfg.ServerBaseURL)
		assert.Equal(t, 30, cfg.StoreInterval)
	})

	t.Run("negative store interval", func(t *testing.T) {
		_, err := NewServerConfig([]string{"-i", "-1"})
		assert.Error(t, err)
	})
}
