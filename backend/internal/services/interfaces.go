package services

import (
	"context"
	"time"
)

type IURLService interface {
	GetOriginalURL(c context.Context, shortenedURL string) (string, error)
}

type ICacheService interface {
	Get(ctx context.Context, key string) (string, error)
	Set(ctx context.Context, key string, value string, expiresAt time.Duration) error
	Delete(ctx context.Context, key string) error
}