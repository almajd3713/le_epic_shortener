package mocks

import (
	"context"
	"errors"
	"time"

	"shortener.reeler.com/backend/internal/models"
)

type MockURLRepo struct {
	store map[string]string
}

func NewMockURLRepo() *MockURLRepo {
	return &MockURLRepo{store: make(map[string]string)}
}

func (m *MockURLRepo) Create(shortCode, longURL string, expiresAt *time.Time) (*models.URL, error) {
	m.store[shortCode] = longURL
	return &models.URL{ShortCode: shortCode, LongURL: longURL}, nil
}

func (m *MockURLRepo) GetByCode(code string) (*models.URL, error) {
	if longURL, exists := m.store[code]; exists {
		return &models.URL{ShortCode: code, LongURL: longURL}, nil
	}
	return nil, errors.New("URL not found or expired")
}

func (m *MockURLRepo) GetAll() ([]models.URL, error) {
	var urls []models.URL
	for code, longURL := range m.store {
		urls = append(urls, models.URL{ShortCode: code, LongURL: longURL})
	}
	return urls, nil
}

func (m *MockURLRepo) GetOriginalURL(c context.Context, shortenedURL string) (string, error) {
	if longURL, exists := m.store[shortenedURL]; exists {
		return longURL, nil
	}
	return "", nil
}

func (m *MockURLRepo) ActivateByCode(code string) error {
	return nil
}

func (m *MockURLRepo) DeactivateByCode(code string) error {
	return nil
}

func (m *MockURLRepo) DeleteByCode(code string) error {
	delete(m.store, code)
	return nil
}
