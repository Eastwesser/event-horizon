package repository

import (
    "context"
    "database/sql"
)

type GameRepository interface {
    GetHighscore(ctx context.Context, userID, gameID string) (int, error)
    SaveHighscore(ctx context.Context, userID, gameID string, score int) error
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
