package services

import (
	"context"
	"log/slog"
	"time"

	"github.com/redis/go-redis/v9"

	"shortener.reeler.com/backend/internal/cache"
)

type cacheService struct {
	logger *slog.Logger
	cacheClient cache.ICache
}

func NewCacheService(cacheClient cache.ICache, logger *slog.Logger) ICacheService {
	return &cacheService{
		logger:      logger,
		cacheClient: cacheClient,
	}
}

func (c *cacheService) Get(ctx context.Context, key string) (string, error) {
	c.logger.Debug("getting value from cache", "key", key)
	val, err := c.cacheClient.Get(ctx, key)
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
	c.logger.Debug("setting value in cache", "key", key, "expires_at", expiresAt)
	err := c.cacheClient.Set(ctx, key, value, expiresAt)
	if err != nil {
		c.logger.Error("cache error", "error", err)
		return err
	}
	return nil
}

func (c *cacheService) Delete(ctx context.Context, key string) error {
	c.logger.Debug("deleting value from cache", "key", key)
	err := c.cacheClient.Delete(ctx, key)
	if err != nil {
		c.logger.Error("cache error", "error", err)
		return err
	}
	return nil
}

func (c *cacheService) Flush(ctx context.Context) error {
	c.logger.Debug("flushing cache")
	err := c.cacheClient.Flush(ctx)
	if err != nil {
		c.logger.Error("cache error", "error", err)
		return err
	}
	return nil
}
