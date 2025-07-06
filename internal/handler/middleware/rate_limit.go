package middleware

import (
	"net/http"

	"golang.org/x/time/rate"
)

func RateLimit(next http.Handler) http.Handler {
	limiter := rate.NewLimiter(1, 1)
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if !limiter.Allow() {
			http.Error(w, http.StatusText(http.StatusTooManyRequests), http.StatusTooManyRequests)
			return
		}
		next.ServeHTTP(w, r)
	})
}
