package model

import (
	"errors"
	"time"
)

var (
	ErrNotFound     = errors.New("not found")
	ErrInvalidInput = errors.New("invalid input")
)

type Author struct {
	ID          string
	UserID      string
	DisplayName string
	Bio         string
	AvatarURL   string
	Active      bool
	CreatedAt   time.Time
	UpdatedAt   time.Time
}
