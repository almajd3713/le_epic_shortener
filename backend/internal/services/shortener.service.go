package services

import (
	"log/slog"

	nanoid "github.com/matoous/go-nanoid/v2"
	"shortener.reeler.com/backend/internal/models"
	"shortener.reeler.com/backend/internal/repository"
)

type ShortenerService struct {
	repo   repository.URLRepository
	logger *slog.Logger
}

func NewShortenerService(repo repository.URLRepository, logger *slog.Logger) *ShortenerService {
	return &ShortenerService{repo: repo, logger: logger}
}

func (s *ShortenerService) ShortenURL(longUrl string) (*models.URL, error) {
	s.logger.Debug("shortening URL", "long_url", longUrl)

	var code string
	var err error
	for {
		code, err = nanoid.New(8)
		if err != nil {
			s.logger.Error("failed to generate short code", "error", err)
			return nil, err
		}

		if _, err := s.repo.GetByCode(code); err != nil {
			if err.Error() == "URL not found or expired" {
				break
			}
			return nil, err
		}
	}

	// Store code to DB
	newUrl, err := s.repo.Create(code, longUrl, nil)
	if err != nil {
		s.logger.Error("failed to create URL entry", "error", err)
		return nil, err
	}

	s.logger.Info("URL shortened successfully", "short_code", code)
	return newUrl, nil
}
