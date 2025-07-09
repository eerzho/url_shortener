package middleware

import (
	"net/http"
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

func (l *lruCache) getOrPut(key string) *rate.Limiter {
	l.mu.Lock()
	defer l.mu.Unlock()

	item, ok := l.data[key]
	if ok {
		l.remove(item)
		l.addToHead(item)
		return item.value
	}

	if len(l.data) >= l.capacity {
		delete(l.data, l.tail.prev.key)
		l.remove(l.tail.prev)
	}

	value := rate.NewLimiter(rate.Limit(10), 20)
	item = &node{key: key, value: value}
	l.addToHead(item)
	l.data[key] = item

	return value
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

var (
	cache     *lruCache
	cacheOnce sync.Once
)

func getCache() *lruCache {
	cacheOnce.Do(func() {
		cache = newLruCache(10_000)
	})
	return cache
}

func RateLimit(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ip := getIp(r)

		limiter := getCache().getOrPut(ip)
		if !limiter.Allow() {
			http.Error(w, http.StatusText(http.StatusTooManyRequests), http.StatusTooManyRequests)
			return
		}

		next.ServeHTTP(w, r)
	})
}
