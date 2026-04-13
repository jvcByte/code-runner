package middleware

import (
	"log/slog"
	"net/http"
	"strings"
	"time"
)

// responseWriter wraps http.ResponseWriter to capture the status code.
type responseWriter struct {
	http.ResponseWriter
	status int
}

func (rw *responseWriter) WriteHeader(code int) {
	rw.status = code
	rw.ResponseWriter.WriteHeader(code)
}

// Logger logs each request with IP, user agent, method, path, status, and duration.
func Logger(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()

		rw := &responseWriter{ResponseWriter: w, status: http.StatusOK}
		next.ServeHTTP(rw, r)

		ip := realIP(r)
		duration := time.Since(start)

		attrs := []slog.Attr{
			slog.String("method", r.Method),
			slog.String("path", r.URL.Path),
			slog.Int("status", rw.status),
			slog.String("ip", ip),
			slog.String("agent", r.UserAgent()),
			slog.Duration("duration", duration),
		}

		if rw.status >= 500 {
			slog.LogAttrs(r.Context(), slog.LevelError, "request", attrs...)
		} else if rw.status >= 400 {
			slog.LogAttrs(r.Context(), slog.LevelWarn, "request", attrs...)
		} else {
			slog.LogAttrs(r.Context(), slog.LevelInfo, "request", attrs...)
		}
	})
}

// realIP extracts the client IP, respecting X-Forwarded-For from proxies.
func realIP(r *http.Request) string {
	if xff := r.Header.Get("X-Forwarded-For"); xff != "" {
		return strings.Split(xff, ",")[0]
	}
	if xri := r.Header.Get("X-Real-IP"); xri != "" {
		return xri
	}
	// Strip port from RemoteAddr
	addr := r.RemoteAddr
	if i := strings.LastIndex(addr, ":"); i != -1 {
		return addr[:i]
	}
	return addr
}
