package storage

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq"
	"github.com/wb_technoschool/level_3/task_5/models"
)

type Storage interface {
	CreateEvent(ctx context.Context, event *models.Event) error
	GetEvent(ctx context.Context, id string) (*models.Event, error)
	GetAllEvents(ctx context.Context) ([]*models.Event, error)
	CreateBooking(ctx context.Context, booking *models.Booking) error
	GetBooking(ctx context.Context, id string) (*models.Booking, error)
	ConfirmBooking(ctx context.Context, id string) error
	CancelBooking(ctx context.Context, id string) error
	GetExpiredBookings(ctx context.Context) ([]*models.Booking, error)
	GetEventBookings(ctx context.Context, eventID string) ([]*models.Booking, error)
	GetUserBookings(ctx context.Context, email string) ([]*models.BookingWithEvent, error)
}

type PostgresStorage struct {
	db *sqlx.DB
}

func NewPostgresStorage(databaseURL string) (*PostgresStorage, error) {
	db, err := sqlx.Connect("postgres", databaseURL)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to database: %w", err)
	}

	db.SetMaxOpenConns(25)
	db.SetMaxIdleConns(5)
	db.SetConnMaxLifetime(5 * time.Minute)

	storage := &PostgresStorage{db: db}
	if err := storage.initSchema(); err != nil {
		return nil, fmt.Errorf("failed to initialize schema: %w", err)
	}

	return storage, nil
}

func (s *PostgresStorage) initSchema() error {
	schema := `
	CREATE TABLE IF NOT EXISTS events (
		id VARCHAR(36) PRIMARY KEY,
		name VARCHAR(255) NOT NULL,
		description TEXT,
		event_date TIMESTAMP NOT NULL,
		total_seats INTEGER NOT NULL,
		available_seats INTEGER NOT NULL,
		booking_timeout_min INTEGER NOT NULL DEFAULT 15,
		created_at TIMESTAMP NOT NULL DEFAULT NOW()
	);

	CREATE TABLE IF NOT EXISTS bookings (
		id VARCHAR(36) PRIMARY KEY,
		event_id VARCHAR(36) NOT NULL REFERENCES events(id) ON DELETE CASCADE,
		user_email VARCHAR(255) NOT NULL,
		status VARCHAR(20) NOT NULL,
		created_at TIMESTAMP NOT NULL DEFAULT NOW(),
		confirmed_at TIMESTAMP,
		expires_at TIMESTAMP NOT NULL,
		CONSTRAINT fk_event FOREIGN KEY (event_id) REFERENCES events(id)
	);

	CREATE INDEX IF NOT EXISTS idx_bookings_event_id ON bookings(event_id);
	CREATE INDEX IF NOT EXISTS idx_bookings_status ON bookings(status);
	CREATE INDEX IF NOT EXISTS idx_bookings_expires_at ON bookings(expires_at);
	CREATE INDEX IF NOT EXISTS idx_bookings_user_email ON bookings(user_email);
	`

	_, err := s.db.Exec(schema)
	return err
}

func (s *PostgresStorage) CreateEvent(ctx context.Context, event *models.Event) error {
	event.ID = uuid.New().String()
	event.CreatedAt = time.Now()
	event.AvailableSeats = event.TotalSeats

	query := `
		INSERT INTO events (id, name, description, event_date, total_seats, available_seats, booking_timeout_min, created_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
	`

	_, err := s.db.ExecContext(ctx, query,
		event.ID, event.Name, event.Description, event.EventDate,
		event.TotalSeats, event.AvailableSeats, event.BookingTimeoutMin, event.CreatedAt,
	)
	return err
}

func (s *PostgresStorage) GetEvent(ctx context.Context, id string) (*models.Event, error) {
	var event models.Event
	query := `SELECT * FROM events WHERE id = $1`
	err := s.db.GetContext(ctx, &event, query, id)
	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("event not found")
	}
	return &event, err
}

func (s *PostgresStorage) GetAllEvents(ctx context.Context) ([]*models.Event, error) {
	var events []*models.Event
	query := `SELECT * FROM events ORDER BY event_date ASC`
	err := s.db.SelectContext(ctx, &events, query)
	return events, err
}

