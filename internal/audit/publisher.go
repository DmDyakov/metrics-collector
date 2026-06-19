// internal/audit/publisher.go
package audit

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
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
	done     chan struct{}
	stopped  chan struct{}
	handlers []func(Event)
	logger   *zap.Logger
}

// NewPublisher создаёт Publisher на основе конфигурации.
// Возвращает nil, если аудит не настроен.
func NewPublisher(auditFile, auditURL string, logger *zap.Logger) (*Publisher, error) {
	var handlers []func(Event)

	if auditFile != "" {
		f, err := os.OpenFile(auditFile, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
		if err != nil {
			return nil, fmt.Errorf("audit file: %w", err)
		}
		handlers = append(handlers, func(e Event) {
			json.NewEncoder(f).Encode(e)
		})
	}

	if auditURL != "" {
		client := &http.Client{Timeout: 5 * time.Second}
		handlers = append(handlers, func(e Event) {
			var buf bytes.Buffer
			json.NewEncoder(&buf).Encode(e)
			req, _ := http.NewRequest(http.MethodPost, auditURL, &buf)
			req.Header.Set("Content-Type", "application/json")
			resp, err := client.Do(req)
			if err == nil {
				resp.Body.Close()
			}
		})
	}

	if len(handlers) == 0 {
		return nil, nil
	}

	pub := &Publisher{
		events:   make(chan Event, 64),
		done:     make(chan struct{}),
		stopped:  make(chan struct{}),
		handlers: handlers,
		logger:   logger,
	}

	if auditFile != "" {
		logger.Info("Audit publisher: file writer enabled", zap.String("path", auditFile))
	}
	if auditURL != "" {
		logger.Info("Audit publisher: HTTP sender enabled", zap.String("url", auditURL))
	}
	go pub.run()

	return pub, nil
}

func (p *Publisher) run() {
	defer close(p.stopped)

	for {
		select {
		case e := <-p.events:
			for _, h := range p.handlers {
				go h(e)
			}
		case <-p.done:
			for {
				select {
				case e := <-p.events:
					for _, h := range p.handlers {
						go h(e)
					}
				default:
					return
				}
			}
		}
	}
}

// Notify отправляет событие аудита.
func (p *Publisher) Notify(event Event) {
	p.events <- event
}

// Shutdown завершает работу Publisher с таймаутом.
func (p *Publisher) Shutdown(ctx context.Context) error {
	close(p.done)

	select {
	case <-p.stopped:
		return nil
	case <-ctx.Done():
		p.logger.Warn("audit shutdown timed out")
		return ctx.Err()
	}
}
