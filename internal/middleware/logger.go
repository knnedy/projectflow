package middleware

import (
	"log/slog"
	"net/http"
	"time"
)

// responseWriter wraps http.ResponseWriter to capture the status code
type responseWriter struct {
	http.ResponseWriter
	status int
}

func newResponseWriter(w http.ResponseWriter) *responseWriter {
	return &responseWriter{ResponseWriter: w, status: http.StatusOK}
}

func (rw *responseWriter) WriteHeader(status int) {
	rw.status = status
	rw.ResponseWriter.WriteHeader(status)
}

// Logger logs each request with method, path, status code and duration
func Logger(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		wrappedWriter := newResponseWriter(w)

		next.ServeHTTP(wrappedWriter, r)

		slog.Info("request",
			"method", r.Method,
			"path", r.URL.Path,
			"status", wrappedWriter.status,
			"duration", time.Since(start).String(),
			"ip", r.RemoteAddr,
		)
	})
}
