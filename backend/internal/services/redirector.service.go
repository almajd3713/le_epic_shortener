package services

import (
	"context"
	"log/slog"
	"time"
)



type RedirectorService struct {
	urlSvc IURLService
	cacheSvc ICacheService
	logger *slog.Logger
}

func NewRedirectorService(urlSvc IURLService, cacheSvc ICacheService, logger *slog.Logger) *RedirectorService {
	return &RedirectorService{urlSvc: urlSvc, cacheSvc: cacheSvc, logger: logger}
}

func (r *RedirectorService) Redirect(c context.Context, shortCode string) (string, error) {
	r.logger.Debug("redirecting short code", "short_code", shortCode)

	url, err := r.urlSvc.GetOriginalURL(c, shortCode)
	if err != nil {
		r.logger.Error("failed to redirect URL", "error", err)
		return "", err
	}

	// Store in cache for future requests
	err = r.cacheSvc.Set(c, shortCode, url, 24 * time.Hour) // Cache for 24 hours
	if err != nil {
		r.logger.Error("failed to cache original URL", "error", err)
	}

	return url, nil
}
