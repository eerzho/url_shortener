package middleware

import (
	"net/http"
	"sync"
	inmemory "url_shortener/pkg/in_memory"

	"golang.org/x/time/rate"
)

const (
	DefaultLruCacheSize = 1_000
	DefaultRate         = 10
	DefaultBurst        = 20
)

var (
	_lruCache     *inmemory.LruCache[*rate.Limiter]
	_lruCacheOnce sync.Once
)

func lruCache() *inmemory.LruCache[*rate.Limiter] {
	_lruCacheOnce.Do(func() {
		_lruCache = inmemory.New[*rate.Limiter](DefaultLruCacheSize)
	})
	return _lruCache
}

func RateLimit(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ip := getIp(r)

		limiter, ok := lruCache().Get(ip)
		if !ok {
			limiter = rate.NewLimiter(DefaultRate, DefaultBurst)
			lruCache().Put(ip, limiter)
		}

		if !limiter.Allow() {
			http.Error(w, http.StatusText(http.StatusTooManyRequests), http.StatusTooManyRequests)
			return
		}

		next.ServeHTTP(w, r)
	})
}
