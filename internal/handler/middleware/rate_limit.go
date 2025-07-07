package middleware

import (
	"net"
	"net/http"
	"strings"
	"sync"

	"golang.org/x/time/rate"
)

const (
	defaultRateLimit = 10
	defaultBurst     = 20
	defaultCacheSize = 10_000
)

type safeLimiter struct {
	limiter *rate.Limiter
	mu      sync.RWMutex
}

func newSafeLimiter(r rate.Limit, b int) *safeLimiter {
	return &safeLimiter{
		limiter: rate.NewLimiter(r, b),
	}
}

func (s *safeLimiter) allow() bool {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.limiter.Allow()
}

type node struct {
	key   string
	value *safeLimiter
	prev  *node
	next  *node
}

type lruCache struct {
	mu       sync.RWMutex
	capacity int
	head     *node
	tail     *node
	data     map[string]*node
}

func newLruCache(capacity int) *lruCache {
	head := &node{}
	tail := &node{}

	head.next = tail
	tail.prev = head

	return &lruCache{
		capacity: capacity,
		data:     make(map[string]*node, capacity),
		head:     head,
		tail:     tail,
	}
}

func (l *lruCache) get(key string) (*safeLimiter, bool) {
	l.mu.RLock()
	node, exists := l.data[key]
	l.mu.RUnlock()

	if !exists {
		return nil, false
	}

	// Move to head under write lock
	l.mu.Lock()
	l.moveToHead(node)
	l.mu.Unlock()

	return node.value, true
}

func (l *lruCache) put(key string, value *safeLimiter) {
	l.mu.Lock()
	defer l.mu.Unlock()

	if node, exists := l.data[key]; exists {
		node.value = value
		l.moveToHead(node)
		return
	}

	newNode := &node{key: key, value: value}
	l.data[key] = newNode
	l.addToHead(newNode)

	if len(l.data) > l.capacity {
		tail := l.removeTail()
		delete(l.data, tail.key)
	}
}

func (l *lruCache) getOrPut(key string) *safeLimiter {
	if limiter, exists := l.get(key); exists {
		return limiter
	}

	limiter := newSafeLimiter(rate.Limit(defaultRateLimit), defaultBurst)
	l.put(key, limiter)
	return limiter
}

func (l *lruCache) addToHead(node *node) {
	node.prev = l.head
	node.next = l.head.next

	l.head.next.prev = node
	l.head.next = node
}

func (l *lruCache) removeNode(node *node) {
	node.prev.next = node.next
	node.next.prev = node.prev
}

func (l *lruCache) moveToHead(node *node) {
	l.removeNode(node)
	l.addToHead(node)
}

func (l *lruCache) removeTail() *node {
	lastNode := l.tail.prev
	l.removeNode(lastNode)
	return lastNode
}

var (
	cache     *lruCache
	cacheOnce sync.Once
)

func getCache() *lruCache {
	cacheOnce.Do(func() {
		cache = newLruCache(defaultCacheSize)
	})
	return cache
}

func RateLimit(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ip := getIP(r)
		if ip == "" {
			// If we can't get IP, allow the request but log it
			next.ServeHTTP(w, r)
			return
		}

		limiter := getCache().getOrPut(ip)
		if !limiter.allow() {
			http.Error(w, "Rate limit exceeded", http.StatusTooManyRequests)
			return
		}

		next.ServeHTTP(w, r)
	})
}

func getIP(r *http.Request) string {
	// Check X-Forwarded-For header
	if forwardedFor := r.Header.Get("X-Forwarded-For"); forwardedFor != "" {
		// Take the first IP from the list
		ips := strings.Split(forwardedFor, ",")
		for _, ip := range ips {
			ip = strings.TrimSpace(ip)
			if net.ParseIP(ip) != nil {
				return ip
			}
		}
	}

	// Check X-Real-IP header
	if realIP := r.Header.Get("X-Real-IP"); realIP != "" {
		if net.ParseIP(realIP) != nil {
			return realIP
		}
	}

	// Check X-Forwarded-Proto and CF-Connecting-IP for Cloudflare
	if cfIP := r.Header.Get("CF-Connecting-IP"); cfIP != "" {
		if net.ParseIP(cfIP) != nil {
			return cfIP
		}
	}

	// Fall back to RemoteAddr
	if ip, _, err := net.SplitHostPort(r.RemoteAddr); err == nil {
		if net.ParseIP(ip) != nil {
			return ip
		}
	}

	// Return RemoteAddr as-is if all else fails
	return r.RemoteAddr
}

// RateLimitWithConfig allows custom rate limiting configuration
func RateLimitWithConfig(rateLimit rate.Limit, burst int, cacheSize int) func(http.Handler) http.Handler {
	customCache := newLruCache(cacheSize)

	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			ip := getIP(r)
			if ip == "" {
				next.ServeHTTP(w, r)
				return
			}

			var limiter *safeLimiter
			if existingLimiter, exists := customCache.get(ip); exists {
				limiter = existingLimiter
			} else {
				limiter = newSafeLimiter(rateLimit, burst)
				customCache.put(ip, limiter)
			}

			if !limiter.allow() {
				http.Error(w, "Rate limit exceeded", http.StatusTooManyRequests)
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}
