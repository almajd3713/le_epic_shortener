package cache

import (
	"context"
	"time"

	"github.com/redis/go-redis/v9"
	"shortener.reeler.com/backend/internal/config"
)

type ICache interface {
	Set(ctx context.Context, key string, value interface{}, expiration time.Duration) error
	Get(ctx context.Context, key string) (string, error)
	Delete(ctx context.Context, key string) error
	Flush(ctx context.Context) error
}

type Cache struct {
	client *redis.Client
}

func NewCache(client *redis.Client) *Cache {
	return &Cache{client: client}
}

func (c *Cache) Get(ctx context.Context, key string) (string, error) {
	return c.client.Get(ctx, key).Result()
}

func (c *Cache) Set(ctx context.Context, key string, value interface{}, expiration time.Duration) error {
	return c.client.Set(ctx, key, value, expiration).Err()
}

func (c *Cache) Delete(ctx context.Context, key string) error {
	return c.client.Del(ctx, key).Err()
}

func (c *Cache) Flush(ctx context.Context) error {
	return c.client.FlushDB(ctx).Err()
}

func NewRedisClient(ctx context.Context, config config.CacheConfig) (*redis.Client, error) {
	options, err := redis.ParseURL(config.URL)
	if err != nil {
		return nil, err
	}

	options.MaxRetries = config.MaxRetries
	options.MinRetryBackoff = config.MinRetryBackoff
	options.MaxRetryBackoff = config.MaxRetryBackoff	

	client := redis.NewClient(options)

	// Test the connection
	if err := client.Ping(ctx).Err(); err != nil {
		return nil, err
	}

	return client, nil
}