func (s *PostgresStorage) CreateBooking(ctx context.Context, booking *models.Booking) error {
	tx, err := s.db.BeginTxx(ctx, nil)
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback()

	var availableSeats int
	query := `SELECT available_seats FROM events WHERE id = $1 FOR UPDATE`
	if err := tx.GetContext(ctx, &availableSeats, query, booking.EventID); err != nil {
		return fmt.Errorf("failed to get event: %w", err)
	}

	if availableSeats <= 0 {
		return fmt.Errorf("no available seats")
	}

	booking.ID = uuid.New().String()
	booking.CreatedAt = time.Now()
	booking.Status = models.StatusPending

	insertQuery := `
		INSERT INTO bookings (id, event_id, user_email, status, created_at, expires_at)
		VALUES ($1, $2, $3, $4, $5, $6)
	`
	if _, err := tx.ExecContext(ctx, insertQuery,
		booking.ID, booking.EventID, booking.UserEmail, booking.Status,
		booking.CreatedAt, booking.ExpiresAt,
	); err != nil {
		return fmt.Errorf("failed to create booking: %w", err)
	}

	updateQuery := `UPDATE events SET available_seats = available_seats - 1 WHERE id = $1`
	if _, err := tx.ExecContext(ctx, updateQuery, booking.EventID); err != nil {
		return fmt.Errorf("failed to update seats: %w", err)
	}

	return tx.Commit()
}

func (s *PostgresStorage) GetBooking(ctx context.Context, id string) (*models.Booking, error) {
	var booking models.Booking
	query := `SELECT * FROM bookings WHERE id = $1`
	err := s.db.GetContext(ctx, &booking, query, id)
	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("booking not found")
	}
	return &booking, err
}

func (s *PostgresStorage) ConfirmBooking(ctx context.Context, id string) error {
	now := time.Now()
	query := `
		UPDATE bookings 
		SET status = $1, confirmed_at = $2 
		WHERE id = $3 AND status = $4
	`
	result, err := s.db.ExecContext(ctx, query, models.StatusConfirmed, now, id, models.StatusPending)
	if err != nil {
		return err
	}

	rows, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if rows == 0 {
		return fmt.Errorf("booking not found or already processed")
	}

	return nil
}

func (s *PostgresStorage) CancelBooking(ctx context.Context, id string) error {
	tx, err := s.db.BeginTxx(ctx, nil)
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback()

	var booking models.Booking
	query := `SELECT * FROM bookings WHERE id = $1 FOR UPDATE`
	if err := tx.GetContext(ctx, &booking, query, id); err != nil {
		return fmt.Errorf("booking not found: %w", err)
	}

	if booking.Status == models.StatusCancelled {
		return tx.Commit()
	}

	updateBooking := `UPDATE bookings SET status = $1 WHERE id = $2`
	if _, err := tx.ExecContext(ctx, updateBooking, models.StatusCancelled, id); err != nil {
		return fmt.Errorf("failed to cancel booking: %w", err)
	}

	if booking.Status == models.StatusPending {
		updateSeats := `UPDATE events SET available_seats = available_seats + 1 WHERE id = $1`
		if _, err := tx.ExecContext(ctx, updateSeats, booking.EventID); err != nil {
			return fmt.Errorf("failed to update seats: %w", err)
		}
	}

	return tx.Commit()
}

func (s *PostgresStorage) GetExpiredBookings(ctx context.Context) ([]*models.Booking, error) {
	var bookings []*models.Booking
	query := `
		SELECT * FROM bookings 
		WHERE status = $1 AND expires_at < $2
	`
	err := s.db.SelectContext(ctx, &bookings, query, models.StatusPending, time.Now())
	return bookings, err
}

func (s *PostgresStorage) GetEventBookings(ctx context.Context, eventID string) ([]*models.Booking, error) {
	var bookings []*models.Booking
	query := `
		SELECT * FROM bookings 
		WHERE event_id = $1 
		ORDER BY created_at DESC
	`
	err := s.db.SelectContext(ctx, &bookings, query, eventID)
	return bookings, err
}

func (s *PostgresStorage) GetUserBookings(ctx context.Context, email string) ([]*models.BookingWithEvent, error) {
	var bookings []*models.BookingWithEvent
	query := `
		SELECT b.*, e.name as event_name, e.event_date 
		FROM bookings b
		JOIN events e ON b.event_id = e.id
		WHERE b.user_email = $1
		ORDER BY b.created_at DESC
	`
	err := s.db.SelectContext(ctx, &bookings, query, email)
	return bookings, err
}

func (s *PostgresStorage) Close() error {
	return s.db.Close()
}


