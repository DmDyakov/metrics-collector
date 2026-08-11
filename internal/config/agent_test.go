package config

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewAgentConfig_Defaults(t *testing.T) {
	t.Setenv("AGENT_IP", "10.0.0.1")

	cfg, err := NewAgentConfig([]string{})
	require.NoError(t, err)
	assert.Equal(t, "localhost:8080", cfg.HTTPAddress)
	assert.Equal(t, 2, cfg.PollInterval)
	assert.Equal(t, 10, cfg.ReportInterval)
	assert.Equal(t, 2, cfg.RateLimit)
}
func TestNewAgentConfig_AddressRequired(t *testing.T) {
	os.Unsetenv("ADDRESS")
	os.Unsetenv("AGENT_IP")

	_, err := NewAgentConfig([]string{"-a", ""})

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "server URL can not be empty")
}

func TestNewAgentConfig_AgentIPRequired(t *testing.T) {
	os.Unsetenv("AGENT_IP")

	_, err := NewAgentConfig([]string{})

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "agent ip can not be empty")
}

func TestNewAgentConfig_GRPCAddress(t *testing.T) {
	t.Setenv("AGENT_IP", "10.0.0.1")
	defer os.Unsetenv("AGENT_IP")

	cfg, err := NewAgentConfig([]string{"-grpc-addr", ":9090"})

	require.NoError(t, err)
	assert.Equal(t, ":9090", cfg.GRPCAddress)
}

func TestNewAgentConfig_PollIntervalPositive(t *testing.T) {
	t.Setenv("AGENT_IP", "10.0.0.1")
	defer os.Unsetenv("AGENT_IP")

	_, err := NewAgentConfig([]string{"-p", "0"})

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "poll interval must be positive")
}

func TestNewAgentConfig_ReportLessThanPoll(t *testing.T) {
	t.Setenv("AGENT_IP", "10.0.0.1")
	defer os.Unsetenv("AGENT_IP")

	_, err := NewAgentConfig([]string{"-p", "10", "-r", "5"})

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "must be >= poll interval")
}
