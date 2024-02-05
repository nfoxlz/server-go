// caches
package components

import (
	"sync"
	"time"
)

type Cache[K comparable, V any] struct {
	data map[K]V
	lock sync.RWMutex
}

func (c *Cache[K, V]) Initialize() {
	c.data = make(map[K]V)
}

func (c *Cache[K, V]) Set(key K, value V, timeout time.Duration) {
	c.lock.Lock()
	defer c.lock.Unlock()

	c.data[key] = value
	time.AfterFunc(timeout, func() {
		c.lock.Lock()
		defer c.lock.Unlock()
		delete(c.data, key)
	})
}

func (c *Cache[K, V]) Get(key K) (V, bool) {
	c.lock.RLock()
	defer c.lock.RUnlock()

	value, ok := c.data[key]
	return value, ok
}

type CacheHandler[K comparable, V any] func(*K) (V, error)

func (c *Cache[K, V]) TryGetWithSetTimeout(key K, timeout time.Duration, handler CacheHandler[K, V]) (V, error) {
	if value, ok := c.Get(key); ok {
		return value, nil
	}

	value, err := handler(&key)
	if nil == err {
		c.Set(key, value, timeout)
	}
	return value, err
}

func (c *Cache[K, V]) TryGet(key K, handler CacheHandler[K, V]) (V, error) {
	return c.TryGetWithSetTimeout(key, 2*time.Hour, handler)
}
