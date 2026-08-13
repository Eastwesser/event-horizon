package service

import (
	"context"
	"time"

	"github.com/Eastwesser/event-horizon/services/authors/internal/model"
	"github.com/Eastwesser/event-horizon/services/authors/internal/repository"
)

type AuthorsService struct {
	repo  *repository.PostgresRepo
	cache *repository.RedisRepo
}

func New(repo *repository.PostgresRepo, cache *repository.RedisRepo) *AuthorsService {
	return &AuthorsService{repo: repo, cache: cache}
}

func (s *AuthorsService) UpsertProfile(ctx context.Context, userID, displayName, bio, avatar string) (*model.Author, error) {
	if userID == "" || displayName == "" {
		return nil, model.ErrInvalidInput
	}
	a := &model.Author{
		UserID:      userID,
		DisplayName: displayName,
		Bio:         bio,
		AvatarURL:   avatar,
		Active:      true,
	}
	event := map[string]any{
		"event":        "author.upserted",
		"user_id":      userID,
		"display_name": displayName,
		"timestamp":    time.Now().Unix(),
	}
	if err := s.repo.Upsert(ctx, a, "author.upserted", event); err != nil {
		return nil, err
	}
	if s.cache != nil {
		_ = s.cache.Set(ctx, a)
	}
	return a, nil
}

func (s *AuthorsService) GetAuthor(ctx context.Context, userID string) (*model.Author, error) {
	if s.cache != nil {
		if a, err := s.cache.Get(ctx, userID); err == nil {
			return a, nil
		}
	}
	a, err := s.repo.GetByUserID(ctx, userID)
	if err != nil {
		return nil, err
	}
	if s.cache != nil {
		_ = s.cache.Set(ctx, a)
	}
	return a, nil
}

func (s *AuthorsService) ListAuthors(ctx context.Context, limit, offset int) ([]*model.Author, int64, error) {
	return s.repo.List(ctx, limit, offset)
}
