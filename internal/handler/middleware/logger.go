package middleware

import (
	"net/http"
	"time"

	"github.com/google/uuid"
	"github.com/rs/zerolog/log"
)

type Logger struct {
	ipService IPService
}

func NewLogger(ipService IPService) *Logger {
	return &Logger{ipService: ipService}
}

func (l *Logger) Handle(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		requestID := uuid.New().String()
		w.Header().Set("X-Request-ID", requestID)
		rw := &responseWriter{ResponseWriter: w}
		logger := log.With().
			Str("ip", l.ipService.GetIP(r.Context(), r)).
			Str("path", r.URL.Path).
			Str("method", r.Method).
			Str("query", r.URL.RawQuery).
			Str("request_id", requestID).
			Str("user_agent", r.UserAgent()).
			Int("request_size", int(r.ContentLength)).
			Logger()
		logger.Info().Msg("request started")
		next.ServeHTTP(rw, r)
		logger.Info().
			Int("response_size", rw.size).
			Int("status_code", rw.statusCode).
			Dur("duration", time.Since(start)).
			Msg("request finished")
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
