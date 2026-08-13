package jwtauth

import (
	"errors"
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
)

const (
	TokenTypeAccess  = "access"
	TokenTypeRefresh = "refresh"
)

// Claims carries EH identity fields.
type Claims struct {
	UserID string `json:"user_id"`
	Email  string `json:"email"`
	Role   string `json:"role"`
	Type   string `json:"type"`
	JTI    string `json:"jti"`
	jwt.RegisteredClaims
}

type Generator interface {
	GenerateAccessToken(userID, email, role string) (token string, jti string, expiresAt time.Time, err error)
	GenerateRefreshToken(userID, email, role string) (token string, jti string, expiresAt time.Time, err error)
}

type Validator interface {
	ValidateAccessToken(token string) (*Claims, error)
	ValidateRefreshToken(token string) (*Claims, error)
}

type Manager struct {
	secret        []byte
	accessTTL     time.Duration
	refreshTTL    time.Duration
}

func NewManager(secret string, accessTTL, refreshTTL time.Duration) *Manager {
	if accessTTL <= 0 {
		accessTTL = 15 * time.Minute
	}
	if refreshTTL <= 0 {
		refreshTTL = 7 * 24 * time.Hour
	}
	return &Manager{
		secret:     []byte(secret),
		accessTTL:  accessTTL,
		refreshTTL: refreshTTL,
	}
}

func (m *Manager) AccessTTL() time.Duration  { return m.accessTTL }
func (m *Manager) RefreshTTL() time.Duration { return m.refreshTTL }

func (m *Manager) GenerateAccessToken(userID, email, role string) (string, string, time.Time, error) {
	return m.generate(userID, email, role, TokenTypeAccess, m.accessTTL)
}

func (m *Manager) GenerateRefreshToken(userID, email, role string) (string, string, time.Time, error) {
	return m.generate(userID, email, role, TokenTypeRefresh, m.refreshTTL)
}

func (m *Manager) generate(userID, email, role, typ string, ttl time.Duration) (string, string, time.Time, error) {
	jti := uuid.NewString()
	exp := time.Now().Add(ttl)
	claims := Claims{
		UserID: userID,
		Email:  email,
		Role:   role,
		Type:   typ,
		JTI:    jti,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(exp),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			ID:        jti,
		},
	}
	t := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	s, err := t.SignedString(m.secret)
	return s, jti, exp, err
}

func (m *Manager) ValidateAccessToken(token string) (*Claims, error) {
	return m.validate(token, TokenTypeAccess)
}

func (m *Manager) ValidateRefreshToken(token string) (*Claims, error) {
	return m.validate(token, TokenTypeRefresh)
}

func (m *Manager) validate(tokenString, wantType string) (*Claims, error) {
	token, err := jwt.ParseWithClaims(tokenString, &Claims{}, func(t *jwt.Token) (interface{}, error) {
		if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method")
		}
		return m.secret, nil
	})
	if err != nil {
		return nil, err
	}
	claims, ok := token.Claims.(*Claims)
	if !ok || !token.Valid {
		return nil, errors.New("invalid token")
	}
	if claims.Type != wantType {
		// Legacy access tokens (pre-W6) had no type claim.
		if !(wantType == TokenTypeAccess && claims.Type == "") {
			return nil, fmt.Errorf("unexpected token type %q", claims.Type)
		}
	}
	if claims.UserID == "" {
		return nil, errors.New("missing user_id")
	}
	if claims.Role == "" {
		claims.Role = "user"
	}
	return claims, nil
}

// ParseUnverified extracts claims without signature check (logout helper).
func ParseUnverified(tokenString string) (*Claims, error) {
	parser := jwt.NewParser()
	token, _, err := parser.ParseUnverified(tokenString, &Claims{})
	if err != nil {
		return nil, err
	}
	claims, ok := token.Claims.(*Claims)
	if !ok {
		return nil, errors.New("invalid claims")
	}
	return claims, nil
}
