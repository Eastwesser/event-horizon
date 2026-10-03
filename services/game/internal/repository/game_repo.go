package repository

import (
	"context"
	"database/sql"
	"encoding/json"
)

type GameRepository interface {
	GetHighscore(ctx context.Context, userID, gameID string) (int, error)
	SaveHighscore(ctx context.Context, userID, gameID string, score int) error
	// EnqueueOutbox inserts a NATS event for the outbox worker (no highscore change).
	EnqueueOutbox(ctx context.Context, eventType string, payload []byte) error
	// SaveHighscoreAndEnqueueOutbox writes highscore + outbox row in one transaction.
	SaveHighscoreAndEnqueueOutbox(ctx context.Context, userID, gameID string, score int, eventType string, payload []byte) error

	CreateRunBoost(ctx context.Context, id, userID, gameID string) error
	GetActiveRunBoost(ctx context.Context, userID, gameID string) (id string, ok bool, err error)
	// ConsumeActiveRunBoost marks the active boost consumed. If boostID is non-empty,
	// only that row is consumed; otherwise any active boost for the pair is consumed.
	ConsumeActiveRunBoost(ctx context.Context, userID, gameID, boostID string) (consumedID string, ok bool, err error)
}

// OutboxRecord is optional metadata for callers that build payloads separately.
type OutboxRecord struct {
	EventType string
	Payload   json.RawMessage
}

type PostgresGameRepo struct {
	db *sql.DB
}

func NewPostgresGameRepo(db *sql.DB) *PostgresGameRepo {
	return &PostgresGameRepo{db: db}
}

func (r *PostgresGameRepo) GetHighscore(ctx context.Context, userID, gameID string) (int, error) {
	var score int
	err := r.db.QueryRowContext(ctx,
		"SELECT COALESCE(score, 0) FROM highscores WHERE user_id = $1 AND game_id = $2",
		userID, gameID,
	).Scan(&score)
	if err == sql.ErrNoRows {
		return 0, nil
	}
	return score, err
}

func (r *PostgresGameRepo) SaveHighscore(ctx context.Context, userID, gameID string, score int) error {
	_, err := r.db.ExecContext(ctx,
		`INSERT INTO highscores (user_id, game_id, score, updated_at)
         VALUES ($1, $2, $3, NOW())
         ON CONFLICT (user_id, game_id) DO UPDATE
         SET score = EXCLUDED.score, updated_at = NOW()`,
		userID, gameID, score,
	)
	return err
}

func (r *PostgresGameRepo) EnqueueOutbox(ctx context.Context, eventType string, payload []byte) error {
	_, err := r.db.ExecContext(ctx,
		`INSERT INTO outbox (event_type, payload) VALUES ($1, $2)`,
		eventType, payload,
	)
	return err
}

func (r *PostgresGameRepo) SaveHighscoreAndEnqueueOutbox(ctx context.Context, userID, gameID string, score int, eventType string, payload []byte) error {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer func() { _ = tx.Rollback() }()

	if _, err := tx.ExecContext(ctx,
		`INSERT INTO highscores (user_id, game_id, score, updated_at)
         VALUES ($1, $2, $3, NOW())
         ON CONFLICT (user_id, game_id) DO UPDATE
         SET score = EXCLUDED.score, updated_at = NOW()`,
		userID, gameID, score,
	); err != nil {
		return err
	}
	if _, err := tx.ExecContext(ctx,
		`INSERT INTO outbox (event_type, payload) VALUES ($1, $2)`,
		eventType, payload,
	); err != nil {
		return err
	}
	return tx.Commit()
}

func (r *PostgresGameRepo) CreateRunBoost(ctx context.Context, id, userID, gameID string) error {
	_, err := r.db.ExecContext(ctx,
		`INSERT INTO run_boosts (id, user_id, game_id, created_at)
         VALUES ($1::uuid, $2, $3, NOW())`,
		id, userID, gameID,
	)
	return err
}

func (r *PostgresGameRepo) GetActiveRunBoost(ctx context.Context, userID, gameID string) (string, bool, error) {
	var id string
	err := r.db.QueryRowContext(ctx,
		`SELECT id::text FROM run_boosts
         WHERE user_id = $1 AND game_id = $2 AND consumed_at IS NULL
         ORDER BY created_at DESC
         LIMIT 1`,
		userID, gameID,
	).Scan(&id)
	if err == sql.ErrNoRows {
		return "", false, nil
	}
	if err != nil {
		return "", false, err
	}
	return id, true, nil
}

func (r *PostgresGameRepo) ConsumeActiveRunBoost(ctx context.Context, userID, gameID, boostID string) (string, bool, error) {
	var id string
	var err error
	if boostID != "" {
		err = r.db.QueryRowContext(ctx,
			`UPDATE run_boosts
             SET consumed_at = NOW()
             WHERE id = $1::uuid AND user_id = $2 AND game_id = $3 AND consumed_at IS NULL
             RETURNING id::text`,
			boostID, userID, gameID,
		).Scan(&id)
	} else {
		err = r.db.QueryRowContext(ctx,
			`UPDATE run_boosts
             SET consumed_at = NOW()
             WHERE id = (
                 SELECT id FROM run_boosts
                 WHERE user_id = $1 AND game_id = $2 AND consumed_at IS NULL
                 ORDER BY created_at DESC
                 LIMIT 1
             )
             RETURNING id::text`,
			userID, gameID,
		).Scan(&id)
	}
	if err == sql.ErrNoRows {
		return "", false, nil
	}
	if err != nil {
		return "", false, err
	}
	return id, true, nil
}
