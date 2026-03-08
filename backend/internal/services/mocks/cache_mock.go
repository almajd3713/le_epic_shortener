package mocks

import (
	"context"
	"time"	
)

type MockCacheService struct {
	store map[string]string
}

func NewMockCacheService() *MockCacheService {
	return &MockCacheService{store: make(map[string]string)}
}

func (m *MockCacheService) Get(ctx context.Context, key string) (string, error) {
	if val, exists := m.store[key]; exists {
		return val, nil
	}
	return "", nil
}

func (m *MockCacheService) Set(ctx context.Context, key string, value string, expiresAt time.Duration) error {
	m.store[key] = value
	return nil
}

func (m *MockCacheService) Delete(ctx context.Context, key string) error {
	delete(m.store, key)
	return nil
}

func (m *MockCacheService) Flush(ctx context.Context) error {
	m.store = make(map[string]string)
	return nil
}
