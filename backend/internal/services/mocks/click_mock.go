package mocks

import (
	"context"

	"shortener.reeler.com/backend/internal/models"
	"shortener.reeler.com/backend/internal/repository/mocks"
)

type MockClickService struct {
	repo *mocks.MockClickRepo
}

func NewMockClickService() *MockClickService {
	return &MockClickService{repo: mocks.NewMockClickRepo()}
}

func (m *MockClickService) RecordClick(c context.Context, code string) error {
	_, err := m.repo.Create(code)
	return err
}

func (m *MockClickService) GetClicksByCode(c context.Context, code string) ([]*models.Click, error) {
	return m.repo.GetByCode(code)
}
