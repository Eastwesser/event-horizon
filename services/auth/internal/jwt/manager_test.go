package jwtauth

import "testing"

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
