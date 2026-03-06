package services

import (
	nanoid "github.com/matoous/go-nanoid/v2"
	"shortener.reeler.com/backend/internal/repository"
)

type ShortenerService struct {
	repo repository.URLRepository
}

func NewShortenerService(repo repository.URLRepository) *ShortenerService {
	return &ShortenerService{repo: repo}
}

func (s *ShortenerService) ShortenURL(longUrl string) (string, error) {
	var code string
	for {
		code, err := nanoid.New(8)
		if err != nil {
			return "", err
		}

		if _, err := s.repo.GetByCode(code); err != nil {
			if err.Error() == "URL not found or expired" {
				break
			}
			return "", err
		}
	}

	return code, nil
}

func (s *ShortenerService) GetOriginalURL(shortenedURL string) (string, error) {
	url, err := s.repo.GetByCode(shortenedURL)
	if err != nil {
		return "", err
	}
	return url.LongURL, nil
}