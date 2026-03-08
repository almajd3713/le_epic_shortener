package mocks

import (
	"context"
	"time"

	"shortener.reeler.com/backend/internal/repository/mocks"
)

type MockURLService struct {
	repo     *mocks.MockURLRepo
	cacheSvc *MockCacheService
}

func (m *MockURLService) GetOriginalURL(c context.Context, shortenedURL string) (string, error) {
	if longURL, exists := m.cacheSvc.store[shortenedURL]; exists {
		return longURL, nil
	}
	url, err := m.repo.GetByCode(shortenedURL)
	if err != nil {
		return "", err
	}
	if url != nil {
		m.cacheSvc.Set(c, shortenedURL, url.LongURL, time.Hour)
		return url.LongURL, nil
	}
	return "", nil
}
