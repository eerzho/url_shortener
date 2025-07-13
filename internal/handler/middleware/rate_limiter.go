package middleware

import (
	"net/http"
	inmemory "url_shortener/pkg/in_memory"

	"golang.org/x/time/rate"
)

const (
	DefaultRateLimit     = 10
	DefaultBurstSize     = 20
	DefaultCacheCapacity = 1_000
)

type RateLimiter struct {
	ipService IPService
	rateLimit int
	burstSize int
	cache     *inmemory.LruCache[*rate.Limiter]
}

func NewRateLimiter(ipService IPService) *RateLimiter {
	return &RateLimiter{
		ipService: ipService,
		rateLimit: DefaultRateLimit,
		burstSize: DefaultBurstSize,
		cache:     inmemory.NewLruCache[*rate.Limiter](DefaultCacheCapacity),
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
