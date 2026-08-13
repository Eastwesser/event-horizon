package service

import (
	"context"
	"log"
	"time"

	"github.com/Eastwesser/event-horizon/services/profile/internal/repository"
)

type ProfileService interface {
	GetProfile(ctx context.Context, userID string) (*repository.UserProfile, error)
	UpdateProfile(ctx context.Context, profile *repository.UserProfile) error
}

type profileService struct {
	repo     repository.ProfileRepository
	cache    *repository.RedisProfileRepo
	cacheTTL time.Duration
}

func NewProfileService(repo repository.ProfileRepository, cache *repository.RedisProfileRepo, cacheTTL time.Duration) ProfileService {
	return &profileService{repo: repo, cache: cache, cacheTTL: cacheTTL}
}

func (s *profileService) GetProfile(ctx context.Context, userID string) (*repository.UserProfile, error) {
	if s.cache != nil {
		cached, err := s.cache.GetProfile(ctx, userID)
		if err != nil {
			log.Printf("profile cache read error for %s: %v", userID, err)
		} else if cached != nil {
			return cached, nil
		}
	}

	profile, err := s.repo.GetProfile(ctx, userID)
	if err != nil || profile == nil {
		return profile, err
	}

	if s.cache != nil {
		if err := s.cache.SetProfile(ctx, profile, s.cacheTTL); err != nil {
			log.Printf("profile cache write error for %s: %v", userID, err)
		}
	}

	return profile, nil
}

func (s *profileService) UpdateProfile(ctx context.Context, profile *repository.UserProfile) error {
	if err := s.repo.UpsertProfile(ctx, profile); err != nil {
		return err
	}

	if s.cache != nil {
		if err := s.cache.InvalidateProfile(ctx, profile.UserID); err != nil {
			log.Printf("profile cache invalidate error for %s: %v", profile.UserID, err)
		}
	}

	return nil
}
