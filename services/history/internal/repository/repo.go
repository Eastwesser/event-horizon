package repository

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/Eastwesser/event-horizon/services/history/internal/model"
)

type PostgresRepo struct{ db *pgxpool.Pool }

func NewPostgresRepo(db *pgxpool.Pool) *PostgresRepo { return &PostgresRepo{db: db} }

func (r *PostgresRepo) Insert(ctx context.Context, userID, eventType, payload string) (string, error) {
	id := uuid.NewString()
	_, err := r.db.Exec(ctx, `
		INSERT INTO events (id, user_id, event_type, payload, created_at)
		VALUES ($1,$2,$3,$4::jsonb,$5)`,
		id, userID, eventType, payload, time.Now().UTC())
	return id, err
}

func (r *PostgresRepo) List(ctx context.Context, userID, eventType string, limit, offset int) ([]*model.Event, int64, error) {
	if limit <= 0 {
		limit = 50
	}
	var total int64
	countQ := `SELECT COUNT(*) FROM events WHERE ($1 = '' OR user_id = $1) AND ($2 = '' OR event_type = $2)`
	if err := r.db.QueryRow(ctx, countQ, userID, eventType).Scan(&total); err != nil {
		return nil, 0, err
	}
	rows, err := r.db.Query(ctx, `
		SELECT id, user_id, event_type, payload::text, created_at FROM events
		WHERE ($1 = '' OR user_id = $1) AND ($2 = '' OR event_type = $2)
		ORDER BY created_at DESC LIMIT $3 OFFSET $4`, userID, eventType, limit, offset)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()
	var out []*model.Event
	for rows.Next() {
		var e model.Event
		if err := rows.Scan(&e.ID, &e.UserID, &e.EventType, &e.Payload, &e.CreatedAt); err != nil {
			return nil, 0, err
		}
		out = append(out, &e)
	}
	return out, total, nil
}

func (r *PostgresRepo) DeleteOlderThan(ctx context.Context, before time.Time) (int64, error) {
	tag, err := r.db.Exec(ctx, `DELETE FROM events WHERE created_at < $1`, before)
	if err != nil {
		return 0, err
	}
	return tag.RowsAffected(), nil
}
