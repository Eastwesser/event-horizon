package repository

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/Eastwesser/event-horizon/services/notification/internal/model"
)

type PostgresRepo struct{ db *pgxpool.Pool }

func NewPostgresRepo(db *pgxpool.Pool) *PostgresRepo { return &PostgresRepo{db: db} }

func (r *PostgresRepo) Insert(ctx context.Context, n *model.Notification) (string, error) {
	if n.ID == "" {
		n.ID = uuid.NewString()
	}
	if n.CreatedAt.IsZero() {
		n.CreatedAt = time.Now().UTC()
	}
	_, err := r.db.Exec(ctx, `
		INSERT INTO notifications (id, user_id, title, body, link, source_event, created_at)
		VALUES ($1,$2,$3,$4,$5,$6,$7)
		ON CONFLICT (user_id, source_event) WHERE source_event <> '' DO NOTHING`,
		n.ID, n.UserID, n.Title, n.Body, n.Link, n.Source, n.CreatedAt)
	return n.ID, err
}

func (r *PostgresRepo) List(ctx context.Context, userID string, limit, offset int, unreadOnly bool) ([]*model.Notification, int64, int64, error) {
	if limit <= 0 {
		limit = 50
	}
	if offset < 0 {
		offset = 0
	}
	var unread int64
	if err := r.db.QueryRow(ctx, `
		SELECT COUNT(*) FROM notifications WHERE user_id = $1 AND read_at IS NULL`, userID).Scan(&unread); err != nil {
		return nil, 0, 0, err
	}
	var total int64
	countQ := `SELECT COUNT(*) FROM notifications WHERE user_id = $1`
	if unreadOnly {
		countQ += ` AND read_at IS NULL`
	}
	if err := r.db.QueryRow(ctx, countQ, userID).Scan(&total); err != nil {
		return nil, 0, 0, err
	}
	q := `
		SELECT id, user_id, title, body, link, source_event, created_at, read_at
		FROM notifications WHERE user_id = $1`
	if unreadOnly {
		q += ` AND read_at IS NULL`
	}
	q += ` ORDER BY created_at DESC LIMIT $2 OFFSET $3`
	rows, err := r.db.Query(ctx, q, userID, limit, offset)
	if err != nil {
		return nil, 0, 0, err
	}
	defer rows.Close()
	var out []*model.Notification
	for rows.Next() {
		var n model.Notification
		var readAt *time.Time
		if err := rows.Scan(&n.ID, &n.UserID, &n.Title, &n.Body, &n.Link, &n.Source, &n.CreatedAt, &readAt); err != nil {
			return nil, 0, 0, err
		}
		n.ReadAt = readAt
		out = append(out, &n)
	}
	return out, total, unread, rows.Err()
}

func (r *PostgresRepo) MarkRead(ctx context.Context, userID, notificationID string) (int64, error) {
	now := time.Now().UTC()
	if notificationID == "" {
		tag, err := r.db.Exec(ctx, `
			UPDATE notifications SET read_at = $1
			WHERE user_id = $2 AND read_at IS NULL`, now, userID)
		if err != nil {
			return 0, err
		}
		return tag.RowsAffected(), nil
	}
	tag, err := r.db.Exec(ctx, `
		UPDATE notifications SET read_at = $1
		WHERE id = $2 AND user_id = $3 AND read_at IS NULL`, now, notificationID, userID)
	if err != nil {
		return 0, err
	}
	n := tag.RowsAffected()
	if n == 0 {
		var exists bool
		if err := r.db.QueryRow(ctx,
			`SELECT EXISTS(SELECT 1 FROM notifications WHERE id=$1 AND user_id=$2)`,
			notificationID, userID).Scan(&exists); err != nil {
			return 0, err
		}
		if !exists {
			return 0, model.ErrNotFound
		}
	}
	return n, nil
}
