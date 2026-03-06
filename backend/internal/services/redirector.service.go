package services

import (
	"log/slog"

	"shortener.reeler.com/backend/internal/repository"
)

type RedirectorService struct {
	repo   repository.URLRepository
	logger *slog.Logger
}

func NewRedirectorService(repo repository.URLRepository, logger *slog.Logger) *RedirectorService {
	return &RedirectorService{repo: repo, logger: logger}
}

func (r *RedirectorService) Redirect(shortCode string) (string, error) {
	r.logger.Debug("redirecting short code", "short_code", shortCode)
	url, err := r.GetOriginalURL(shortCode)
	if err != nil {
		r.logger.Error("failed to redirect URL", "error", err)
		return "", err
	}
	r.logger.Info("redirect successful", "original_url", url)
	return url, nil
}

func (r *RedirectorService) GetOriginalURL(shortenedURL string) (string, error) {
	url, err := r.repo.GetByCode(shortenedURL)
	if err != nil {
		r.logger.Error("failed to get original URL", "error", err)
		return "", err
	}
	return url.LongURL, nil
}
