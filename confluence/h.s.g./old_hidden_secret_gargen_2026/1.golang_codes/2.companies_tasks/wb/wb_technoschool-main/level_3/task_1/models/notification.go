package models

import (
	"time"
)

type NotificationStatus string

const (
	StatusPending   NotificationStatus = "pending"
	StatusScheduled NotificationStatus = "scheduled"
	StatusSending   NotificationStatus = "sending"
	StatusSent      NotificationStatus = "sent"
	StatusFailed    NotificationStatus = "failed"
	StatusCancelled NotificationStatus = "cancelled"
)

type NotificationChannel string

const (
	ChannelEmail    NotificationChannel = "email"
	ChannelTelegram NotificationChannel = "telegram"
	ChannelConsole  NotificationChannel = "console"
)

type Notification struct {
	ID          string              `json:"id"`
	Channel     NotificationChannel `json:"channel"`
	Recipient   string              `json:"recipient"`
	Subject     string              `json:"subject,omitempty"`
	Message     string              `json:"message"`
	ScheduledAt time.Time           `json:"scheduled_at"`
	Status      NotificationStatus  `json:"status"`
	CreatedAt   time.Time           `json:"created_at"`
	UpdatedAt   time.Time           `json:"updated_at"`
	SentAt      *time.Time          `json:"sent_at,omitempty"`
	RetryCount  int                 `json:"retry_count"`
	LastError   string              `json:"last_error,omitempty"`
}

type CreateNotificationRequest struct {
	Channel     NotificationChannel `json:"channel" binding:"required"`
	Recipient   string              `json:"recipient" binding:"required"`
	Subject     string              `json:"subject,omitempty"`
	Message     string              `json:"message" binding:"required"`
	ScheduledAt time.Time           `json:"scheduled_at" binding:"required"`
}
