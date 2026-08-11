//go:build integration
// +build integration

package agent

import (
	"context"
	"testing"
	"time"

	"metrics-collector/internal/config"

	"go.uber.org/zap"
)

func TestAgent_Integration(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test")
	}

	logger := zap.NewNop()
	cfg := &config.AgentConfig{
		PollInterval:   1,
		ReportInterval: 1,
		HTTPAddress:    "localhost:0",
		RateLimit:      1,
	}

	agent := NewAgent(cfg, logger)

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	errCh := make(chan error, 1)
	go func() {
		errCh <- agent.Run(ctx)
	}()

	// Проверяем что агент запустился и работает
	time.Sleep(2 * time.Second)
	cancel()

	if err := <-errCh; err != nil {
		t.Errorf("Agent stopped with error: %v", err)
	}

	// Проверяем что метрики собирались
	if len(agent.store.GetMetricsSnapshot()) == 0 {
		t.Error("No metrics collected during test")
	}
}
