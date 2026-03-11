package mocks

import (
	"errors"
	"time"

	"shortener.reeler.com/backend/internal/models"
)

type MockClickRepo struct {
	store map[string]int
}

func NewMockClickRepo() *MockClickRepo {
	return &MockClickRepo{store: make(map[string]int)}
}

func (m *MockClickRepo) Create(shortCode string) (*models.Click, error) {
	m.store[shortCode]++
	return &models.Click{
		ID:        int64(m.store[shortCode]),
		ShortCode: shortCode,
		ClickedAt: time.Now(),
	}, nil
}

func (m *MockClickRepo) GetByCode(code string) ([]*models.Click, error) {
	if count, exists := m.store[code]; exists {
		clicks := make([]*models.Click, count)
		for i := 0; i < count; i++ {
			clicks[i] = &models.Click{
				ID:        int64(i + 1),
				ShortCode: code,
				ClickedAt: time.Now().Add(-time.Duration(i) * time.Minute),
			}
		}
		return clicks, nil
	}
	return nil, errors.New("no clicks found for code")
}

func (m *MockClickRepo) GetCountByCode(code string) (int, error) {
	if count, exists := m.store[code]; exists {
		return count, nil
	}
	return 0, nil
}
