package middleware

import (
	"net/http"
	inmemory "url_shortener/pkg/in_memory"

	"golang.org/x/time/rate"
)

type RateLimiter struct {
	rateLimit int
	burstSize int
	cache     *inmemory.LruCache[*rate.Limiter]
}

func NewRateLimiter(cacheCapacity, rateLimit, burstSize int) *RateLimiter {
	return &RateLimiter{
		rateLimit: rateLimit,
		burstSize: burstSize,
		cache:     inmemory.NewLruCache[*rate.Limiter](cacheCapacity),
	}
}

func (rl *RateLimiter) Handle(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ip := getIp(r)

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
