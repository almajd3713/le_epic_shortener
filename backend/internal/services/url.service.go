package services

import (
	"log/slog"

	"shortener.reeler.com/backend/internal/repository"
	"shortener.reeler.com/backend/internal/models"
)

type URLService struct {
	repo   repository.URLRepository
	logger *slog.Logger
}

func NewURLService(repo repository.URLRepository, logger *slog.Logger) *URLService {
	return &URLService{repo: repo, logger: logger}
}

func (s *URLService) GetAllURLs() ([]models.URL, error) {
	urls, err := s.repo.GetAll()
	if err != nil {
		s.logger.Error("failed to get all URLs", "error", err)
		return nil, err
	}
	return urls, nil
}

func (s *URLService) GetOriginalURL(shortenedURL string) (string, error) {
	url, err := s.repo.GetByCode(shortenedURL)
	if err != nil {
		s.logger.Error("failed to get original URL", "error", err)
		return "", err
	}
	return url.LongURL, nil
}