// internal/audit/publisher.go
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

// Event представляет событие аудита.
type Event struct {
	Timestamp int64    `json:"ts"`
	Metrics   []string `json:"metrics"`
	IPAddress string   `json:"ip_address"`
}

// Publisher принимает события аудита и отправляет их всем подписчикам.
type Publisher struct {
	events   chan Event
	handlers []func(Event)
	logger   *zap.Logger
	wg       sync.WaitGroup
}

// NewPublisher создаёт Publisher на основе конфигурации.
// Возвращает nil, если аудит не настроен.
func NewPublisher(auditFile, auditURL string, logger *zap.Logger) (*Publisher, error) {
	var handlers []func(Event)

	if auditFile != "" {
		h, err := newFileHandler(auditFile)
		if err != nil {
			return nil, err
		}
		handlers = append(handlers, h)
		logger.Info("Audit publisher: file writer enabled", zap.String("path", auditFile))
	}

	if auditURL != "" {
		handlers = append(handlers, newHTTPHandler(auditURL))
		logger.Info("Audit publisher: HTTP sender enabled", zap.String("url", auditURL))
	}

	if len(handlers) == 0 {
		return nil, nil
	}

	return &Publisher{
		events:   make(chan Event, 64),
		handlers: handlers,
		logger:   logger,
	}, nil
}

func (p *Publisher) Run(ctx context.Context) error {
	p.logger.Info("Audit publisher started")

	for {
		select {
		case e := <-p.events:
			for _, h := range p.handlers {
				p.wg.Add(1)
				go func(h func(Event)) {
					defer p.wg.Done()
					h(e)
				}(h)
			}
		case <-ctx.Done():
			p.logger.Info("Audit publisher shutting down")

			for {
				select {
				case e := <-p.events:
					for _, h := range p.handlers {
						p.wg.Add(1)
						go func(h func(Event)) {
							defer p.wg.Done()
							h(e)
						}(h)
					}
				default:
					p.wg.Wait()
					p.logger.Info("Audit publisher stopped gracefully")
					return nil
				}
			}
		}
	}
}

// Notify отправляет событие аудита.
func (p *Publisher) Notify(event Event) {
	p.events <- event
}

func newFileHandler(path string) (func(Event), error) {
	f, err := os.OpenFile(path, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return nil, fmt.Errorf("audit file: %w", err)
	}
	return func(e Event) {
		json.NewEncoder(f).Encode(e)
	}, nil
}

func newHTTPHandler(url string) func(Event) {
	client := &http.Client{Timeout: 5 * time.Second}
	return func(e Event) {
		var buf bytes.Buffer
		json.NewEncoder(&buf).Encode(e)
		req, _ := http.NewRequest(http.MethodPost, url, &buf)
		req.Header.Set("Content-Type", "application/json")
		resp, err := client.Do(req)
		if err == nil {
			resp.Body.Close()
		}
	}
}
