package cache

import (
	"context"
	"encoding/json"
	"fmt"
	"net/url"
	"reflect"
	"time"

	"github.com/redis/go-redis/v9"
)

// RedisCache implements Cache interface using Redis as backend
type RedisCache struct {
	client *redis.Client
	ctx    context.Context
}

// RedisConfig holds configuration for Redis connection
type RedisConfig struct {
	Host     string
	Port     string
	Password string
	DB       int
}

// NewRedisCache creates a new Redis cache instance
func NewRedisCache(connectionString string, db int) (*RedisCache, error) {
	u, err := url.Parse(connectionString)
	if err != nil {
		return nil, err
	}

	password, ok := u.User.Password()
	if !ok {
		return nil, fmt.Errorf("redis password does not exist")
	}

	client := redis.NewClient(&redis.Options{
		Addr:     u.Host,
		Password: password,
		DB:       db,
	})

	// Test connection
	ctx := context.Background()
	if err := client.Ping(ctx).Err(); err != nil {
		return nil, err
	}

	return &RedisCache{
		client: client,
		ctx:    ctx,
	}, nil
}

func (c *RedisCache) Set(key string, value interface{}, ttl time.Duration) error {
	if key == "" {
		return ErrInvalidKey
	}

	// Serialize value to JSON
	data, err := json.Marshal(value)
	if err != nil {
		return err
	}

	return c.client.Set(c.ctx, key, data, ttl).Err()
}

func (c *RedisCache) Get(key string, value interface{}) error {
	data, err := c.client.Get(c.ctx, key).Bytes()
	if err != nil {
		if err == redis.Nil {
			return ErrKeyNotFound
		}
		return err
	}

	return json.Unmarshal(data, value)
}

func (c *RedisCache) GetWithExpiration(key string, value interface{}) (*time.Time, error) {
	// Get the value
	data, err := c.client.Get(c.ctx, key).Bytes()
	if err != nil {
		if err == redis.Nil {
			return nil, ErrKeyNotFound
		}
		return nil, err
	}

	// Get TTL
	ttl, err := c.client.TTL(c.ctx, key).Result()
	if err != nil {
		return nil, err
	}

	// Unmarshal the value
	if err := json.Unmarshal(data, value); err != nil {
		return nil, err
	}

	var expTime *time.Time
	if ttl > 0 {
		t := time.Now().Add(ttl)
		expTime = &t
	}

	return expTime, nil
}

func (c *RedisCache) Has(key string) (bool, error) {
	exists, err := c.client.Exists(c.ctx, key).Result()
	if err != nil {
		return false, err
	}

	return exists > 0, nil
}

func (c *RedisCache) Delete(key string) error {
	return c.client.Del(c.ctx, key).Err()
}

func (c *RedisCache) Clear() error {
	return c.client.FlushDB(c.ctx).Err()
}

func (c *RedisCache) GetMultiple(keys []string, valueType reflect.Type) (map[string]interface{}, error) {
	result := make(map[string]interface{})

	// Use Redis pipeline for better performance
	pipe := c.client.Pipeline()
	cmds := make(map[string]*redis.StringCmd)

	for _, key := range keys {
		cmds[key] = pipe.Get(c.ctx, key)
	}

	_, err := pipe.Exec(c.ctx)
	if err != nil && err != redis.Nil {
		return nil, err
	}

	for key, cmd := range cmds {
		data, err := cmd.Bytes()
		if err == nil {
			var value interface{}
			if err := json.Unmarshal(data, &value); err == nil {
				result[key] = value
			}
		}
	}

	return result, nil
}

func (c *RedisCache) SetMultiple(items map[string]interface{}, ttl time.Duration) error {
	pipe := c.client.Pipeline()

	for key, value := range items {
		if key == "" {
			return ErrInvalidKey
		}

		data, err := json.Marshal(value)
		if err != nil {
			return err
		}

		pipe.Set(c.ctx, key, data, ttl)
	}

	_, err := pipe.Exec(c.ctx)
	return err
}

// Close closes the Redis connection
func (c *RedisCache) Close() error {
	return c.client.Close()
}
