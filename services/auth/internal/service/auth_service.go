// Package service — auth business logic (Week 6: access + refresh JWT).
package service

import (
	"context"
	"errors"
	"fmt"
	"log"
	"time"

	"golang.org/x/crypto/bcrypt"

	jwtauth "github.com/Eastwesser/event-horizon/services/auth/internal/jwt"
	"github.com/Eastwesser/event-horizon/services/auth/internal/model"
	"github.com/Eastwesser/event-horizon/services/auth/internal/repository"
)

const bcryptCost = 12

var validPublicRoles = map[string]bool{"user": true, "author": true}
var validRoles = map[string]bool{"user": true, "author": true, "admin": true}

// TokenPair is returned by Login / RefreshToken.
type TokenPair struct {
	AccessToken  string
	RefreshToken string
	ExpiresIn    int64
	UserID       string
	Role         string
}

type authService struct {
	repo  repository.UserRepository
	cache *repository.RedisAuthRepo
	tokens *jwtauth.Manager
}

type AuthService interface {
	Register(ctx context.Context, email, password, role string) (userID, resolvedRole string, err error)
	Login(ctx context.Context, email, password string) (*TokenPair, error)
	ValidateToken(ctx context.Context, tokenString string) (userID, email, role string, err error)
	RefreshToken(ctx context.Context, refreshToken string) (*TokenPair, error)
	Whoami(ctx context.Context, accessToken string) (*model.User, error)
	Logout(ctx context.Context, tokenString string) error
	GetUser(ctx context.Context, userID string) (*model.User, error)
	UpdateNickname(ctx context.Context, userID, nickname string) error
	UpdateRole(ctx context.Context, userID, role string) error
	GetUserScores(ctx context.Context, userID string) (map[string]int32, int32, error)
}

func NewAuthService(repo repository.UserRepository, cache *repository.RedisAuthRepo, tokens *jwtauth.Manager) AuthService {
	return &authService{repo: repo, cache: cache, tokens: tokens}
}

// NewAuthServiceLegacy keeps older call sites working (tests / gradual migrate).
func NewAuthServiceLegacy(repo repository.UserRepository, cache *repository.RedisAuthRepo, jwtSecret string, expHours int) AuthService {
	access := 15 * time.Minute
	refresh := 7 * 24 * time.Hour
	if expHours > 0 && expHours < 24 {
		// treat small values as access hours for backward-compat tests
		access = time.Duration(expHours) * time.Hour
	}
	return NewAuthService(repo, cache, jwtauth.NewManager(jwtSecret, access, refresh))
}

func (s *authService) Register(ctx context.Context, email, password, role string) (string, string, error) {
	existing, err := s.repo.GetByEmail(ctx, email)
	if err != nil {
		return "", "", err
	}
	if existing != nil {
		return "", "", model.ErrUserAlreadyExists
	}
	if role == "" {
		role = "user"
	}
	if !validPublicRoles[role] {
		return "", "", fmt.Errorf("%w: %q cannot be self-assigned at registration", model.ErrInvalidRole, role)
	}
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcryptCost)
	if err != nil {
		return "", "", err
	}
	userID, err := s.repo.Create(ctx, email, string(hashedPassword), role)
	if err != nil {
		return "", "", err
	}
	return userID, role, nil
}

func (s *authService) Login(ctx context.Context, email, password string) (*TokenPair, error) {
	user, err := s.repo.GetByEmail(ctx, email)
	if err != nil {
		return nil, err
	}
	if user == nil {
		return nil, model.ErrInvalidCredentials
	}
	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(password)); err != nil {
		return nil, model.ErrInvalidCredentials
	}
	role := user.Role
	if role == "" {
		role = "user"
	}
	return s.issuePair(ctx, user.ID, user.Email, role)
}

func (s *authService) issuePair(ctx context.Context, userID, email, role string) (*TokenPair, error) {
	access, accessJTI, accessExp, err := s.tokens.GenerateAccessToken(userID, email, role)
	if err != nil {
		return nil, err
	}
	refresh, refreshJTI, refreshExp, err := s.tokens.GenerateRefreshToken(userID, email, role)
	if err != nil {
		return nil, err
	}
	if s.cache != nil {
		if err := s.cache.CreateSession(ctx, accessJTI, userID, time.Until(accessExp)); err != nil {
			log.Printf("auth: store access session failed: %v", err)
		}
		if err := s.cache.SaveRefresh(ctx, refreshJTI, userID, time.Until(refreshExp)); err != nil {
			log.Printf("auth: store refresh failed: %v", err)
		}
	}
	return &TokenPair{
		AccessToken:  access,
		RefreshToken: refresh,
		ExpiresIn:    int64(s.tokens.AccessTTL().Seconds()),
		UserID:       userID,
		Role:         role,
	}, nil
}

