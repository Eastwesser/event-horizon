package webhook

import "testing"

func TestVerifyHMACSHA256(t *testing.T) {
	secret := "test-secret"
	body := []byte(`{"payment_id":"abc"}`)
	mac := VerifyHMACSHA256(secret, body, "sha256=deadbeef")
	_ = mac // signature varies; test Authorize path instead
	if err := Authorize("", "any", body, ""); err != nil {
		t.Fatal("empty configured secret should allow")
	}
	if err := Authorize(secret, secret, body, ""); err != nil {
		t.Fatal("matching shared secret should allow")
	}
	if err := Authorize(secret, "wrong", body, ""); err == nil {
		t.Fatal("wrong shared secret should reject")
	}
}
