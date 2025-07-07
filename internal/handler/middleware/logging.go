package middleware

import (
	"context"
	"net/http"
	"time"
	"url_shortener/internal/constant"

	"github.com/google/uuid"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

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

func Logging(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		requestId := generateId()
		rw := &responseWriter{ResponseWriter: w}
		logger := log.With().
			Str("request_id", requestId).
			Str("method", r.Method).
			Str("path", r.URL.Path).
			Str("query", r.URL.RawQuery).
			Str("user_agent", r.UserAgent()).
			Str("remote_addr", r.RemoteAddr).
			Str("referer", r.Referer()).
			Logger()

		ctx := context.WithValue(r.Context(), constant.RequestIdKey, requestId)
		r = r.WithContext(ctx)

		w.Header().Set("X-Request-ID", requestId)

		logger.Info().Msg("request started")

		next.ServeHTTP(rw, r)

		var level zerolog.Level
		if rw.statusCode >= 500 {
			level = zerolog.ErrorLevel
		} else if rw.statusCode >= 400 {
			level = zerolog.WarnLevel
		} else {
			level = zerolog.InfoLevel
		}

		logger.WithLevel(level).
			Int("status_code", rw.statusCode).
			Int("size", rw.size).
			Dur("duration", time.Since(start)).
			Msg("request finished")
	})
}

func generateId() string {
	return uuid.New().String()
}
