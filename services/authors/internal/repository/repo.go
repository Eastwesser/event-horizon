package repository

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/redis/go-redis/v9"

	"github.com/Eastwesser/event-horizon/services/authors/internal/model"
)

type PostgresRepo struct{ db *pgxpool.Pool }

func NewPostgresRepo(db *pgxpool.Pool) *PostgresRepo { return &PostgresRepo{db: db} }

func (r *PostgresRepo) Upsert(ctx context.Context, a *model.Author, eventType string, eventPayload map[string]any) error {
	tx, err := r.db.Begin(ctx)
	if err != nil {
		return err
	}
	defer func() { _ = tx.Rollback(ctx) }()

	now := time.Now().UTC()
	a.UpdatedAt = now
	var existingID string
	err = tx.QueryRow(ctx, `SELECT id FROM authors WHERE user_id = $1`, a.UserID).Scan(&existingID)
	if errors.Is(err, pgx.ErrNoRows) {
		a.ID = uuid.NewString()
		a.CreatedAt = now
		_, err = tx.Exec(ctx, `
			INSERT INTO authors (id, user_id, display_name, bio, avatar_url, active, created_at, updated_at)
			VALUES ($1,$2,$3,$4,$5,$6,$7,$8)`,
			a.ID, a.UserID, a.DisplayName, a.Bio, a.AvatarURL, a.Active, a.CreatedAt, a.UpdatedAt)
	} else if err != nil {
		return err
	} else {
		a.ID = existingID
		_, err = tx.Exec(ctx, `
			UPDATE authors SET display_name=$2, bio=$3, avatar_url=$4, active=$5, updated_at=$6 WHERE user_id=$1`,
			a.UserID, a.DisplayName, a.Bio, a.AvatarURL, a.Active, a.UpdatedAt)
	}
	if err != nil {
		return err
	}

	body, err := json.Marshal(eventPayload)
	if err != nil {
		return err
	}
	_, err = tx.Exec(ctx, `INSERT INTO outbox (id, event_type, payload) VALUES ($1,$2,$3)`,
		uuid.NewString(), eventType, body)
	if err != nil {
		return err
	}
	return tx.Commit(ctx)
}

func (r *PostgresRepo) GetByUserID(ctx context.Context, userID string) (*model.Author, error) {
	row := r.db.QueryRow(ctx, `
		SELECT id, user_id, display_name, bio, avatar_url, active, created_at, updated_at
		FROM authors WHERE user_id = $1`, userID)
	var a model.Author
	if err := row.Scan(&a.ID, &a.UserID, &a.DisplayName, &a.Bio, &a.AvatarURL, &a.Active, &a.CreatedAt, &a.UpdatedAt); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, model.ErrNotFound
		}
		return nil, err
	}
	return &a, nil
}

func (r *PostgresRepo) GetPendingApplicationByUserID(ctx context.Context, userID string) (*model.AuthorApplication, error) {
	row := r.db.QueryRow(ctx, `
		SELECT id, user_id, status, payload, created_at, reviewed_at, reviewed_by, reviewer_note
		FROM author_applications
		WHERE user_id = $1 AND status = 'pending'
		LIMIT 1`, userID)
	return scanApplication(row)
}

func (r *PostgresRepo) GetLatestApplicationByUserID(ctx context.Context, userID string) (*model.AuthorApplication, error) {
	row := r.db.QueryRow(ctx, `
		SELECT id, user_id, status, payload, created_at, reviewed_at, reviewed_by, reviewer_note
		FROM author_applications
		WHERE user_id = $1
		ORDER BY created_at DESC
		LIMIT 1`, userID)
	return scanApplication(row)
}

func (r *PostgresRepo) InsertApplication(ctx context.Context, a *model.AuthorApplication) error {
	body, err := json.Marshal(a.Payload)
	if err != nil {
		return err
	}
	if a.ID == "" {
		a.ID = uuid.NewString()
	}
	if a.CreatedAt.IsZero() {
		a.CreatedAt = time.Now().UTC()
	}
	_, err = r.db.Exec(ctx, `
		INSERT INTO author_applications
			(id, user_id, status, payload, created_at, reviewed_at, reviewed_by, reviewer_note)
		VALUES ($1,$2,$3,$4,$5,$6,$7,$8)`,
		a.ID, a.UserID, string(a.Status), body, a.CreatedAt, a.ReviewedAt, a.ReviewedBy, a.ReviewerNote)
	return err
}

type scannable interface {
	Scan(dest ...any) error
}

func scanApplication(row scannable) (*model.AuthorApplication, error) {
	var (
		a       model.AuthorApplication
		status  string
		payload []byte
		reviewedAt *time.Time
	)
	if err := row.Scan(&a.ID, &a.UserID, &status, &payload, &a.CreatedAt, &reviewedAt, &a.ReviewedBy, &a.ReviewerNote); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, model.ErrApplicationMissing
		}
		return nil, err
	}
	a.Status = model.ApplicationStatus(status)
	a.ReviewedAt = reviewedAt
	if len(payload) > 0 {
		_ = json.Unmarshal(payload, &a.Payload)
	}
	return &a, nil
}

func (r *PostgresRepo) List(ctx context.Context, limit, offset int) ([]*model.Author, int64, error) {
	if limit <= 0 {
		limit = 20
	}
	var total int64
	if err := r.db.QueryRow(ctx, `SELECT COUNT(*) FROM authors WHERE active = true`).Scan(&total); err != nil {
		return nil, 0, err
	}
	rows, err := r.db.Query(ctx, `
		SELECT id, user_id, display_name, bio, avatar_url, active, created_at, updated_at
		FROM authors WHERE active = true
		ORDER BY updated_at DESC LIMIT $1 OFFSET $2`, limit, offset)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()
	var out []*model.Author
	for rows.Next() {
		var a model.Author
		if err := rows.Scan(&a.ID, &a.UserID, &a.DisplayName, &a.Bio, &a.AvatarURL, &a.Active, &a.CreatedAt, &a.UpdatedAt); err != nil {
			return nil, 0, err
		}
		out = append(out, &a)
	}
	return out, total, nil
}

type RedisRepo struct {
	client *redis.Client
	ttl    time.Duration
}

func NewRedisRepo(addr string, ttl time.Duration) *RedisRepo {
	return &RedisRepo{client: redis.NewClient(&redis.Options{Addr: addr, PoolSize: 10}), ttl: ttl}
}

func (r *RedisRepo) Ping(ctx context.Context) error { return r.client.Ping(ctx).Err() }
func (r *RedisRepo) Close() error                   { return r.client.Close() }

func authorKey(userID string) string { return fmt.Sprintf("authors:user:%s", userID) }

func (r *RedisRepo) Get(ctx context.Context, userID string) (*model.Author, error) {
	b, err := r.client.Get(ctx, authorKey(userID)).Bytes()
	if err != nil {
		return nil, err
	}
	var a model.Author
	if err := json.Unmarshal(b, &a); err != nil {
		return nil, err
	}
	return &a, nil
}

func (r *RedisRepo) Set(ctx context.Context, a *model.Author) error {
	b, err := json.Marshal(a)
	if err != nil {
		return err
	}
	return r.client.Set(ctx, authorKey(a.UserID), b, r.ttl).Err()
}

func (r *RedisRepo) Delete(ctx context.Context, userID string) error {
	return r.client.Del(ctx, authorKey(userID)).Err()
}