func (s *authService) ValidateToken(ctx context.Context, tokenString string) (string, string, string, error) {
	claims, err := s.tokens.ValidateAccessToken(tokenString)
	if err != nil {
		// legacy tokens without type=access may still validate as MapClaims via fallback
		return "", "", "", model.ErrInvalidToken
	}
	if s.cache != nil && claims.JTI != "" {
		exists, cacheErr := s.cache.SessionExists(ctx, claims.JTI)
		if cacheErr != nil {
			log.Printf("auth: Redis session check failed: %v", cacheErr)
		} else if !exists {
			return "", "", "", model.ErrSessionRevoked
		}
	}
	return claims.UserID, claims.Email, claims.Role, nil
}

func (s *authService) RefreshToken(ctx context.Context, refreshToken string) (*TokenPair, error) {
	claims, err := s.tokens.ValidateRefreshToken(refreshToken)
	if err != nil {
		return nil, model.ErrInvalidToken
	}
	if s.cache != nil && claims.JTI != "" {
		ok, cacheErr := s.cache.RefreshExists(ctx, claims.JTI)
		if cacheErr != nil {
			log.Printf("auth: Redis refresh check failed: %v", cacheErr)
		} else if !ok {
			return nil, model.ErrSessionRevoked
		}
		_ = s.cache.DeleteRefresh(ctx, claims.JTI, claims.UserID)
	}
	return s.issuePair(ctx, claims.UserID, claims.Email, claims.Role)
}

func (s *authService) Whoami(ctx context.Context, accessToken string) (*model.User, error) {
	userID, _, _, err := s.ValidateToken(ctx, accessToken)
	if err != nil {
		return nil, err
	}
	return s.GetUser(ctx, userID)
}

func (s *authService) Logout(ctx context.Context, tokenString string) error {
	if s.cache == nil {
		return nil
	}
	claims, err := jwtauth.ParseUnverified(tokenString)
	if err != nil {
		return err
	}
	if claims.JTI == "" {
		return nil
	}
	switch claims.Type {
	case jwtauth.TokenTypeRefresh:
		return s.cache.DeleteRefresh(ctx, claims.JTI, claims.UserID)
	default:
		// access or legacy
		_ = s.cache.DeleteSession(ctx, claims.JTI)
		if claims.Type == jwtauth.TokenTypeRefresh {
			_ = s.cache.DeleteRefresh(ctx, claims.JTI, claims.UserID)
		}
		return nil
	}
}

func (s *authService) GetUser(ctx context.Context, userID string) (*model.User, error) {
	if s.cache != nil {
		cached, err := s.cache.GetUserCache(ctx, userID)
		if err == nil {
			return cached, nil
		}
		if !errors.Is(err, repository.ErrCacheMiss) {
			log.Printf("auth: Redis GetUserCache failed: %v", err)
		}
	}
	user, err := s.repo.GetByID(ctx, userID)
	if err != nil {
		return nil, err
	}
	if user != nil && s.cache != nil {
		if cacheErr := s.cache.SetUserCache(ctx, user); cacheErr != nil {
			log.Printf("auth: Redis SetUserCache failed: %v", cacheErr)
		}
	}
	return user, nil
}

func (s *authService) UpdateNickname(ctx context.Context, userID, nickname string) error {
	if err := s.repo.UpdateNickname(ctx, userID, nickname); err != nil {
		return err
	}
	s.invalidateUserCache(ctx, userID)
	return nil
}

func (s *authService) UpdateRole(ctx context.Context, userID, role string) error {
	if !validRoles[role] {
		return fmt.Errorf("%w: %q", model.ErrInvalidRole, role)
	}
	if err := s.repo.UpdateRole(ctx, userID, role); err != nil {
		return err
	}
	s.invalidateUserCache(ctx, userID)
	return nil
}

func (s *authService) invalidateUserCache(ctx context.Context, userID string) {
	if s.cache == nil {
		return
	}
	if err := s.cache.InvalidateUserCache(ctx, userID); err != nil {
		log.Printf("auth: Redis InvalidateUserCache failed: %v", err)
	}
}

func (s *authService) GetUserScores(ctx context.Context, userID string) (map[string]int32, int32, error) {
	return s.repo.GetUserScores(ctx, userID)
}
