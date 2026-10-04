package repository

import (
	"context"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

type Achievement struct {
	Code        string
	Title       string
	Description string
	Icon        string
	UnlockedAt  time.Time
}

type AchievementRepository interface {
	ListUnlocked(ctx context.Context, userID string) ([]Achievement, error)
	Unlock(ctx context.Context, userID string, codes []string) (int64, error)
}

type PostgresAchievementRepo struct {
	db *pgxpool.Pool
}

func NewPostgresAchievementRepo(db *pgxpool.Pool) *PostgresAchievementRepo {
	return &PostgresAchievementRepo{db: db}
}

func (r *PostgresAchievementRepo) ListUnlocked(ctx context.Context, userID string) ([]Achievement, error) {
	rows, err := r.db.Query(ctx, `
		SELECT a.code, a.title, a.description, a.icon, ua.unlocked_at
		FROM user_achievements ua
		JOIN achievements a ON a.code = ua.code
		WHERE ua.user_id = $1
		ORDER BY a.sort_order ASC, ua.unlocked_at ASC`, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	out := make([]Achievement, 0)
	for rows.Next() {
		var a Achievement
		if err := rows.Scan(&a.Code, &a.Title, &a.Description, &a.Icon, &a.UnlockedAt); err != nil {
			return nil, err
		}
		out = append(out, a)
	}
	return out, rows.Err()
}

func (r *PostgresAchievementRepo) Unlock(ctx context.Context, userID string, codes []string) (int64, error) {
	if len(codes) == 0 {
		return 0, nil
	}
	// Ensure profile row exists for FK (no-op if already present).
	_, _ = r.db.Exec(ctx, `
		INSERT INTO user_profiles (user_id, email, nickname, total_score, best_scores, lamps, tickets, updated_at)
		VALUES ($1, '', '', 0, '{}', 0, 0, NOW())
		ON CONFLICT (user_id) DO NOTHING`, userID)

	tag, err := r.db.Exec(ctx, `
		INSERT INTO user_achievements (user_id, code, unlocked_at)
		SELECT $1, c.code, NOW()
		FROM UNNEST($2::text[]) AS c(code)
		INNER JOIN achievements a ON a.code = c.code
		ON CONFLICT (user_id, code) DO NOTHING`, userID, codes)
	if err != nil {
		return 0, err
	}
	return tag.RowsAffected(), nil
}
