package model

import (
	"errors"
	"time"
)

var (
	ErrNotFound           = errors.New("not found")
	ErrInvalidInput       = errors.New("invalid input")
	ErrAlreadyPrivileged  = errors.New("user is already author or admin")
	ErrApplicationMissing = errors.New("application not found")
	ErrAlreadyReviewed    = errors.New("application already reviewed")
)

type Author struct {
	ID          string
	UserID      string
	DisplayName string
	Bio         string
	AvatarURL   string
	Portfolio   string
	Active      bool
	CreatedAt   time.Time
	UpdatedAt   time.Time
	VerifiedAt  *time.Time
}

type ApplicationStatus string

const (
	ApplicationPending  ApplicationStatus = "pending"
	ApplicationApproved ApplicationStatus = "approved"
	ApplicationRejected ApplicationStatus = "rejected"
)

type ApplicationPayload struct {
	DisplayName  string `json:"display_name"`
	Portfolio    string `json:"portfolio"`
	Motivation   string `json:"motivation"`
	ContactEmail string `json:"contact_email"`
}

type AuthorApplication struct {
	ID           string
	UserID       string
	Status       ApplicationStatus
	Payload      ApplicationPayload
	CreatedAt    time.Time
	ReviewedAt   *time.Time
	ReviewedBy   string
	ReviewerNote string
}
