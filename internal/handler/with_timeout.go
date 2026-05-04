package handler

import (
	"context"
	"net/http"
	"time"
)

func (h *Handler) WithTimeout(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctx, cancel := context.WithTimeout(r.Context(), h.requestTimeout)
		defer cancel()

		ctx = context.WithValue(ctx, startTimeKey, time.Now())

		next.ServeHTTP(w, r.WithContext(ctx))
	})
}
