package inmemory

import (
	"sync"
)

const (
	DefaultCapacity = 100
)

type node[V any] struct {
	key   string
	value V
	prev  *node[V]
	next  *node[V]
}

type LruCache[V any] struct {
	mu       sync.Mutex
	capacity int
	data     map[string]*node[V]
	head     *node[V]
	tail     *node[V]
}

func NewLruCache[V any](capacity int) *LruCache[V] {
	capacity = max(capacity, DefaultCapacity)

	head := &node[V]{}
	tail := &node[V]{}

	head.next = tail
	tail.prev = head

	return &LruCache[V]{
		capacity: capacity,
		data:     make(map[string]*node[V]),
		head:     head,
		tail:     tail,
	}
}

func (l *LruCache[V]) Put(key string, value V) {
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

	item = &node[V]{key: key, value: value}
	l.addToHead(item)
	l.data[key] = item
}

func (l *LruCache[V]) Get(key string) (V, bool) {
	l.mu.Lock()
	defer l.mu.Unlock()

	item, ok := l.data[key]
	if !ok {
		var zero V
		return zero, false
	}

	l.remove(item)
	l.addToHead(item)

	return item.value, true
}

func (l *LruCache[V]) addToHead(item *node[V]) {
	tmp := l.head.next

	tmp.prev = item
	item.next = tmp

	l.head.next = item
	item.prev = l.head
}

func (l *LruCache[V]) remove(item *node[V]) {
	prev := item.prev
	next := item.next

	prev.next = next
	next.prev = prev
}
