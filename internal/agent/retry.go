package agent

import (
	"errors"
	"fmt"
	"metrics-collector/internal/errs"
	"net"
	"net/http"
	"syscall"
	"time"

	"go.uber.org/zap"
)

func (a *Agent) withRetry(doRequest func() (*http.Response, error)) error {
	delays := []time.Duration{1 * time.Second, 3 * time.Second, 5 * time.Second}
	const maxAttempts = 4

	for attempt := 1; attempt <= maxAttempts; attempt++ {
		resp, err := doRequest()

		if err == nil && resp.StatusCode == http.StatusOK {
			return nil
		}

		var signatureError *errs.SignatureError
		if errors.As(err, &signatureError) {
			return err
		}

		if !isRetriable(resp, err) || attempt == maxAttempts {
			if err != nil {
				return err
			}
			if resp != nil {
				return fmt.Errorf("unexpected status: %s", resp.Status)
			}
			return errors.New("empty response")
		}

		delay := delays[attempt-1]
		a.logger.Info("retrying after delay",
			zap.Int("attempt", attempt),
			zap.Duration("delay", delay),
			zap.Error(err),
		)

		time.Sleep(delay)
	}
	return nil
}

func isRetriable(resp *http.Response, err error) bool {
	if err != nil {
		var netErr net.Error
		if errors.As(err, &netErr) && netErr.Timeout() {
			return true
		}

		if errors.Is(err, syscall.ECONNREFUSED) || errors.Is(err, syscall.ECONNRESET) {
			return true
		}
		return false
	}

	switch resp.StatusCode {
	case http.StatusRequestTimeout,
		http.StatusBadGateway,
		http.StatusServiceUnavailable,
		http.StatusGatewayTimeout,
		http.StatusTooManyRequests:
		return true
	default:
		return false
	}
}
