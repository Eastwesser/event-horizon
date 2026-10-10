package model

import "errors"

var (
	ErrUserAlreadyExists       = errors.New("user already exists")
	ErrInvalidCredentials      = errors.New("invalid credentials")
	ErrInvalidToken            = errors.New("invalid token")
	ErrSessionRevoked          = errors.New("session revoked or expired")
	ErrSessionStoreUnavailable = errors.New("session store unavailable")
	ErrUserNotFound            = errors.New("user not found")
	ErrInvalidRole             = errors.New("invalid role")
)
