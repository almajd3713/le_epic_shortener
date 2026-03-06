package services

import (
	"shortener.reeler.com/backend/internal/repository"
)

type RedirectorService struct {
	repo repository.URLRepository
}

func NewRedirectorService(repo repository.URLRepository) *RedirectorService {
	return &RedirectorService{repo: repo}
}

func (r *RedirectorService) Redirect(shortCode string) (string, error) {
	url, err := r.GetOriginalURL(shortCode)
	if err != nil {
		return "", err
	}
	return url, nil
}

func (r *RedirectorService) GetOriginalURL(shortenedURL string) (string, error) {
	url, err := r.repo.GetByCode(shortenedURL)
	if err != nil {
		return "", err
	}
	return url.LongURL, nil
}