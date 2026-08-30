package webhook

import (
	"crypto/hmac"
	"crypto/sha256"
	"crypto/subtle"
	"encoding/hex"
	"strings"
)

// Authorize verifies Boosty-style webhook auth: HMAC header (preferred) or shared secret field.
func Authorize(configuredSecret, providedSecret string, rawBody []byte, signatureHeader string) error {
	if configuredSecret == "" {
		return nil // dev: no secret configured
	}
	if signatureHeader != "" {
		if VerifyHMACSHA256(configuredSecret, rawBody, signatureHeader) {
			return nil
		}
		return ErrInvalidSignature
	}
	if providedSecret != "" && subtle.ConstantTimeCompare([]byte(providedSecret), []byte(configuredSecret)) == 1 {
		return nil
	}
	return ErrUnauthorized
}

var (
	ErrUnauthorized       = errString("webhook unauthorized")
	ErrInvalidSignature   = errString("webhook signature invalid")
)

type errString string

func (e errString) Error() string { return string(e) }

func VerifyHMACSHA256(secret string, rawBody []byte, signatureHeader string) bool {
	if secret == "" || signatureHeader == "" {
		return false
	}
	sig := strings.TrimSpace(signatureHeader)
	if i := strings.Index(sig, "="); i >= 0 {
		sig = sig[i+1:]
	}
	mac := hmac.New(sha256.New, []byte(secret))
	mac.Write(rawBody)
	sum := mac.Sum(nil)
	got, err := hex.DecodeString(sig)
	if err != nil {
		return false
	}
	return subtle.ConstantTimeCompare(sum, got) == 1
}
