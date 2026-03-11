package mocks

import (
	"context"

	"shortener.reeler.com/backend/internal/repository/mocks"
	"shortener.reeler.com/backend/internal/models"
)

type MockClickService struct {
	repo *mocks.MockClickRepo
}

func NewMockClickService() *MockClickService {
	return &MockClickService{repo: mocks.NewMockClickRepo()}
}

func (m *MockClickService) RecordClick(c context.Context, code string) error {
	return m.repo.Create(code)
}

func (m *MockClickService) GetClicksByCode(c context.Context, code string) ([]*models.Click, error) {
	return m.repo.GetByCode(code)
}