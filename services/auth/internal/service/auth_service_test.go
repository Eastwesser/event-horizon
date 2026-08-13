package service

import (
	"context"
	"errors"
	"testing"

	"golang.org/x/crypto/bcrypt"

	"github.com/Eastwesser/event-horizon/services/auth/internal/model"
	"github.com/Eastwesser/event-horizon/services/auth/internal/repository/mocks"
)

func newTestService(repo *mocks.UserRepository) AuthService {
	return NewAuthServiceLegacy(repo, nil, "test-secret-key-for-jwt", 1)
}

func TestRegisterSuccess(t *testing.T) {
	repo := &mocks.UserRepository{
		GetByEmailFn: func(ctx context.Context, email string) (*model.User, error) {
			return nil, nil
		},
		CreateFn: func(ctx context.Context, email, passwordHash, role string) (string, error) {
			if role != "user" || email != "user@example.com" || passwordHash == "" {
				t.Fatalf("unexpected create args: %s %s %s", email, role, passwordHash)
			}
			return "uid-1", nil
		},
	}
	id, role, err := newTestService(repo).Register(context.Background(), "user@example.com", "password123", "")
	if err != nil {
		t.Fatal(err)
	}
	if id != "uid-1" || role != "user" {
		t.Fatalf("got %s %s", id, role)
	}
}

func TestRegisterAlreadyExists(t *testing.T) {
	repo := &mocks.UserRepository{
		GetByEmailFn: func(ctx context.Context, email string) (*model.User, error) {
			return &model.User{ID: "x", Email: email}, nil
		},
	}
	_, _, err := newTestService(repo).Register(context.Background(), "user@example.com", "password123", "")
	if !errors.Is(err, model.ErrUserAlreadyExists) {
		t.Fatalf("want ErrUserAlreadyExists, got %v", err)
	}
}

func TestRegisterInvalidRole(t *testing.T) {
	repo := &mocks.UserRepository{
		GetByEmailFn: func(ctx context.Context, email string) (*model.User, error) {
			return nil, nil
		},
	}
	_, _, err := newTestService(repo).Register(context.Background(), "user@example.com", "password123", "admin")
	if !errors.Is(err, model.ErrInvalidRole) {
		t.Fatalf("want ErrInvalidRole, got %v", err)
	}
}

func TestLoginSuccess(t *testing.T) {
	hash, err := bcrypt.GenerateFromPassword([]byte("password123"), bcryptCost)
	if err != nil {
		t.Fatal(err)
	}
	repo := &mocks.UserRepository{
		GetByEmailFn: func(ctx context.Context, email string) (*model.User, error) {
			return &model.User{ID: "uid-1", Email: email, PasswordHash: string(hash), Role: "author"}, nil
		},
	}
	pair, err := newTestService(repo).Login(context.Background(), "user@example.com", "password123")
	if err != nil {
		t.Fatal(err)
	}
	if pair.AccessToken == "" || pair.RefreshToken == "" || pair.UserID != "uid-1" || pair.Role != "author" {
		t.Fatalf("unexpected login result: %+v", pair)
	}
}

func TestLoginBadPassword(t *testing.T) {
	hash, err := bcrypt.GenerateFromPassword([]byte("password123"), bcryptCost)
	if err != nil {
		t.Fatal(err)
	}
	repo := &mocks.UserRepository{
		GetByEmailFn: func(ctx context.Context, email string) (*model.User, error) {
			return &model.User{ID: "uid-1", Email: email, PasswordHash: string(hash), Role: "user"}, nil
		},
	}
	_, err = newTestService(repo).Login(context.Background(), "user@example.com", "wrong-password")
	if !errors.Is(err, model.ErrInvalidCredentials) {
		t.Fatalf("want ErrInvalidCredentials, got %v", err)
	}
}

func TestUpdateRoleInvalid(t *testing.T) {
	err := newTestService(&mocks.UserRepository{}).UpdateRole(context.Background(), "uid-1", "superadmin")
	if !errors.Is(err, model.ErrInvalidRole) {
		t.Fatalf("want ErrInvalidRole, got %v", err)
	}
}

func TestUpdateRoleSuccess(t *testing.T) {
	called := false
	repo := &mocks.UserRepository{
		UpdateRoleFn: func(ctx context.Context, userID, role string) error {
			called = true
			if userID != "uid-1" || role != "admin" {
				t.Fatalf("bad args")
			}
			return nil
		},
	}
	if err := newTestService(repo).UpdateRole(context.Background(), "uid-1", "admin"); err != nil {
		t.Fatal(err)
	}
	if !called {
		t.Fatal("UpdateRole not called")
	}
}

func TestGetUserFromRepo(t *testing.T) {
	repo := &mocks.UserRepository{
		GetByIDFn: func(ctx context.Context, id string) (*model.User, error) {
			return &model.User{ID: id, Email: "a@b.c", Role: "user"}, nil
		},
	}
	u, err := newTestService(repo).GetUser(context.Background(), "uid-1")
	if err != nil {
		t.Fatal(err)
	}
	if u.ID != "uid-1" {
		t.Fatalf("got %s", u.ID)
	}
}
