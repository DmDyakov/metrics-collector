// Package publisher implements an asynchronous event publisher.
package publisher

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"sync"
	"time"

	"go.uber.org/zap"
)

// Publisher is an asynchronous event publisher.
type Publisher[T any] struct {
	events   chan T
	handlers []func(T)
	logger   *zap.Logger
	wg       sync.WaitGroup
	stopCh   chan struct{}
	file     *os.File
	client   *http.Client
}

// NewPublisher creates a new Publisher[T].
func NewPublisher[T any](auditFile, auditURL string, timeout time.Duration, logger *zap.Logger) (*Publisher[T], error) {
	var handlers []func(T)
	var file *os.File
	var client *http.Client

	if auditFile != "" {
		var err error
		file, err = os.OpenFile(auditFile, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
		if err != nil {
			return nil, fmt.Errorf("audit file: %w", err)
		}
		handlers = append(handlers, newFileHandler[T](file, logger))
		logger.Info("Audit publisher: file writer enabled", zap.String("path", auditFile))
	}

	if auditURL != "" {
		client := &http.Client{Timeout: timeout}
		handlers = append(handlers, newHTTPHandler[T](auditURL, client, logger))
		logger.Info("Audit publisher: HTTP sender enabled", zap.String("url", auditURL))
	}

	if len(handlers) == 0 {
		return nil, nil
	}

	return &Publisher[T]{
		events:   make(chan T, 64),
		handlers: handlers,
		logger:   logger,
		file:     file,
		client:   client,
		stopCh:   make(chan struct{}),
	}, nil
}

// Run starts the publisher's event processing loop.
func (p *Publisher[T]) Run(ctx context.Context) error {
	if p == nil {
		return nil
	}

	defer func() {
		if p.file != nil {
			p.file.Close()
		}
		if p.client != nil {
			p.client.CloseIdleConnections()
		}
	}()

	p.logger.Info("Publisher started")

	go func() {
		for e := range p.events {
			for _, h := range p.handlers {
				p.wg.Add(1)
				go func(h func(T)) {
					defer p.wg.Done()
					h(e)
				}(h)
			}
		}
	}()

	<-ctx.Done()
	p.logger.Info("Publisher shutting down")
	close(p.stopCh)
	close(p.events)

	p.wg.Wait()

	return nil
}

// Publish publishes an event asynchronously.
func (p *Publisher[T]) Publish(event T) {
	select {
	case <-p.stopCh:
		return
	default:
	}

	select {
	case p.events <- event:
	case <-p.stopCh:
		return
	}
}

// newFileHandler returns a handler function that writes events to file.
func newFileHandler[T any](file *os.File, logger *zap.Logger) func(T) {
	encoder := json.NewEncoder(file)
	return func(e T) {
		if err := encoder.Encode(e); err != nil {
			logger.Error("failed to encode audit event to file", zap.Error(err))
		}
	}
}

// newHTTPHandler returns a handler function that sends events via HTTP POST.
func newHTTPHandler[T any](url string, client *http.Client, logger *zap.Logger) func(T) {
	return func(e T) {
		var buf bytes.Buffer
		if err := json.NewEncoder(&buf).Encode(e); err != nil {
			logger.Error("failed to encode", zap.Error(err))
			return
		}

		req, err := http.NewRequest(http.MethodPost, url, &buf)
		if err != nil {
			logger.Error("failed to create HTTP request", zap.Error(err))
			return
		}
		req.Header.Set("Content-Type", "application/json")

		resp, err := client.Do(req)
		if err != nil {
			logger.Error("failed to send", zap.Error(err))
			return
		}
		defer resp.Body.Close()
		if _, err := io.Copy(io.Discard, resp.Body); err != nil {
			logger.Error("failed to read response body", zap.Error(err))
		}
	}
}
