package cache

import (
	"errors"
	"time"
)

// Item represents a cache item with its value and metadata
type Item[T any] struct {
	Value      T
	Expiration *time.Time
}

// Cache interface defines the standard operations for a cache implementation
type Cache[T any] interface {
	// Set stores a value in the cache with an optional TTL
	// If ttl is 0, the item never expires
	Set(key string, value T, ttl time.Duration) error

	// Get retrieves a value from the cache
	// Returns ErrKeyNotFound if the key doesn't exist
	Get(key string) (T, error)

	// GetWithExpiration retrieves both the value and its expiration time
	GetWithExpiration(key string) (T, *time.Time, error)

	// Has checks if a key exists in the cache
	Has(key string) bool

	// Delete removes a specific item from the cache
	Delete(key string) error

	// Clear removes all items from the cache
	Clear() error

	// GetMultiple retrieves multiple values from cache
	// Returns a map of found items and any error encountered
	GetMultiple(keys []string) (map[string]T, error)

	// SetMultiple stores multiple values in cache
	// All items will have the same TTL
	SetMultiple(items map[string]T, ttl time.Duration) error
}

// Common cache errors
var (
	ErrKeyNotFound  = errors.New("key not found in cache")
	ErrKeyExpired   = errors.New("key has expired")
	ErrInvalidTTL   = errors.New("invalid TTL value")
	ErrInvalidKey   = errors.New("invalid key")
	ErrInvalidValue = errors.New("invalid value")
)
