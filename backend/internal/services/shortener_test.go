package services

import (
	"context"
	"log/slog"
	"os"
	"testing"

	urlMocks "shortener.reeler.com/backend/internal/repository/mocks"
	serviceMocks"shortener.reeler.com/backend/internal/services/mocks"
)

func TestShortenerService_ShortenURL(t *testing.T) {
	mockRepo := urlMocks.NewMockURLRepo()
	mockCache := serviceMocks.NewMockCacheService()
	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))

	service := NewShortenerService(mockRepo, mockCache, logger)

	longURL := "https://www.example.com/some/very/long/url"
	url, err := service.ShortenURL(context.Background(), longURL, nil)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if url.LongURL != longURL {
		t.Errorf("expected long URL to be %s, got %s", longURL, url.LongURL)
	}
	if url.ShortCode == "" {
		t.Error("expected short code to be generated")
	}
}
