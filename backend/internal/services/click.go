package services

import (
	"context"
	"log/slog"

	"shortener.reeler.com/backend/internal/models"
	"shortener.reeler.com/backend/internal/repository"
)

type ClickService struct {
	repo repository.IClickRepository
	logger *slog.Logger
}

func NewClickService(repo repository.IClickRepository, logger *slog.Logger) *ClickService {
	return &ClickService{repo: repo, logger: logger}
}

func (s *ClickService) RecordClick(c context.Context, code string) error {
	s.logger.Debug("recording click", "code", code)
	_, err := s.repo.Create(code)
	if err != nil {
		s.logger.Error("failed to record click", "error", err)
		return err
	}
	return nil
}

func (s *ClickService) GetClicksByCode(c context.Context, code string) ([]*models.Click, error) {
	s.logger.Debug("fetching clicks by code", "code", code)
	return s.repo.GetByCode(code)
}

func (s *ClickService) GetClickCountByCode(c context.Context, code string) (int, error) {
	s.logger.Debug("fetching click count by code", "code", code)
	return s.repo.GetCountByCode(code)
}
