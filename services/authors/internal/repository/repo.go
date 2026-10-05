package repository

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
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
			INSERT INTO authors (id, user_id, display_name, bio, avatar_url, portfolio, active, created_at, updated_at, verified_at)
			VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10)`,
			a.ID, a.UserID, a.DisplayName, a.Bio, a.AvatarURL, a.Portfolio, a.Active, a.CreatedAt, a.UpdatedAt, a.VerifiedAt)
	} else if err != nil {
		return err
	} else {
		a.ID = existingID
		_, err = tx.Exec(ctx, `
			UPDATE authors SET display_name=$2, bio=$3, avatar_url=$4, portfolio=$5, active=$6, updated_at=$7, verified_at=$8
			WHERE user_id=$1`,
			a.UserID, a.DisplayName, a.Bio, a.AvatarURL, a.Portfolio, a.Active, a.UpdatedAt, a.VerifiedAt)
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
		SELECT id, user_id, display_name, bio, avatar_url, portfolio, active, created_at, updated_at, verified_at
		FROM authors WHERE user_id = $1`, userID)
	return scanAuthor(row)
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

func (r *PostgresRepo) GetApplicationByID(ctx context.Context, id string) (*model.AuthorApplication, error) {
	row := r.db.QueryRow(ctx, `
		SELECT id, user_id, status, payload, created_at, reviewed_at, reviewed_by, reviewer_note
		FROM author_applications WHERE id = $1`, id)
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

func (r *PostgresRepo) ListApplications(ctx context.Context, status string, limit, offset int) ([]*model.AuthorApplication, int64, error) {
	if limit <= 0 {
		limit = 50
	}
	if offset < 0 {
		offset = 0
	}
	status = strings.ToLower(strings.TrimSpace(status))

	var (
		total int64
		rows  pgx.Rows
		err   error
	)
	if status == "" {
		if err := r.db.QueryRow(ctx, `SELECT COUNT(*) FROM author_applications`).Scan(&total); err != nil {
			return nil, 0, err
		}
		rows, err = r.db.Query(ctx, `
			SELECT id, user_id, status, payload, created_at, reviewed_at, reviewed_by, reviewer_note
			FROM author_applications
			ORDER BY created_at DESC
			LIMIT $1 OFFSET $2`, limit, offset)
	} else {
		if err := r.db.QueryRow(ctx, `SELECT COUNT(*) FROM author_applications WHERE status = $1`, status).Scan(&total); err != nil {
			return nil, 0, err
		}
		rows, err = r.db.Query(ctx, `
			SELECT id, user_id, status, payload, created_at, reviewed_at, reviewed_by, reviewer_note
			FROM author_applications
			WHERE status = $1
			ORDER BY created_at DESC
			LIMIT $2 OFFSET $3`, status, limit, offset)
	}
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	out := make([]*model.AuthorApplication, 0)
	for rows.Next() {
		app, err := scanApplication(rows)
		if err != nil {
			return nil, 0, err
		}
		out = append(out, app)
	}
	return out, total, rows.Err()
}

// ApproveApplicationInTx marks pending→approved and upserts authors profile in one transaction.
func (r *PostgresRepo) ApproveApplicationInTx(ctx context.Context, applicationID, reviewerID string) (*model.AuthorApplication, *model.Author, error) {
	tx, err := r.db.Begin(ctx)
	if err != nil {
		return nil, nil, err
	}
	defer func() { _ = tx.Rollback(ctx) }()

	row := tx.QueryRow(ctx, `
		SELECT id, user_id, status, payload, created_at, reviewed_at, reviewed_by, reviewer_note
		FROM author_applications WHERE id = $1 FOR UPDATE`, applicationID)
	app, err := scanApplication(row)
	if err != nil {
		return nil, nil, err
	}
	if app.Status != model.ApplicationPending {
		return nil, nil, model.ErrAlreadyReviewed
	}

	now := time.Now().UTC()
	app.Status = model.ApplicationApproved
	app.ReviewedAt = &now
	app.ReviewedBy = reviewerID
	_, err = tx.Exec(ctx, `
		UPDATE author_applications
		SET status = 'approved', reviewed_at = $2, reviewed_by = $3, reviewer_note = ''
		WHERE id = $1`, applicationID, now, reviewerID)
	if err != nil {
		return nil, nil, err
	}

	author := &model.Author{
		UserID:      app.UserID,
		DisplayName: app.Payload.DisplayName,
		Bio:         app.Payload.Motivation,
		Portfolio:   app.Payload.Portfolio,
		Active:      true,
		VerifiedAt:  &now,
		UpdatedAt:   now,
	}

	var existingID string
	err = tx.QueryRow(ctx, `SELECT id FROM authors WHERE user_id = $1`, author.UserID).Scan(&existingID)
	if errors.Is(err, pgx.ErrNoRows) {
		author.ID = uuid.NewString()
		author.CreatedAt = now
		_, err = tx.Exec(ctx, `
			INSERT INTO authors (id, user_id, display_name, bio, avatar_url, portfolio, active, created_at, updated_at, verified_at)
			VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10)`,
			author.ID, author.UserID, author.DisplayName, author.Bio, author.AvatarURL, author.Portfolio,
			author.Active, author.CreatedAt, author.UpdatedAt, author.VerifiedAt)
	} else if err != nil {
		return nil, nil, err
	} else {
		author.ID = existingID
		author.CreatedAt = now // filled from DB below if needed; keep updated
		_, err = tx.Exec(ctx, `
			UPDATE authors
			SET display_name=$2, bio=$3, portfolio=$4, active=true, updated_at=$5, verified_at=$6
			WHERE user_id=$1`,
			author.UserID, author.DisplayName, author.Bio, author.Portfolio, author.UpdatedAt, author.VerifiedAt)
		if err == nil {
			_ = tx.QueryRow(ctx, `SELECT created_at FROM authors WHERE user_id = $1`, author.UserID).Scan(&author.CreatedAt)
		}
	}
	if err != nil {
		return nil, nil, err
	}

	event := map[string]any{
		"event":        "author.upserted",
		"user_id":      author.UserID,
		"display_name": author.DisplayName,
		"source":       "application_approved",
		"application_id": applicationID,
		"timestamp":    now.Unix(),
	}
	body, err := json.Marshal(event)
	if err != nil {
		return nil, nil, err
	}
	_, err = tx.Exec(ctx, `INSERT INTO outbox (id, event_type, payload) VALUES ($1,$2,$3)`,
		uuid.NewString(), "author.upserted", body)
	if err != nil {
		return nil, nil, err
	}

	if err := tx.Commit(ctx); err != nil {
		return nil, nil, err
	}
	return app, author, nil
}

func (r *PostgresRepo) RejectApplication(ctx context.Context, applicationID, reviewerID, note string) (*model.AuthorApplication, error) {
	tx, err := r.db.Begin(ctx)
	if err != nil {
		return nil, err
	}
	defer func() { _ = tx.Rollback(ctx) }()

	row := tx.QueryRow(ctx, `
		SELECT id, user_id, status, payload, created_at, reviewed_at, reviewed_by, reviewer_note
		FROM author_applications WHERE id = $1 FOR UPDATE`, applicationID)
	app, err := scanApplication(row)
	if err != nil {
		return nil, err
	}
	if app.Status != model.ApplicationPending {
		return nil, model.ErrAlreadyReviewed
	}

	now := time.Now().UTC()
	app.Status = model.ApplicationRejected
	app.ReviewedAt = &now
	app.ReviewedBy = reviewerID
	app.ReviewerNote = note
	_, err = tx.Exec(ctx, `
		UPDATE author_applications
		SET status = 'rejected', reviewed_at = $2, reviewed_by = $3, reviewer_note = $4
		WHERE id = $1`, applicationID, now, reviewerID, note)
	if err != nil {
		return nil, err
	}
	if err := tx.Commit(ctx); err != nil {
		return nil, err
	}
	return app, nil
}

// RevertApplication restores approved→pending and deletes the authors profile (Auth.UpdateRole compensation).
func (r *PostgresRepo) RevertApplication(ctx context.Context, applicationID string) (*model.AuthorApplication, error) {
	tx, err := r.db.Begin(ctx)
	if err != nil {
		return nil, err
	}
	defer func() { _ = tx.Rollback(ctx) }()

	row := tx.QueryRow(ctx, `
		SELECT id, user_id, status, payload, created_at, reviewed_at, reviewed_by, reviewer_note
		FROM author_applications WHERE id = $1 FOR UPDATE`, applicationID)
	app, err := scanApplication(row)
	if err != nil {
		return nil, err
	}
	if app.Status != model.ApplicationApproved {
		return nil, model.ErrInvalidInput
	}

	_, err = tx.Exec(ctx, `
		UPDATE author_applications
		SET status = 'pending', reviewed_at = NULL, reviewed_by = '', reviewer_note = ''
		WHERE id = $1`, applicationID)
	if err != nil {
		return nil, err
	}
	_, err = tx.Exec(ctx, `DELETE FROM authors WHERE user_id = $1`, app.UserID)
	if err != nil {
		return nil, err
	}

	app.Status = model.ApplicationPending
	app.ReviewedAt = nil
	app.ReviewedBy = ""
	app.ReviewerNote = ""

	if err := tx.Commit(ctx); err != nil {
		return nil, err
	}
	return app, nil
}

type scannable interface {
	Scan(dest ...any) error
}

func scanAuthor(row scannable) (*model.Author, error) {
	var a model.Author
	var verifiedAt *time.Time
	if err := row.Scan(
		&a.ID, &a.UserID, &a.DisplayName, &a.Bio, &a.AvatarURL, &a.Portfolio,
		&a.Active, &a.CreatedAt, &a.UpdatedAt, &verifiedAt,
	); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, model.ErrNotFound
		}
		return nil, err
	}
	a.VerifiedAt = verifiedAt
	return &a, nil
}

func scanApplication(row scannable) (*model.AuthorApplication, error) {
	var (
		a          model.AuthorApplication
		status     string
		payload    []byte
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
		SELECT id, user_id, display_name, bio, avatar_url, portfolio, active, created_at, updated_at, verified_at
		FROM authors WHERE active = true
		ORDER BY updated_at DESC LIMIT $1 OFFSET $2`, limit, offset)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()
	var out []*model.Author
	for rows.Next() {
		a, err := scanAuthor(rows)
		if err != nil {
			return nil, 0, err
		}
		out = append(out, a)
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
