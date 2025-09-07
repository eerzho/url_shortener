package middleware

import (
	"log/slog"
	"net/http"
	"time"

	"github.com/google/uuid"
)

type Logger struct {
	logger *slog.Logger
}

func NewLogger(
	logger *slog.Logger,
) *Logger {
	return &Logger{
		logger: logger,
	}
}

func (l *Logger) Handle(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()

		requestID := r.Header.Get("X-Request-ID")
		if requestID == "" {
			requestID = uuid.NewString()
		}
		w.Header().Set("X-Request-ID", requestID)

		logger := l.logger.With(
			slog.String("path", r.URL.Path),
			slog.String("method", r.Method),
			slog.String("query", r.URL.RawQuery),
			slog.String("request_id", requestID),
			slog.String("user_agent", r.UserAgent()),
			slog.Int("request_size", int(r.ContentLength)),
		)

		logger.InfoContext(r.Context(), "request started")

		rw := &responseWriter{ResponseWriter: w}
		next.ServeHTTP(rw, r)

		logger.InfoContext(r.Context(), "request finished",
			slog.Int("response_size", rw.size),
			slog.Int("status_code", rw.statusCode),
			slog.Duration("duration", time.Since(start)),
		)
	})
}

type responseWriter struct {
	http.ResponseWriter

	statusCode int
	size       int
}

func (rw *responseWriter) WriteHeader(statusCode int) {
	rw.statusCode = statusCode
	rw.ResponseWriter.WriteHeader(statusCode)
}

func (rw *responseWriter) Write(b []byte) (int, error) {
	size, err := rw.ResponseWriter.Write(b)
	rw.size += size
	return size, err
}
