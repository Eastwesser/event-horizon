package service

import (
	"context"

	"github.com/Eastwesser/event-horizon/services/analytics/internal/model"
	"github.com/Eastwesser/event-horizon/services/analytics/internal/repository"
)

type AnalyticsService struct {
	repo *repository.AnalyticsRepo
}

func New(repo *repository.AnalyticsRepo) *AnalyticsService {
	return &AnalyticsService{repo: repo}
}

func (s *AnalyticsService) RecordEvent(ctx context.Context, userID, eventType, payload string) error {
	if eventType == "" {
		return model.ErrInvalidInput
	}
	if payload == "" {
		payload = "{}"
	}
	return s.repo.Record(ctx, userID, eventType, payload)
}

func (s *AnalyticsService) GetDAU(ctx context.Context, days int) ([]model.DayCount, error) {
	return s.repo.DAU(ctx, days)
}

func (s *AnalyticsService) GetMAU(ctx context.Context, days int) (int64, int, error) {
	if days <= 0 {
		days = 30
	}
	n, err := s.repo.MAU(ctx, days)
	return n, days, err
}

func (s *AnalyticsService) GetRetention(ctx context.Context, cohortDaysAgo, windowDays int) (*model.Retention, error) {
	return s.repo.Retention(ctx, cohortDaysAgo, windowDays)
}
