package cache

import (
	"sync"
	"time"
)

// MemoryCache implements Cache interface with in-memory storage
type MemoryCache[T any] struct {
	items map[string]Item[T]
	mu    sync.RWMutex
}

// NewMemoryCache creates a new in-memory cache instance
func NewMemoryCache[T any]() *MemoryCache[T] {
	return &MemoryCache[T]{
		items: make(map[string]Item[T]),
	}
}

func (c *MemoryCache[T]) Set(key string, value T, ttl time.Duration) error {
	if key == "" {
		return ErrInvalidKey
	}

	c.mu.Lock()
	defer c.mu.Unlock()

	var exp *time.Time
	if ttl > 0 {
		expTime := time.Now().Add(ttl)
		exp = &expTime
	}

	c.items[key] = Item[T]{
		Value:      value,
		Expiration: exp,
	}
	return nil
}

func (c *MemoryCache[T]) Get(key string) (T, error) {
	c.mu.RLock()
	defer c.mu.RUnlock()

	item, exists := c.items[key]
	if !exists {
		var zero T
		return zero, ErrKeyNotFound
	}

	if item.Expiration != nil && time.Now().After(*item.Expiration) {
		var zero T
		delete(c.items, key)
		return zero, ErrKeyExpired
	}

	return item.Value, nil
}

func (c *MemoryCache[T]) GetWithExpiration(key string) (T, *time.Time, error) {
	c.mu.RLock()
	defer c.mu.RUnlock()

	item, exists := c.items[key]
	if !exists {
		var zero T
		return zero, nil, ErrKeyNotFound
	}

	if item.Expiration != nil && time.Now().After(*item.Expiration) {
		var zero T
		delete(c.items, key)
		return zero, nil, ErrKeyExpired
	}

	return item.Value, item.Expiration, nil
}

func (c *MemoryCache[T]) Has(key string) bool {
	c.mu.RLock()
	defer c.mu.RUnlock()

	item, exists := c.items[key]
	if !exists {
		return false
	}

	if item.Expiration != nil && time.Now().After(*item.Expiration) {
		delete(c.items, key)
		return false
	}

	return true
}

func (c *MemoryCache[T]) Delete(key string) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	delete(c.items, key)
	return nil
}

func (c *MemoryCache[T]) Clear() error {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.items = make(map[string]Item[T])
	return nil
}

func (c *MemoryCache[T]) GetMultiple(keys []string) (map[string]T, error) {
	result := make(map[string]T)

	c.mu.RLock()
	defer c.mu.RUnlock()

	for _, key := range keys {
		if item, exists := c.items[key]; exists {
			if item.Expiration == nil || time.Now().Before(*item.Expiration) {
				result[key] = item.Value
			}
		}
	}

	return result, nil
}

func (c *MemoryCache[T]) SetMultiple(items map[string]T, ttl time.Duration) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	var exp *time.Time
	if ttl > 0 {
		expTime := time.Now().Add(ttl)
		exp = &expTime
	}

	for key, value := range items {
		if key == "" {
			return ErrInvalidKey
		}
		c.items[key] = Item[T]{
			Value:      value,
			Expiration: exp,
		}
	}

	return nil
}
