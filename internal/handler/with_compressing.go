package handler

import (
	"compress/gzip"
	"io"
	"net/http"
	"strings"
)

type gzipResponseWriter struct {
	http.ResponseWriter
	writer      *gzip.Writer
	wroteHeader bool
}

func (gzw *gzipResponseWriter) Write(b []byte) (int, error) {
	if !gzw.wroteHeader {
		gzw.WriteHeader(http.StatusOK)
	}
	return gzw.writer.Write(b)
}

func (gzw *gzipResponseWriter) WriteHeader(statusCode int) {
	if gzw.wroteHeader {
		return
	}
	gzw.wroteHeader = true

	gzw.Header().Set("Content-Encoding", "gzip")
	gzw.Header().Del("Content-Length")

	gzw.ResponseWriter.WriteHeader(statusCode)
}

func (gzw *gzipResponseWriter) Close() error {
	return gzw.writer.Close()
}

type gzipRequestReader struct {
	io.ReadCloser
	originalBody io.ReadCloser
}

func (gzr *gzipRequestReader) Close() error {
	if err := gzr.ReadCloser.Close(); err != nil {
		return err
	}
	return gzr.originalBody.Close()
}

func (h *Handler) WithCompressing(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// ========== Распаковка запроса ==========
		if strings.Contains(r.Header.Get("Content-Encoding"), "gzip") {
			gzReader, err := gzip.NewReader(r.Body)
			if err != nil {
				http.Error(w, "Bad request: invalid gzip", http.StatusBadRequest)
				return
			}
			r.Body = &gzipRequestReader{
				ReadCloser:   gzReader,
				originalBody: r.Body,
			}
			r.Header.Del("Content-Encoding")
			r.Header.Del("Content-Length")
		}

		// ========== Сжатие ответа ==========
		if strings.Contains(r.Header.Get("Accept-Encoding"), "gzip") {
			gzWriter := gzip.NewWriter(w)
			gzw := &gzipResponseWriter{
				ResponseWriter: w,
				writer:         gzWriter,
			}
			next.ServeHTTP(gzw, r)
			gzWriter.Close()
			return
		}
		// ========== Если нет нужных заголовков ==========
		next.ServeHTTP(w, r)
	})
}
