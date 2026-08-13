package jwtauth

import (
	"testing"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

func TestAccessAndRefreshRoundTrip(t *testing.T) {
	m := NewManager("secret", 0, 0)
	access, _, _, err := m.GenerateAccessToken("u1", "a@b.c", "user")
	if err != nil {
		t.Fatal(err)
	}
	refresh, _, _, err := m.GenerateRefreshToken("u1", "a@b.c", "user")
	if err != nil {
		t.Fatal(err)
	}
	ac, err := m.ValidateAccessToken(access)
	if err != nil || ac.UserID != "u1" || ac.Type != TokenTypeAccess {
		t.Fatalf("access: %+v %v", ac, err)
	}
	rc, err := m.ValidateRefreshToken(refresh)
	if err != nil || rc.Type != TokenTypeRefresh {
		t.Fatalf("refresh: %+v %v", rc, err)
	}
	if _, err := m.ValidateAccessToken(refresh); err == nil {
		t.Fatal("refresh must not validate as access")
	}
}

func TestDefaultTTLs(t *testing.T) {
	m := NewManager("x", 0, 0)
	if m.AccessTTL() != 15*time.Minute {
		t.Fatalf("access ttl %v", m.AccessTTL())
	}
	if m.RefreshTTL() != 7*24*time.Hour {
		t.Fatalf("refresh ttl %v", m.RefreshTTL())
	}
}

func TestExpiredAccessRejected(t *testing.T) {
	m := NewManager("secret", time.Millisecond, time.Hour)
	tok, _, _, err := m.GenerateAccessToken("u1", "a@b.c", "user")
	if err != nil {
		t.Fatal(err)
	}
	time.Sleep(5 * time.Millisecond)
	if _, err := m.ValidateAccessToken(tok); err == nil {
		t.Fatal("expected expiry error")
	}
}

func TestTamperedTokenRejected(t *testing.T) {
	m := NewManager("secret", time.Hour, time.Hour)
	tok, _, _, err := m.GenerateAccessToken("u1", "a@b.c", "admin")
	if err != nil {
		t.Fatal(err)
	}
	bad := tok + "x"
	if _, err := m.ValidateAccessToken(bad); err == nil {
		t.Fatal("expected signature error")
	}
}

func TestLegacyAccessWithoutTypeClaim(t *testing.T) {
	m := NewManager("secret", time.Hour, time.Hour)
	claims := Claims{
		UserID: "u1",
		Email:  "a@b.c",
		Role:   "",
		Type:   "",
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(time.Hour)),
		},
	}
	raw, err := jwt.NewWithClaims(jwt.SigningMethodHS256, claims).SignedString([]byte("secret"))
	if err != nil {
		t.Fatal(err)
	}
	got, err := m.ValidateAccessToken(raw)
	if err != nil {
		t.Fatal(err)
	}
	if got.Role != "user" {
		t.Fatalf("default role want user, got %q", got.Role)
	}
}

func TestParseUnverified(t *testing.T) {
	m := NewManager("secret", time.Hour, time.Hour)
	tok, _, _, err := m.GenerateAccessToken("u9", "x@y.z", "author")
	if err != nil {
		t.Fatal(err)
	}
	c, err := ParseUnverified(tok)
	if err != nil || c.UserID != "u9" || c.Role != "author" {
		t.Fatalf("%+v %v", c, err)
	}
}

func TestMissingUserIDRejected(t *testing.T) {
	m := NewManager("secret", time.Hour, time.Hour)
	claims := Claims{
		UserID: "",
		Email:  "a@b.c",
		Type:   TokenTypeAccess,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(time.Hour)),
		},
	}
	raw, err := jwt.NewWithClaims(jwt.SigningMethodHS256, claims).SignedString([]byte("secret"))
	if err != nil {
		t.Fatal(err)
	}
	if _, err := m.ValidateAccessToken(raw); err == nil {
		t.Fatal("expected missing user_id")
	}
}

func TestWrongSecretRejected(t *testing.T) {
	m := NewManager("secret", time.Hour, time.Hour)
	tok, _, _, err := m.GenerateAccessToken("u1", "a@b.c", "user")
	if err != nil {
		t.Fatal(err)
	}
	other := NewManager("other", time.Hour, time.Hour)
	if _, err := other.ValidateAccessToken(tok); err == nil {
		t.Fatal("expected signature error")
	}
}

func TestParseUnverifiedBadToken(t *testing.T) {
	if _, err := ParseUnverified("not-a-jwt"); err == nil {
		t.Fatal("expected error")
	}
}
