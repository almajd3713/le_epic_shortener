package services

import (
	"context"
	"log/slog"
	"os"
	"testing"

	urlMocks "shortener.reeler.com/backend/internal/repository/mocks"
	serviceMocks "shortener.reeler.com/backend/internal/services/mocks"
)

func TestRedirectorService_Redirect(t *testing.T) {
	mockURLSvc := urlMocks.NewMockURLRepo()
	mockCache := serviceMocks.NewMockCacheService()
	mockClickSvc := serviceMocks.NewMockClickService()
	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))

	redirector := NewRedirectorService(mockURLSvc, mockClickSvc, mockCache, logger)

	// First, create a short code for testing
	shortCode := "abc123"
	longURL := "https://www.example.com/some/very/long/url"
	_, err := mockURLSvc.Create(shortCode, longURL, nil)
	if err != nil {
		t.Fatalf("failed to create mock URL: %v", err)
	}

	// Test redirecting the short code
	redirectedURL, err := redirector.Redirect(context.Background(), shortCode)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if redirectedURL != longURL {
		t.Errorf("expected redirected URL to be %s, got %s", longURL, redirectedURL)
	}

	// Test cache functionality
	cachedURL, err := mockCache.Get(context.Background(), shortCode)
	if err != nil {
		t.Fatalf("expected no error when getting from cache, got %v", err)
	}
	if cachedURL != longURL {
		t.Errorf("expected cached URL to be %s, got %s", longURL, cachedURL)
	}
}