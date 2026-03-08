package cache

import (
	"context"
	"github.com/redis/go-redis/v9"
	"shortener.reeler.com/backend/internal/config"
)

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