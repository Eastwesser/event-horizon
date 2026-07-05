// # бизнес-логика
package service

import (
	"context"
	"errors"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/crypto/bcrypt"

	"github.com/Eastwesser/event-horizon/services/auth/internal/repository"
)

type authService struct {
	repo      repository.UserRepository
	jwtSecret string
	expHours  int
}

type AuthService interface {
	Register(ctx context.Context, email, password string) (string, error)
	Login(ctx context.Context, email, password string) (string, error)
	ValidateToken(ctx context.Context, token string) (string, string, error)
	GetUser(ctx context.Context, userID string) (*repository.User, error)
	UpdateNickname(ctx context.Context, userID, nickname string) error
	GetUserScores(ctx context.Context, userID string) (map[string]int32, int32, error)
}

func NewAuthService(repo repository.UserRepository, jwtSecret string, expHours int) AuthService {
	return &authService{
		repo:      repo,
		jwtSecret: jwtSecret,
		expHours:  expHours,
	}
}

func (s *authService) Register(ctx context.Context, email, password string) (string, error) {
    // Проверяем, существует ли пользователь
    existing, err := s.repo.GetByEmail(ctx, email)
    if err != nil {
        return "", err
    }
    if existing != nil {
        return "", errors.New("user already exists")
    }
    
    // Хэшируем пароль
    hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
    if err != nil {
        return "", err
    }
    
    // Создаём пользователя
    userID, err := s.repo.Create(ctx, email, string(hashedPassword))
    if err != nil {
        return "", err
    }
    
    return userID, nil
}

func (s *authService) Login(ctx context.Context, email, password string) (string, error) {
	// Находим пользователя
	user, err := s.repo.GetByEmail(ctx, email)
	if err != nil {
		return "", err
	}
	if user == nil {
		return "", errors.New("invalid credentials")
	}
	
	// Проверяем пароль
	err = bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(password))
	if err != nil {
		return "", errors.New("invalid credentials")
	}
	
	// Генерируем JWT
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"user_id": user.ID,
		"email":   user.Email,
		"exp":     time.Now().Add(time.Duration(s.expHours) * time.Hour).Unix(),
	})
	
	tokenString, err := token.SignedString([]byte(s.jwtSecret))
	if err != nil {
		return "", err
	}
	
	return tokenString, nil
}

func (s *authService) ValidateToken(ctx context.Context, tokenString string) (string, string, error) {
	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, errors.New("unexpected signing method")
		}
		return []byte(s.jwtSecret), nil
	})
	
	if err != nil {
		return "", "", err
	}
	
	if claims, ok := token.Claims.(jwt.MapClaims); ok && token.Valid {
		userID := claims["user_id"].(string)
		email := claims["email"].(string)
		return userID, email, nil
	}
	
	return "", "", errors.New("invalid token")
}

func (s *authService) GetUser(ctx context.Context, userID string) (*repository.User, error) {
	return s.repo.GetByID(ctx, userID)
}

func (s *authService) UpdateNickname(ctx context.Context, userID, nickname string) error {
    return s.repo.UpdateNickname(ctx, userID, nickname)
}

func (s *authService) GetUserScores(ctx context.Context, userID string) (map[string]int32, int32, error) {
    return s.repo.GetUserScores(ctx, userID)
}
