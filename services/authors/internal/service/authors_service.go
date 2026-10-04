package service

import (
	"context"
	"errors"
	"strings"
	"time"

	"github.com/google/uuid"

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

func (s *AuthorsService) SubmitApplication(
	ctx context.Context,
	userID, callerRole, displayName, portfolio, motivation, contactEmail string,
) (*model.AuthorApplication, error) {
	userID = strings.TrimSpace(userID)
	displayName = strings.TrimSpace(displayName)
	portfolio = strings.TrimSpace(portfolio)
	motivation = strings.TrimSpace(motivation)
	contactEmail = strings.TrimSpace(contactEmail)
	role := strings.ToLower(strings.TrimSpace(callerRole))

	if userID == "" || displayName == "" || motivation == "" || contactEmail == "" {
		return nil, model.ErrInvalidInput
	}
	if role == "author" || role == "admin" {
		return nil, model.ErrAlreadyPrivileged
	}

	if pending, err := s.repo.GetPendingApplicationByUserID(ctx, userID); err == nil {
		return pending, nil
	} else if !errors.Is(err, model.ErrApplicationMissing) {
		return nil, err
	}

	app := &model.AuthorApplication{
		ID:     uuid.NewString(),
		UserID: userID,
		Status: model.ApplicationPending,
		Payload: model.ApplicationPayload{
			DisplayName:  displayName,
			Portfolio:    portfolio,
			Motivation:   motivation,
			ContactEmail: contactEmail,
		},
		CreatedAt: time.Now().UTC(),
	}
	if err := s.repo.InsertApplication(ctx, app); err != nil {
		// Race: another pending insert — return existing.
		if pending, getErr := s.repo.GetPendingApplicationByUserID(ctx, userID); getErr == nil {
			return pending, nil
		}
		return nil, err
	}
	return app, nil
}

func (s *AuthorsService) GetMyApplication(ctx context.Context, userID string) (*model.AuthorApplication, error) {
	userID = strings.TrimSpace(userID)
	if userID == "" {
		return nil, model.ErrInvalidInput
	}
	return s.repo.GetLatestApplicationByUserID(ctx, userID)
}
