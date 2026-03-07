package services

import (
	"context"
	"log/slog"
	"time"

	"shortener.reeler.com/backend/internal/models"
	"shortener.reeler.com/backend/internal/repository"
)

type URLService struct {
	repo     repository.URLRepository
	cacheSvc ICacheService
	logger   *slog.Logger
}

func NewURLService(repo repository.URLRepository, cacheSvc ICacheService, logger *slog.Logger) *URLService {
	return &URLService{repo: repo, cacheSvc: cacheSvc, logger: logger}
}

func (s *URLService) GetAllURLs() ([]models.URL, error) {
	urls, err := s.repo.GetAll()
	if err != nil {
		s.logger.Error("failed to get all URLs", "error", err)
		return nil, err
	}
	return urls, nil
}

func (s *URLService) GetOriginalURL(c context.Context, shortenedURL string) (string, error) {
	// Cache attempt
	s.logger.Debug("attempting to get original URL from cache", "shortened_url", shortenedURL)
	cachedURL, err := s.cacheSvc.Get(c, shortenedURL)
	if err != nil {
		s.logger.Error("cache error", "error", err)
	} else if cachedURL != "" {
		s.logger.Debug("cache hit for original URL", "shortened_url", shortenedURL)
		return cachedURL, nil
	} else {
		s.logger.Debug("cache miss for original URL", "shortened_url", shortenedURL)
	}

	// DB fallback
	s.logger.Debug("getting original URL from database", "shortened_url", shortenedURL)
	url, err := s.repo.GetByCode(shortenedURL)
	if err != nil {
		s.logger.Error("failed to get original URL", "error", err)
		return "", err
	}

	// Cache the result for future requests
	s.logger.Debug("caching original URL", "shortened_url", shortenedURL)
	var ttl time.Duration
	if url.ExpiresAt != nil {
		ttl = time.Until(*url.ExpiresAt)
		if ttl <= 0 {
			// Already expired — don't cache it
			return url.LongURL, nil
		}
	}
	err = s.cacheSvc.Set(c, shortenedURL, url.LongURL, ttl)
	if err != nil {
		s.logger.Error("cache error", "error", err)
	}
	return url.LongURL, nil
}

func (s *URLService) ActivateURL(c context.Context, shortenedURL string) error {
	err := s.repo.ActivateByCode(shortenedURL)
	if err != nil {
		s.logger.Error("failed to activate URL", "error", err)
		return err
	}
	// Drop from cache to ensure fresh data on next redirect
	err = s.cacheSvc.Delete(c, shortenedURL)
	if err != nil {
		s.logger.Error("failed to delete URL from cache", "error", err)
	}
	return nil
}

func (s *URLService) DeactivateURL(c context.Context, shortenedURL string) error {
	err := s.repo.DeactivateByCode(shortenedURL)
	if err != nil {
		s.logger.Error("failed to deactivate URL", "error", err)
		return err
	}

	// Drop from cache
	err = s.cacheSvc.Delete(c, shortenedURL)
	if err != nil {
		s.logger.Error("failed to delete URL from cache", "error", err)
	}
	return nil
}

func (s *URLService) DeleteURL(c context.Context, shortenedURL string) error {
	// Delete from DB
	err := s.repo.DeleteByCode(shortenedURL)
	if err != nil {
		s.logger.Error("failed to delete URL", "error", err)
		return err
	}

	// Drop from cache
	err = s.cacheSvc.Delete(c, shortenedURL)
	if err != nil {
		s.logger.Error("failed to delete URL from cache", "error", err)
	}
	return nil
}
