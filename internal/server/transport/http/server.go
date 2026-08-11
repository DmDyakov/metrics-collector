// Package http реализует HTTP-сервер.
package http

import (
	"context"
	"errors"
	"net/http"
	"time"

	"metrics-collector/internal/config"
	"metrics-collector/internal/domain/audit"

	"metrics-collector/internal/server/transport/http/handler"
	"metrics-collector/internal/server/transport/http/middleware"
	"metrics-collector/pkg/encryptor"
	"metrics-collector/pkg/publisher"
	"metrics-collector/pkg/signer"

	"github.com/go-chi/chi/v5"
	chimw "github.com/go-chi/chi/v5/middleware"
	"go.uber.org/zap"
)

type Server struct {
	*http.Server
	logger          *zap.Logger
	shutdownTimeout time.Duration
}

// NewServer создаёт HTTP-сервер с настроенными маршрутами и middleware.
func NewServer(
	cfg *config.ServerConfig,
	healthHandler *handler.HealthHandler,
	metricsHandler *handler.MetricsHandler,
	auditPub *publisher.Publisher[audit.Event],
	signer *signer.Signer,
	enc *encryptor.Encryptor,
	logger *zap.Logger,
) *Server {
	r := chi.NewRouter()
	r.Use(chimw.StripSlashes)
	r.Use(middleware.WithTimeout(cfg.RequestTimeout))
	r.Use(middleware.WithLogging(logger))
	r.Use(middleware.WithSignature(logger, signer))
	r.Use(middleware.WithDecryption(logger, enc))
	r.Use(middleware.WithCompressing)

	r.Get("/ping", healthHandler.HealthDB)

	r.Group(func(r chi.Router) {
		r.Get("/", metricsHandler.ListMetrics)
		r.Get("/value/{type}/{name}", metricsHandler.GetMetricValue)
		r.Post("/value", metricsHandler.GetMetric)
	})

	r.Group(func(r chi.Router) {
		r.Use(middleware.WithTrustedSubnet(cfg.TrustedSubnet))

		if auditPub != nil {
			r.Use(middleware.WithAudit(logger, auditPub))
		}

		r.Post("/update/{type}/{name}/{value}", metricsHandler.UpdateMetricByURL)
		r.Post("/update", metricsHandler.UpdateMetricByJSON)
		r.Post("/updates", metricsHandler.UpdateMetricsBatch)
	})

	srv := &http.Server{
		Addr:         cfg.HTTPAddress,
		Handler:      r,
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 10 * time.Second,
		IdleTimeout:  30 * time.Second,
	}

	return &Server{
		Server:          srv,
		logger:          logger,
		shutdownTimeout: 5 * time.Second,
	}
}

// Run запускает HTTP-сервер.
func (s *Server) Run(ctx context.Context) error {
	s.logger.Info("HTTP server started", zap.String("addr", s.Addr))

	go func() {
		if err := s.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			s.logger.Error("HTTP server failed", zap.Error(err))
		}
	}()

	<-ctx.Done()
	s.logger.Info("Shutting down HTTP server...")
	shCtx, cancel := context.WithTimeout(context.Background(), s.shutdownTimeout)
	defer cancel()

	if err := s.Shutdown(shCtx); err != nil {
		s.logger.Error("HTTP server failed to shutdown", zap.Error(err))
	}
	s.logger.Info("HTTP server stopped gracefully")
	return nil
}
