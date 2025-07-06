package middleware

import (
	"net"
	"net/http"
	"strings"
	"sync"

	"golang.org/x/time/rate"
)

type node struct {
	key   string
	value *rate.Limiter
	prev  *node
	next  *node
}

type lruCache struct {
	mu       sync.Mutex
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
		data:     make(map[string]*node),
		head:     head,
		tail:     tail,
	}
}

func (l *lruCache) get(key string) *rate.Limiter {
	l.mu.Lock()
	defer l.mu.Unlock()

	item, ok := l.data[key]
	if !ok {
		return nil
	}

	l.remove(item)
	l.addToHead(item)

	return item.value
}

func (l *lruCache) put(key string, value *rate.Limiter) {
	l.mu.Lock()
	defer l.mu.Unlock()

	item, ok := l.data[key]
	if ok {
		item.value = value

		l.remove(item)
		l.addToHead(item)

		return
	}

	if len(l.data) >= l.capacity {
		delete(l.data, l.tail.prev.key)
		l.remove(l.tail.prev)
	}

	item = &node{key: key, value: value}
	l.addToHead(item)
	l.data[key] = item
}

func (l *lruCache) addToHead(item *node) {
	tmp := l.head.next

	tmp.prev = item
	item.next = tmp

	l.head.next = item
	item.prev = l.head
}

func (l *lruCache) remove(item *node) {
	prev := item.prev
	next := item.next

	prev.next = next
	next.prev = prev
}

var cache = newLruCache(10_000)

func RateLimit(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ip := getIp(r)

		limiter := cache.get(ip)
		if limiter == nil {
			limiter = rate.NewLimiter(rate.Limit(10), 20)
			cache.put(ip, limiter)
		}

		if !limiter.Allow() {
			http.Error(w, http.StatusText(http.StatusTooManyRequests), http.StatusTooManyRequests)
			return
		}

		next.ServeHTTP(w, r)
	})
}

func getIp(r *http.Request) string {
	if forwardedFor := r.Header.Get("X-Forwarded-For"); forwardedFor != "" {
		parts := strings.Split(forwardedFor, ",")
		ip := strings.TrimSpace(parts[0])
		if net.ParseIP(ip) != nil {
			return ip
		}
	}

	if realIp := r.Header.Get("X-Real-IP"); realIp != "" {
		if net.ParseIP(realIp) != nil {
			return realIp
		}
	}

	if ip, _, err := net.SplitHostPort(r.RemoteAddr); err == nil {
		return ip
	}

	return r.RemoteAddr
}
