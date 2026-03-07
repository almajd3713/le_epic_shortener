package services

import (
	"log/slog"
	"context"

	"shortener.reeler.com/backend/internal/models"
	"shortener.reeler.com/backend/internal/repository"
)

type URLService struct {
	repo   repository.URLRepository
	cacheSvc ICacheService
	logger *slog.Logger
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
	

	url, err := s.repo.GetByCode(shortenedURL)
	if err != nil {
		s.logger.Error("failed to get original URL", "error", err)
		return "", err
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