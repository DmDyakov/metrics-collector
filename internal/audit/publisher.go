// Package audit реализует аудит
package audit

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"sync"
	"time"

	"go.uber.org/zap"
)

type Event struct {
	Timestamp int64    `json:"ts"`
	Metrics   []string `json:"metrics"`
	IPAddress string   `json:"ip_address"`
}

type Publisher struct {
	events   chan Event
	handlers []func(Event)
	logger   *zap.Logger
	wg       sync.WaitGroup
	ctx      context.Context
	cancel   context.CancelFunc
	file     *os.File
}

// NewPublisher creates a Publisher for writing to file and/or sending via HTTP.
// Returns nil if audit is not configured.
func NewPublisher(auditFile, auditURL string, logger *zap.Logger) (*Publisher, error) {
	var handlers []func(Event)
	var file *os.File

	if auditFile != "" {
		var err error
		file, err = os.OpenFile(auditFile, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
		if err != nil {
			return nil, fmt.Errorf("audit file: %w", err)
		}
		handlers = append(handlers, func(e Event) {
			if err := json.NewEncoder(file).Encode(e); err != nil {
				logger.Error("failed to encode audit event to file", zap.Error(err))
			}
		})
		logger.Info("Audit publisher: file writer enabled", zap.String("path", auditFile))
	}

	if auditURL != "" {
		handlers = append(handlers, newHTTPHandler(auditURL, logger))
		logger.Info("Audit publisher: HTTP sender enabled", zap.String("url", auditURL))
	}

	if len(handlers) == 0 {
		return nil, nil
	}

	return &Publisher{
		events:   make(chan Event, 64),
		handlers: handlers,
		logger:   logger,
		file:     file,
	}, nil
}

// Run starts processing audit events until the context is done.
func (p *Publisher) Run(ctx context.Context) error {
	p.ctx, p.cancel = context.WithCancel(ctx)
	defer p.cancel()

	p.logger.Info("Audit publisher started")

	go func() {
		for e := range p.events {
			for _, h := range p.handlers {
				p.wg.Add(1)
				go func(h func(Event)) {
					defer p.wg.Done()
					h(e)
				}(h)
			}
		}
	}()

	<-ctx.Done()
	p.logger.Info("Audit publisher shutting down")

	p.cancel()
	p.wg.Wait()

	if p.file != nil {
		p.file.Close()
	}

	p.logger.Info("Audit publisher stopped")
	return nil
}

// Notify sends an audit event to the channel.
func (p *Publisher) Notify(event Event) {
	select {
	case p.events <- event:
	case <-p.ctx.Done():
	}
}

func newHTTPHandler(url string, logger *zap.Logger) func(Event) {
	client := &http.Client{Timeout: 5 * time.Second}
	return func(e Event) {
		var buf bytes.Buffer
		if err := json.NewEncoder(&buf).Encode(e); err != nil {
			logger.Error("failed to encode audit event for HTTP", zap.Error(err))
			return
		}

		req, err := http.NewRequest(http.MethodPost, url, &buf)
		if err != nil {
			logger.Error("failed to create HTTP request for audit", zap.Error(err))
			return
		}

		req.Header.Set("Content-Type", "application/json")
		resp, err := client.Do(req)
		if err != nil {
			logger.Error("failed to send audit event via HTTP", zap.Error(err))
			return
		}
		defer resp.Body.Close()
	}
}
