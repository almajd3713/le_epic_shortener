package services

import (
	"context"
	"log/slog"
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
	// Check cache first
	cachedURL, err := r.cacheSvc.Get(c, shortCode)
	if err != nil {
		r.logger.Error("cache error during redirect", "error", err)
	} else if cachedURL != "" {
		r.logger.Debug("cache hit during redirect", "short_code", shortCode)
		return cachedURL, nil
	} else {
		r.logger.Debug("cache miss during redirect", "short_code", shortCode)
	}

	// Cache miss, get from DB
	url, err := r.urlSvc.GetOriginalURL(c, shortCode)
	if err != nil {
		r.logger.Error("failed to redirect URL", "error", err)
		return "", err
	}

	// Store in cache for future requests
	err = r.cacheSvc.Set(c, shortCode, url, 24*3600) // Cache for 24 hours
	
	
	return url, nil
}
