package services

import (
	"log/slog"
)

type IURLService interface {
	GetOriginalURL(shortenedURL string) (string, error)
}

type RedirectorService struct {
	urlSvc IURLService
	logger *slog.Logger
}

func NewRedirectorService(urlSvc IURLService, logger *slog.Logger) *RedirectorService {
	return &RedirectorService{urlSvc: urlSvc, logger: logger}
}

func (r *RedirectorService) Redirect(shortCode string) (string, error) {
	r.logger.Debug("redirecting short code", "short_code", shortCode)
	url, err := r.urlSvc.GetOriginalURL(shortCode)
	if err != nil {
		r.logger.Error("failed to redirect URL", "error", err)
		return "", err
	}
	r.logger.Info("redirect successful", "original_url", url)
	return url, nil
}
