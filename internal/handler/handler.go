package handler

import (
	"html/template"
	"net/http"

	"metrics-collector/internal/service"
	"metrics-collector/internal/templates"

	"github.com/go-chi/chi/middleware"
	"github.com/go-chi/chi/v5"
	"go.uber.org/zap"
)

type Handler struct {
	service                *service.MetricsService
	logger                 *zap.SugaredLogger
	allMetricsHTMLTemplate *template.Template
}

func NewHandler(service *service.MetricsService, logger *zap.SugaredLogger) (*Handler, error) {
	tmpl, err := template.ParseFS(templates.FS, "metrics.html")
	if err != nil {
		return nil, err
	}

	return &Handler{
		service:                service,
		logger:                 logger,
		allMetricsHTMLTemplate: tmpl,
	}, nil
}

func (h *Handler) NewMetricsRouter() chi.Router {
	r := chi.NewRouter()
	r.Use(middleware.StripSlashes)
	r.Use(func(next http.Handler) http.Handler {
		return WithLogging(next, h.logger)
	})

	r.Get("/", h.RootHandle)
	r.Get("/value/{type}/{name}", h.ValueHandle)
	r.Post("/update/{type}/{name}/{value}", h.UpdateHandle)
	r.Post("/value", h.ValueHandleV2)
	r.Post("/update", h.UpdateHandleV2)

	return r
}
