package model

import (
	"errors"
	"time"
)

var (
	ErrInvalidInput = errors.New("invalid input")
	ErrNotFound     = errors.New("notification not found")
)

type Notification struct {
	ID        string
	UserID    string
	Title     string
	Body      string
	Link      string
	Source    string
	CreatedAt time.Time
	ReadAt    *time.Time
}
