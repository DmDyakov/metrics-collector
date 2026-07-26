package config

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewAgentConfig(t *testing.T) {
	os.Unsetenv("ADDRESS")
	os.Unsetenv("POLL_INTERVAL")
	os.Unsetenv("REPORT_INTERVAL")
	os.Unsetenv("KEY")
	os.Unsetenv("CRYPTO_KEY")
	os.Unsetenv("RATE_LIMIT")
	os.Unsetenv("CONFIG")

	t.Run("default values", func(t *testing.T) {
		cfg, err := NewAgentConfig([]string{})
		require.NoError(t, err)
		assert.Equal(t, defaultServerBaseURL, cfg.ServerBaseURL)
		assert.Equal(t, defaultPollInterval, cfg.PollInterval)
		assert.Equal(t, defaultReportInterval, cfg.ReportInterval)
		assert.Equal(t, defaultRateLimit, cfg.RateLimit)
		assert.Empty(t, cfg.SecretKey)
	})

	t.Run("flags override defaults", func(t *testing.T) {
		cfg, err := NewAgentConfig([]string{
			"-a", "localhost:9090",
			"-p", "5",
			"-r", "15",
			"-k", "mykey",
			"-l", "3",
		})
		require.NoError(t, err)
		assert.Equal(t, "localhost:9090", cfg.ServerBaseURL)
		assert.Equal(t, 5, cfg.PollInterval)
		assert.Equal(t, 15, cfg.ReportInterval)
		assert.Equal(t, "mykey", cfg.SecretKey)
		assert.Equal(t, 3, cfg.RateLimit)
	})

	t.Run("env overrides defaults", func(t *testing.T) {
		t.Setenv("ADDRESS", "localhost:7070")
		t.Setenv("POLL_INTERVAL", "7")
		t.Setenv("RATE_LIMIT", "8")

		cfg, err := NewAgentConfig([]string{})
		require.NoError(t, err)
		assert.Equal(t, "localhost:7070", cfg.ServerBaseURL)
		assert.Equal(t, 7, cfg.PollInterval)
		assert.Equal(t, 10, cfg.ReportInterval)
		assert.Equal(t, 8, cfg.RateLimit)
	})

	t.Run("flags override env", func(t *testing.T) {
		t.Setenv("ADDRESS", "localhost:7070")
		t.Setenv("POLL_INTERVAL", "7")

		cfg, err := NewAgentConfig([]string{
			"-a", "localhost:9090",
			"-p", "5",
		})
		require.NoError(t, err)
		assert.Equal(t, "localhost:9090", cfg.ServerBaseURL)
		assert.Equal(t, 5, cfg.PollInterval)
	})

	t.Run("json overrides defaults", func(t *testing.T) {
		tmpFile, err := os.CreateTemp("", "config_*.json")
		require.NoError(t, err)
		defer os.Remove(tmpFile.Name())

		_, err = tmpFile.WriteString(`{"address": "localhost:6060", "report_interval": 25}`)
		require.NoError(t, err)
		tmpFile.Close()

		cfg, err := NewAgentConfig([]string{"-c", tmpFile.Name()})
		require.NoError(t, err)
		assert.Equal(t, "localhost:6060", cfg.ServerBaseURL)
		assert.Equal(t, 2, cfg.PollInterval)
		assert.Equal(t, 25, cfg.ReportInterval)
	})

	t.Run("env overrides json", func(t *testing.T) {
		tmpFile, err := os.CreateTemp("", "config_*.json")
		require.NoError(t, err)
		defer os.Remove(tmpFile.Name())

		_, err = tmpFile.WriteString(`{"address": "localhost:6060", "report_interval": 25}`)
		require.NoError(t, err)
		tmpFile.Close()

		t.Setenv("ADDRESS", "localhost:7070")

		cfg, err := NewAgentConfig([]string{"-c", tmpFile.Name()})
		require.NoError(t, err)
		assert.Equal(t, "localhost:7070", cfg.ServerBaseURL)
		assert.Equal(t, 25, cfg.ReportInterval)
	})

	t.Run("flags override json", func(t *testing.T) {
		tmpFile, err := os.CreateTemp("", "config_*.json")
		require.NoError(t, err)
		defer os.Remove(tmpFile.Name())

		_, err = tmpFile.WriteString(`{"address": "localhost:6060", "report_interval": 25}`)
		require.NoError(t, err)
		tmpFile.Close()

		cfg, err := NewAgentConfig([]string{"-c", tmpFile.Name(), "-a", "localhost:9090"})
		require.NoError(t, err)
		assert.Equal(t, "localhost:9090", cfg.ServerBaseURL)
		assert.Equal(t, 25, cfg.ReportInterval)
	})

	t.Run("full chain: json -> env -> flag", func(t *testing.T) {
		tmpFile, err := os.CreateTemp("", "config_*.json")
		require.NoError(t, err)
		defer os.Remove(tmpFile.Name())

		_, err = tmpFile.WriteString(`{"address": "localhost:6060", "poll_interval": 3}`)
		require.NoError(t, err)
		tmpFile.Close()

		t.Setenv("POLL_INTERVAL", "7")

		cfg, err := NewAgentConfig([]string{"-c", tmpFile.Name(), "-a", "localhost:9090"})
		require.NoError(t, err)
		assert.Equal(t, "localhost:9090", cfg.ServerBaseURL)
		assert.Equal(t, 7, cfg.PollInterval)
		assert.Equal(t, 10, cfg.ReportInterval)
	})

	t.Run("empty server url", func(t *testing.T) {
		_, err := NewAgentConfig([]string{"-a", ""})
		assert.Error(t, err)
	})
}

func TestNewServerConfig(t *testing.T) {
	os.Unsetenv("ADDRESS")
	os.Unsetenv("STORE_INTERVAL")
	os.Unsetenv("FILE_STORAGE_PATH")
	os.Unsetenv("RESTORE")
	os.Unsetenv("DATABASE_DSN")
	os.Unsetenv("KEY")
	os.Unsetenv("CRYPTO_KEY")
	os.Unsetenv("CONFIG")

	t.Run("default values", func(t *testing.T) {
		cfg, err := NewServerConfig([]string{})
		require.NoError(t, err)
		assert.Equal(t, "localhost:8080", cfg.ServerBaseURL)
		assert.Equal(t, 20, cfg.StoreInterval)
		assert.False(t, cfg.Restore)
	})

	t.Run("flags override defaults", func(t *testing.T) {
		cfg, err := NewServerConfig([]string{
			"-a", "localhost:9090",
			"-i", "30",
			"-r",
		})
		require.NoError(t, err)
		assert.Equal(t, "localhost:9090", cfg.ServerBaseURL)
		assert.Equal(t, 30, cfg.StoreInterval)
		assert.True(t, cfg.Restore)
	})

	t.Run("negative store interval", func(t *testing.T) {
		_, err := NewServerConfig([]string{"-i", "-1"})
		assert.Error(t, err)
	})
}
