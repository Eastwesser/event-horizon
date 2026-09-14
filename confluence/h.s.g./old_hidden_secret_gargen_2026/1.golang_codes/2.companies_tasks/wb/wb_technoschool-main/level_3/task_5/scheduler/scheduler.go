package scheduler

import (
	"context"
	"time"

	"github.com/wb-go/wbf/zlog"
	"github.com/wb_technoschool/level_3/task_5/storage"
)

type BookingScheduler struct {
	storage  storage.Storage
	interval time.Duration
}

func NewBookingScheduler(storage storage.Storage, checkInterval time.Duration) *BookingScheduler {
	return &BookingScheduler{
		storage:  storage,
		interval: checkInterval,
	}
}

func (s *BookingScheduler) Start(ctx context.Context) {
	ticker := time.NewTicker(s.interval)
	defer ticker.Stop()

	zlog.Logger.Info().
		Dur("interval", s.interval).
		Msg("Booking scheduler started")

	for {
		select {
		case <-ctx.Done():
			zlog.Logger.Info().Msg("Booking scheduler stopped")
			return
		case <-ticker.C:
			s.processExpiredBookings(ctx)
		}
	}
}

func (s *BookingScheduler) processExpiredBookings(ctx context.Context) {
	bookings, err := s.storage.GetExpiredBookings(ctx)
	if err != nil {
		zlog.Logger.Error().Err(err).Msg("Failed to get expired bookings")
		return
	}

	if len(bookings) == 0 {
		return
	}

	zlog.Logger.Info().
		Int("count", len(bookings)).
		Msg("Processing expired bookings")

	for _, booking := range bookings {
		if err := s.storage.CancelBooking(ctx, booking.ID); err != nil {
			zlog.Logger.Error().
				Err(err).
				Str("booking_id", booking.ID).
				Msg("Failed to cancel booking")
		} else {
			zlog.Logger.Info().
				Str("booking_id", booking.ID).
				Str("event_id", booking.EventID).
				Str("user_email", booking.UserEmail).
				Msg("Booking cancelled due to timeout")
		}
	}
}


