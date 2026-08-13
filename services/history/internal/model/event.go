package model

import (
	"errors"
	"time"
)

var (
	ErrNotFound     = errors.New("not found")
	ErrInvalidInput = errors.New("invalid input")
)

type Event struct {
	ID        string
	UserID    string
	EventType string
	Payload   string
	CreatedAt time.Time
}
