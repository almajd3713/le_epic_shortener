package services

import (
	"context"
	"log/slog"
	"time"

	"github.com/redis/go-redis/v9"
)

type cacheService struct {
	logger *slog.Logger
	cacheClient *redis.Client
}

func NewCacheService(ctx context.Context, cacheClient *redis.Client, logger *slog.Logger) (*cacheService, error) {
	return &cacheService{
		logger:      logger,
		cacheClient: cacheClient,
	}, nil
}

func (c *cacheService) Get(ctx context.Context, key string) (string, error) {
	c.logger.Debug("getting value from cache", "key", key)
	val, err := c.cacheClient.Get(ctx, key).Result()
	if err == redis.Nil {
		c.logger.Debug("cache miss", "key", key)
		return "", nil
	} else if err != nil {
		c.logger.Error("cache error", "error", err)
		return "", err
	}
	c.logger.Debug("cache hit", "key", key)
	return val, nil
}

func (c *cacheService) Set(ctx context.Context, key string, value string, expiresAt time.Duration) error {
	timeLeft := time.Until(time.Now().Add(expiresAt))
	c.logger.Debug("calculated time left for cache expiration", "time_left", timeLeft)
	c.logger.Debug("setting value in cache", "key", key, "expires_at", timeLeft)
	err := c.cacheClient.Set(ctx, key, value, expiresAt).Err()
	if err != nil {
		c.logger.Error("cache error", "error", err)
		return err
	}
	return nil
}

func (c *cacheService) Delete(ctx context.Context, key string) error {
	c.logger.Debug("deleting value from cache", "key", key)
	err := c.cacheClient.Del(ctx, key).Err()
	if err != nil {
		c.logger.Error("cache error", "error", err)
		return err
	}
	return nil
}

func (c *cacheService) Flush(ctx context.Context) error {
	c.logger.Debug("flushing cache")
	err := c.cacheClient.FlushDB(ctx).Err()
	if err != nil {
		c.logger.Error("cache error", "error", err)
		return err
	}
	return nil
}
