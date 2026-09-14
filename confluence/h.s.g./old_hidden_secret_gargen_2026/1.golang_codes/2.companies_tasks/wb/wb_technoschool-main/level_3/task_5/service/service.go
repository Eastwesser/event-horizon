package service

import (
	"context"
	"fmt"
	"time"

	"github.com/wb_technoschool/level_3/task_5/models"
	"github.com/wb_technoschool/level_3/task_5/storage"
)

type BookingService struct {
	storage           storage.Storage
	defaultTimeoutMin int
}

func NewBookingService(storage storage.Storage, defaultTimeoutMin int) *BookingService {
	return &BookingService{
		storage:           storage,
		defaultTimeoutMin: defaultTimeoutMin,
	}
}

func (s *BookingService) CreateEvent(ctx context.Context, event *models.Event) error {
	if event.Name == "" {
		return fmt.Errorf("event name is required")
	}
	if event.TotalSeats <= 0 {
		return fmt.Errorf("total seats must be greater than 0")
	}
	if event.BookingTimeoutMin <= 0 {
		event.BookingTimeoutMin = s.defaultTimeoutMin
	}
	if event.EventDate.Before(time.Now()) {
		return fmt.Errorf("event date must be in the future")
	}

	return s.storage.CreateEvent(ctx, event)
}

func (s *BookingService) GetEvent(ctx context.Context, id string) (*models.Event, error) {
	return s.storage.GetEvent(ctx, id)
}

func (s *BookingService) GetAllEvents(ctx context.Context) ([]*models.Event, error) {
	return s.storage.GetAllEvents(ctx)
}

func (s *BookingService) CreateBooking(ctx context.Context, eventID, userEmail string) (*models.Booking, error) {
	if userEmail == "" {
		return nil, fmt.Errorf("user email is required")
	}

	event, err := s.storage.GetEvent(ctx, eventID)
	if err != nil {
		return nil, fmt.Errorf("event not found: %w", err)
	}

	if event.AvailableSeats <= 0 {
		return nil, fmt.Errorf("no available seats")
	}

	booking := &models.Booking{
		EventID:   eventID,
		UserEmail: userEmail,
		ExpiresAt: time.Now().Add(time.Duration(event.BookingTimeoutMin) * time.Minute),
	}

	if err := s.storage.CreateBooking(ctx, booking); err != nil {
		return nil, err
	}

	return booking, nil
}

func (s *BookingService) ConfirmBooking(ctx context.Context, id string) error {
	return s.storage.ConfirmBooking(ctx, id)
}

func (s *BookingService) CancelBooking(ctx context.Context, id string) error {
	return s.storage.CancelBooking(ctx, id)
}

func (s *BookingService) GetEventBookings(ctx context.Context, eventID string) ([]*models.Booking, error) {
	return s.storage.GetEventBookings(ctx, eventID)
}

func (s *BookingService) GetUserBookings(ctx context.Context, email string) ([]*models.BookingWithEvent, error) {
	return s.storage.GetUserBookings(ctx, email)
}


