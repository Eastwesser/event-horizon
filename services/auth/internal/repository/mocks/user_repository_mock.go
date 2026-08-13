package mocks

import (
	"context"

	"github.com/Eastwesser/event-horizon/services/auth/internal/model"
)

// UserRepository is a hand-written mock for unit tests.
// Prefer regenerating with mockery when tools/network are available:
//
//	//go:generate mockery --name=UserRepository ...
type UserRepository struct {
	CreateFn          func(ctx context.Context, email, passwordHash, role string) (string, error)
	GetByEmailFn      func(ctx context.Context, email string) (*model.User, error)
	GetByIDFn         func(ctx context.Context, id string) (*model.User, error)
	UpdateNicknameFn  func(ctx context.Context, userID, nickname string) error
	UpdateRoleFn      func(ctx context.Context, userID, role string) error
	GetUserScoresFn   func(ctx context.Context, userID string) (map[string]int32, int32, error)
}

func (m *UserRepository) Create(ctx context.Context, email, passwordHash, role string) (string, error) {
	return m.CreateFn(ctx, email, passwordHash, role)
}

func (m *UserRepository) GetByEmail(ctx context.Context, email string) (*model.User, error) {
	return m.GetByEmailFn(ctx, email)
}

func (m *UserRepository) GetByID(ctx context.Context, id string) (*model.User, error) {
	return m.GetByIDFn(ctx, id)
}

func (m *UserRepository) UpdateNickname(ctx context.Context, userID, nickname string) error {
	return m.UpdateNicknameFn(ctx, userID, nickname)
}

func (m *UserRepository) UpdateRole(ctx context.Context, userID, role string) error {
	return m.UpdateRoleFn(ctx, userID, role)
}

func (m *UserRepository) GetUserScores(ctx context.Context, userID string) (map[string]int32, int32, error) {
	return m.GetUserScoresFn(ctx, userID)
}
