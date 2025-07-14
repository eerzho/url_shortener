package middleware

import (
	"net/http"
	inmemory "url_shortener/pkg/in_memory"

	"golang.org/x/time/rate"
)

type RateLimiter struct {
	rps       int
	burst     int
	ipService IPService
	cache     *inmemory.LruCache[*rate.Limiter]
}

func NewRateLimiter(
	rps int,
	burst int,
	cacheCapacity int,
	ipService IPService,
) *RateLimiter {
	return &RateLimiter{
		rps:       rps,
		burst:     burst,
		ipService: ipService,
		cache:     inmemory.NewLruCache[*rate.Limiter](cacheCapacity),
	}
}

func (rl *RateLimiter) Handle(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ip := rl.ipService.GetIP(r.Context(), r)
		limiter, ok := rl.cache.Get(ip)
		if !ok {
			limiter = rate.NewLimiter(rate.Limit(rl.rps), rl.burst)
			rl.cache.Put(ip, limiter)
		}
		if !limiter.Allow() {
			http.Error(w, http.StatusText(http.StatusTooManyRequests), http.StatusTooManyRequests)
			return
		}
		next.ServeHTTP(w, r)
	})
}
