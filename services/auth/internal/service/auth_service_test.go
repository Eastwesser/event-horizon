package service

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/alicebob/miniredis/v2"
	"golang.org/x/crypto/bcrypt"

	jwtauth "github.com/Eastwesser/event-horizon/services/auth/internal/jwt"
	"github.com/Eastwesser/event-horizon/services/auth/internal/model"
	"github.com/Eastwesser/event-horizon/services/auth/internal/repository"
	"github.com/Eastwesser/event-horizon/services/auth/internal/repository/mocks"
)

const testJWTSecret = "test-secret-key-for-jwt-signing-tests"

func newTestService(repo *mocks.UserRepository) AuthService {
	return NewAuthServiceLegacy(repo, nil, testJWTSecret, 1)
}

func newTestServiceWithManager(repo *mocks.UserRepository) AuthService {
	mgr := jwtauth.NewManager(testJWTSecret, time.Hour, 24*time.Hour)
	return NewAuthService(repo, nil, mgr)
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

func TestRegisterDefaultRole(t *testing.T) {
	var gotRole string
	repo := &mocks.UserRepository{
		GetByEmailFn: func(ctx context.Context, email string) (*model.User, error) { return nil, nil },
		CreateFn: func(ctx context.Context, email, passwordHash, role string) (string, error) {
			gotRole = role
			return "uid-1", nil
		},
	}
	_, role, err := newTestService(repo).Register(context.Background(), "user@example.com", "password123", "")
	if err != nil || role != "user" || gotRole != "user" {
		t.Fatalf("role=%q gotRole=%q err=%v", role, gotRole, err)
	}
}

func TestLoginUserNotFound(t *testing.T) {
	repo := &mocks.UserRepository{
		GetByEmailFn: func(ctx context.Context, email string) (*model.User, error) { return nil, nil },
	}
	_, err := newTestService(repo).Login(context.Background(), "missing@example.com", "password123")
	if !errors.Is(err, model.ErrInvalidCredentials) {
		t.Fatalf("want ErrInvalidCredentials, got %v", err)
	}
}

func TestValidateTokenRoundTrip(t *testing.T) {
	hash, err := bcrypt.GenerateFromPassword([]byte("password123"), bcryptCost)
	if err != nil {
		t.Fatal(err)
	}
	repo := &mocks.UserRepository{
		GetByEmailFn: func(ctx context.Context, email string) (*model.User, error) {
			return &model.User{ID: "uid-1", Email: email, PasswordHash: string(hash), Role: "user"}, nil
		},
	}
	svc := newTestServiceWithManager(repo)
	pair, err := svc.Login(context.Background(), "user@example.com", "password123")
	if err != nil {
		t.Fatal(err)
	}
	uid, email, role, err := svc.ValidateToken(context.Background(), pair.AccessToken)
	if err != nil || uid != "uid-1" || email != "user@example.com" || role != "user" {
		t.Fatalf("uid=%s email=%s role=%s err=%v", uid, email, role, err)
	}
}

func TestValidateTokenInvalid(t *testing.T) {
	_, _, _, err := newTestServiceWithManager(&mocks.UserRepository{}).ValidateToken(context.Background(), "not-a-jwt")
	if !errors.Is(err, model.ErrInvalidToken) {
		t.Fatalf("want ErrInvalidToken, got %v", err)
	}
}

func TestRefreshTokenRoundTrip(t *testing.T) {
	hash, err := bcrypt.GenerateFromPassword([]byte("password123"), bcryptCost)
	if err != nil {
		t.Fatal(err)
	}
	repo := &mocks.UserRepository{
		GetByEmailFn: func(ctx context.Context, email string) (*model.User, error) {
			return &model.User{ID: "uid-1", Email: email, PasswordHash: string(hash), Role: "author"}, nil
		},
	}
	svc := newTestServiceWithManager(repo)
	pair, err := svc.Login(context.Background(), "user@example.com", "password123")
	if err != nil {
		t.Fatal(err)
	}
	refreshed, err := svc.RefreshToken(context.Background(), pair.RefreshToken)
	if err != nil {
		t.Fatal(err)
	}
	if refreshed.AccessToken == "" || refreshed.RefreshToken == "" || refreshed.UserID != "uid-1" {
		t.Fatalf("unexpected refresh: %+v", refreshed)
	}
}

func TestRefreshTokenInvalid(t *testing.T) {
	_, err := newTestServiceWithManager(&mocks.UserRepository{}).RefreshToken(context.Background(), "bad-token")
	if !errors.Is(err, model.ErrInvalidToken) {
		t.Fatalf("want ErrInvalidToken, got %v", err)
	}
}

func TestWhoami(t *testing.T) {
	hash, err := bcrypt.GenerateFromPassword([]byte("password123"), bcryptCost)
	if err != nil {
		t.Fatal(err)
	}
	repo := &mocks.UserRepository{
		GetByEmailFn: func(ctx context.Context, email string) (*model.User, error) {
			return &model.User{ID: "uid-1", Email: email, PasswordHash: string(hash), Role: "user"}, nil
		},
		GetByIDFn: func(ctx context.Context, id string) (*model.User, error) {
			return &model.User{ID: id, Email: "user@example.com", Role: "user"}, nil
		},
	}
	svc := newTestServiceWithManager(repo)
	pair, err := svc.Login(context.Background(), "user@example.com", "password123")
	if err != nil {
		t.Fatal(err)
	}
	u, err := svc.Whoami(context.Background(), pair.AccessToken)
	if err != nil || u.ID != "uid-1" {
		t.Fatalf("user=%+v err=%v", u, err)
	}
}

func TestLogoutNilCache(t *testing.T) {
	hash, err := bcrypt.GenerateFromPassword([]byte("password123"), bcryptCost)
	if err != nil {
		t.Fatal(err)
	}
	repo := &mocks.UserRepository{
		GetByEmailFn: func(ctx context.Context, email string) (*model.User, error) {
			return &model.User{ID: "uid-1", Email: email, PasswordHash: string(hash), Role: "user"}, nil
		},
	}
	svc := newTestServiceWithManager(repo)
	pair, err := svc.Login(context.Background(), "user@example.com", "password123")
	if err != nil {
		t.Fatal(err)
	}
	if err := svc.Logout(context.Background(), pair.AccessToken); err != nil {
		t.Fatal(err)
	}
}

func TestUpdateNickname(t *testing.T) {
	called := false
	repo := &mocks.UserRepository{
		UpdateNicknameFn: func(ctx context.Context, userID, nickname string) error {
			called = true
			if userID != "uid-1" || nickname != "neo" {
				t.Fatalf("bad args %s %s", userID, nickname)
			}
			return nil
		},
	}
	if err := newTestService(repo).UpdateNickname(context.Background(), "uid-1", "neo"); err != nil {
		t.Fatal(err)
	}
	if !called {
		t.Fatal("UpdateNickname not called")
	}
}

func TestGetUserScores(t *testing.T) {
	repo := &mocks.UserRepository{
		GetUserScoresFn: func(ctx context.Context, userID string) (map[string]int32, int32, error) {
			return map[string]int32{"hanoi": 42}, 42, nil
		},
	}
	byGame, total, err := newTestService(repo).GetUserScores(context.Background(), "uid-1")
	if err != nil || total != 42 || byGame["hanoi"] != 42 {
		t.Fatalf("byGame=%v total=%d err=%v", byGame, total, err)
	}
}

func TestRegisterAuthorRole(t *testing.T) {
	repo := &mocks.UserRepository{
		GetByEmailFn: func(ctx context.Context, email string) (*model.User, error) { return nil, nil },
		CreateFn: func(ctx context.Context, email, passwordHash, role string) (string, error) {
			t.Fatal("Create must not be called when role=author is rejected")
			return "", nil
		},
	}
	_, _, err := newTestService(repo).Register(context.Background(), "author@example.com", "password123", "author")
	if !errors.Is(err, model.ErrInvalidRole) {
		t.Fatalf("want ErrInvalidRole, got %v", err)
	}
}

func TestRegisterRepoErrors(t *testing.T) {
	want := errors.New("db down")
	repo := &mocks.UserRepository{
		GetByEmailFn: func(ctx context.Context, email string) (*model.User, error) {
			return nil, want
		},
	}
	_, _, err := newTestService(repo).Register(context.Background(), "user@example.com", "password123", "user")
	if !errors.Is(err, want) {
		t.Fatalf("got %v", err)
	}
}

func TestLoginRepoError(t *testing.T) {
	want := errors.New("db down")
	repo := &mocks.UserRepository{
		GetByEmailFn: func(ctx context.Context, email string) (*model.User, error) {
			return nil, want
		},
	}
	_, err := newTestService(repo).Login(context.Background(), "user@example.com", "password123")
	if !errors.Is(err, want) {
		t.Fatalf("got %v", err)
	}
}

func TestGetUserRepoError(t *testing.T) {
	want := errors.New("db down")
	repo := &mocks.UserRepository{
		GetByIDFn: func(ctx context.Context, id string) (*model.User, error) {
			return nil, want
		},
	}
	_, err := newTestService(repo).GetUser(context.Background(), "uid-1")
	if !errors.Is(err, want) {
		t.Fatalf("got %v", err)
	}
}

func TestRegisterCreateError(t *testing.T) {
	want := errors.New("insert failed")
	repo := &mocks.UserRepository{
		GetByEmailFn: func(ctx context.Context, email string) (*model.User, error) { return nil, nil },
		CreateFn: func(ctx context.Context, email, passwordHash, role string) (string, error) {
			return "", want
		},
	}
	_, _, err := newTestService(repo).Register(context.Background(), "user@example.com", "password123", "user")
	if !errors.Is(err, want) {
		t.Fatalf("got %v", err)
	}
}

func TestUpdateRoleRepoError(t *testing.T) {
	want := errors.New("update failed")
	repo := &mocks.UserRepository{
		UpdateRoleFn: func(ctx context.Context, userID, role string) error { return want },
	}
	err := newTestService(repo).UpdateRole(context.Background(), "uid-1", "admin")
	if !errors.Is(err, want) {
		t.Fatalf("got %v", err)
	}
}

func TestUpdateNicknameRepoError(t *testing.T) {
	want := errors.New("nickname failed")
	repo := &mocks.UserRepository{
		UpdateNicknameFn: func(ctx context.Context, userID, nickname string) error { return want },
	}
	err := newTestService(repo).UpdateNickname(context.Background(), "uid-1", "neo")
	if !errors.Is(err, want) {
		t.Fatalf("got %v", err)
	}
}

func TestWhoamiInvalidToken(t *testing.T) {
	_, err := newTestServiceWithManager(&mocks.UserRepository{}).Whoami(context.Background(), "bad-token")
	if !errors.Is(err, model.ErrInvalidToken) {
		t.Fatalf("got %v", err)
	}
}

func TestListUsers_Delegates(t *testing.T) {
	repo := &mocks.UserRepository{
		ListUsersFn: func(_ context.Context, query string, limit, offset int) ([]*model.User, int64, error) {
			if query != "den" || limit != 10 || offset != 5 {
				t.Fatalf("q=%s limit=%d offset=%d", query, limit, offset)
			}
			return []*model.User{{ID: "u1", Email: "den@x.com"}}, 1, nil
		},
	}
	list, total, err := newTestService(repo).ListUsers(context.Background(), "den", 10, 5)
	if err != nil || total != 1 || len(list) != 1 || list[0].ID != "u1" {
		t.Fatalf("list=%v total=%d err=%v", list, total, err)
	}
}

func TestRegisterRejectsAdminRole(t *testing.T) {
	repo := &mocks.UserRepository{
		GetByEmailFn: func(context.Context, string) (*model.User, error) { return nil, nil },
		CreateFn: func(context.Context, string, string, string) (string, error) {
			t.Fatal("Create must not run for admin self-register")
			return "", nil
		},
	}
	_, _, err := newTestService(repo).Register(context.Background(), "a@b.c", "password123", "admin")
	if !errors.Is(err, model.ErrInvalidRole) {
		t.Fatalf("got %v", err)
	}
}

func withMiniRedis(t *testing.T) (*repository.RedisAuthRepo, func()) {
	t.Helper()
	mr, err := miniredis.Run()
	if err != nil {
		t.Fatal(err)
	}
	cache := repository.NewRedisAuthRepo(mr.Addr(), time.Minute)
	return cache, func() {
		_ = cache.Close()
		mr.Close()
	}
}

func TestLogout_WithCacheDeletesSession(t *testing.T) {
	cache, cleanup := withMiniRedis(t)
	defer cleanup()
	hash, err := bcrypt.GenerateFromPassword([]byte("password123"), bcryptCost)
	if err != nil {
		t.Fatal(err)
	}
	repo := &mocks.UserRepository{
		GetByEmailFn: func(_ context.Context, email string) (*model.User, error) {
			return &model.User{ID: "uid-1", Email: email, PasswordHash: string(hash), Role: "user"}, nil
		},
	}
	mgr := jwtauth.NewManager(testJWTSecret, time.Hour, 24*time.Hour)
	svc := NewAuthService(repo, cache, mgr)
	pair, err := svc.Login(context.Background(), "user@example.com", "password123")
	if err != nil {
		t.Fatal(err)
	}
	if _, _, _, err := svc.ValidateToken(context.Background(), pair.AccessToken); err != nil {
		t.Fatalf("pre-logout validate: %v", err)
	}
	if err := svc.Logout(context.Background(), pair.AccessToken); err != nil {
		t.Fatal(err)
	}
	_, _, _, err = svc.ValidateToken(context.Background(), pair.AccessToken)
	if !errors.Is(err, model.ErrSessionRevoked) {
		t.Fatalf("want ErrSessionRevoked after logout, got %v", err)
	}
}

func TestRefreshToken_RevokesOldRefresh(t *testing.T) {
	cache, cleanup := withMiniRedis(t)
	defer cleanup()
	hash, err := bcrypt.GenerateFromPassword([]byte("password123"), bcryptCost)
	if err != nil {
		t.Fatal(err)
	}
	repo := &mocks.UserRepository{
		GetByEmailFn: func(_ context.Context, email string) (*model.User, error) {
			return &model.User{ID: "uid-1", Email: email, PasswordHash: string(hash), Role: "user"}, nil
		},
	}
	svc := NewAuthService(repo, cache, jwtauth.NewManager(testJWTSecret, time.Hour, 24*time.Hour))
	pair, err := svc.Login(context.Background(), "user@example.com", "password123")
	if err != nil {
		t.Fatal(err)
	}
	next, err := svc.RefreshToken(context.Background(), pair.RefreshToken)
	if err != nil || next.AccessToken == "" {
		t.Fatalf("refresh=%+v err=%v", next, err)
	}
	_, err = svc.RefreshToken(context.Background(), pair.RefreshToken)
	if !errors.Is(err, model.ErrSessionRevoked) {
		t.Fatalf("want revoked old refresh, got %v", err)
	}
}

func TestGetUser_CacheHit(t *testing.T) {
	cache, cleanup := withMiniRedis(t)
	defer cleanup()
	u := &model.User{ID: "uid-1", Email: "a@b.c", Role: "author"}
	if err := cache.SetUserCache(context.Background(), u); err != nil {
		t.Fatal(err)
	}
	repo := &mocks.UserRepository{
		GetByIDFn: func(context.Context, string) (*model.User, error) {
			t.Fatal("repo must not be called on cache hit")
			return nil, nil
		},
	}
	got, err := NewAuthService(repo, cache, jwtauth.NewManager(testJWTSecret, time.Hour, 24*time.Hour)).
		GetUser(context.Background(), "uid-1")
	if err != nil || got.Role != "author" {
		t.Fatalf("got=%+v err=%v", got, err)
	}
}

func TestUpdateRole_InvalidatesCache(t *testing.T) {
	cache, cleanup := withMiniRedis(t)
	defer cleanup()
	if err := cache.SetUserCache(context.Background(), &model.User{ID: "uid-1", Email: "a@b.c", Role: "user"}); err != nil {
		t.Fatal(err)
	}
	repo := &mocks.UserRepository{
		UpdateRoleFn: func(context.Context, string, string) error { return nil },
		GetByIDFn: func(_ context.Context, id string) (*model.User, error) {
			return &model.User{ID: id, Email: "a@b.c", Role: "author"}, nil
		},
	}
	svc := NewAuthService(repo, cache, jwtauth.NewManager(testJWTSecret, time.Hour, 24*time.Hour))
	if err := svc.UpdateRole(context.Background(), "uid-1", "author"); err != nil {
		t.Fatal(err)
	}
	got, err := svc.GetUser(context.Background(), "uid-1")
	if err != nil || got.Role != "author" {
		t.Fatalf("after invalidate should load author from repo: %+v err=%v", got, err)
	}
}
