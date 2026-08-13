package model

import "errors"

var (
	ErrProfileNotFound = errors.New("profile not found")
	ErrInvalidNickname = errors.New("invalid nickname")
)
