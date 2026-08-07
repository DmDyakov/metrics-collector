// Package http реализует HTTP-сервер.
package http

import (
	"context"
	"errors"
	"net/http"
	"time"

	"metrics-collector/internal/config"
	"metrics-collector/internal/server/audit"
	"metrics-collector/internal/server/transport/http/handler"
	"metrics-collector/internal/server/transport/http/middleware"
	"metrics-collector/pkg/encryptor"
	"metrics-collector/pkg/signer"

	"github.com/go-chi/chi/v5"
	chimw "github.com/go-chi/chi/v5/middleware"
	"go.uber.org/zap"
)

// New создаёт HTTP-сервер с настроенными маршрутами и middleware.
func New(
	cfg *config.ServerConfig,
	healthHandler *handler.HealthHandler,
	metricsHandler *handler.MetricsHandler,
	auditPublisher *audit.Publisher,
	signer *signer.Signer,
	enc *encryptor.Encryptor,
	logger *zap.Logger,
) *http.Server {
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

		if auditPublisher != nil {
			r.Use(middleware.WithAudit(logger, auditPublisher))
		}

		r.Post("/update/{type}/{name}/{value}", metricsHandler.UpdateMetricByURL)
		r.Post("/update", metricsHandler.UpdateMetricByJSON)
		r.Post("/updates", metricsHandler.UpdateMetricsBatch)
	})

	return &http.Server{
		Addr:         cfg.ServerBaseURL,
		Handler:      r,
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 10 * time.Second,
		IdleTimeout:  30 * time.Second,
	}
}

// Run запускает HTTP-сервер и graceful shutdown.
func Run(ctx context.Context, srv *http.Server, shutdownTimeout time.Duration, logger *zap.Logger) error {
	logger.Info("HTTP server started", zap.String("addr", srv.Addr))

	go func() {
		if err := srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			logger.Error("HTTP server failed", zap.Error(err))
		}
	}()

	<-ctx.Done()
	logger.Info("Shutting down HTTP server...")
	shutdownCtx, cancel := context.WithTimeout(context.Background(), shutdownTimeout)
	defer cancel()
	return srv.Shutdown(shutdownCtx)
}
