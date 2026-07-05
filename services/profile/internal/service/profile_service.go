package service

import (
    "context"

    "github.com/Eastwesser/event-horizon/services/profile/internal/repository"
)

type ProfileService interface {
    GetProfile(ctx context.Context, userID string) (*repository.UserProfile, error)
    UpdateProfile(ctx context.Context, profile *repository.UserProfile) error
}

type profileService struct {
    repo repository.ProfileRepository
}

func NewProfileService(repo repository.ProfileRepository) ProfileService {
    return &profileService{repo: repo}
}

func (s *profileService) GetProfile(ctx context.Context, userID string) (*repository.UserProfile, error) {
    return s.repo.GetProfile(ctx, userID)
}

func (s *profileService) UpdateProfile(ctx context.Context, profile *repository.UserProfile) error {
    return s.repo.UpsertProfile(ctx, profile)
}
