package middleware

import (
	"net/http"
	inmemory "url_shortener/pkg/in_memory"

	"golang.org/x/time/rate"
)

type RateLimiter struct {
	ipService IPService
	rateLimit int
	burstSize int
	cache     *inmemory.LruCache[*rate.Limiter]
}

func NewRateLimiter(
	ipService IPService,
	rateLimit int,
	burstSize int,
	cacheCapacity int,
) *RateLimiter {
	return &RateLimiter{
		ipService: ipService,
		rateLimit: rateLimit,
		burstSize: burstSize,
		cache:     inmemory.NewLruCache[*rate.Limiter](cacheCapacity),
	}
}

func (rl *RateLimiter) Handle(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ip := rl.ipService.GetIP(r.Context(), r)
		limiter, ok := rl.cache.Get(ip)
		if !ok {
			limiter = rate.NewLimiter(rate.Limit(rl.rateLimit), rl.burstSize)
			rl.cache.Put(ip, limiter)
		}
		if !limiter.Allow() {
			http.Error(w, http.StatusText(http.StatusTooManyRequests), http.StatusTooManyRequests)
			return
		}
		next.ServeHTTP(w, r)
	})
}
