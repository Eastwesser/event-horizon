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
	ID            string
	UserID        string
	Status        ApplicationStatus
	Payload       ApplicationPayload
	CreatedAt     time.Time
	ReviewedAt    *time.Time
	ReviewedBy    string
	ReviewerNote  string
}
