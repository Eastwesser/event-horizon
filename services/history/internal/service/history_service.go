package service

import (
	"context"
	"encoding/json"
	"time"

	"github.com/Eastwesser/event-horizon/services/history/internal/model"
	"github.com/Eastwesser/event-horizon/services/history/internal/repository"
)

type HistoryService struct {
	repo          *repository.PostgresRepo
	retentionDays int
}

func New(repo *repository.PostgresRepo, retentionDays int) *HistoryService {
	if retentionDays <= 0 {
		retentionDays = 30
	}
	return &HistoryService{repo: repo, retentionDays: retentionDays}
}

func (s *HistoryService) RecordEvent(ctx context.Context, userID, eventType, payload string) (string, error) {
	if eventType == "" {
		return "", model.ErrInvalidInput
	}
	if payload == "" {
		payload = "{}"
	}
	if !json.Valid([]byte(payload)) {
		b, _ := json.Marshal(map[string]string{"raw": payload})
		payload = string(b)
	}
	return s.repo.Insert(ctx, userID, eventType, payload)
}

func (s *HistoryService) ListEvents(ctx context.Context, userID, eventType string, limit, offset int) ([]*model.Event, int64, error) {
	return s.repo.List(ctx, userID, eventType, limit, offset)
}

func (s *HistoryService) PurgeExpired(ctx context.Context) (int64, error) {
	before := time.Now().UTC().AddDate(0, 0, -s.retentionDays)
	return s.repo.DeleteOlderThan(ctx, before)
}
