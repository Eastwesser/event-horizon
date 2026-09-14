package models

import (
	"time"
)

type Event struct {
	ID                string    `json:"id" db:"id"`
	Name              string    `json:"name" db:"name"`
	Description       string    `json:"description" db:"description"`
	EventDate         time.Time `json:"event_date" db:"event_date"`
	TotalSeats        int       `json:"total_seats" db:"total_seats"`
	AvailableSeats    int       `json:"available_seats" db:"available_seats"`
	BookingTimeoutMin int       `json:"booking_timeout_min" db:"booking_timeout_min"`
	CreatedAt         time.Time `json:"created_at" db:"created_at"`
}

type Booking struct {
	ID          string        `json:"id" db:"id"`
	EventID     string        `json:"event_id" db:"event_id"`
	UserEmail   string        `json:"user_email" db:"user_email"`
	Status      BookingStatus `json:"status" db:"status"`
	CreatedAt   time.Time     `json:"created_at" db:"created_at"`
	ConfirmedAt *time.Time    `json:"confirmed_at,omitempty" db:"confirmed_at"`
	ExpiresAt   time.Time     `json:"expires_at" db:"expires_at"`
}

type BookingStatus string

const (
	StatusPending   BookingStatus = "pending"
	StatusConfirmed BookingStatus = "confirmed"
	StatusCancelled BookingStatus = "cancelled"
)

type BookingWithEvent struct {
	Booking
	EventName string    `json:"event_name" db:"event_name"`
	EventDate time.Time `json:"event_date" db:"event_date"`
}
