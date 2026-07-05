package repository

import (
    "context"
    "encoding/json"
    "time"

    "github.com/jackc/pgx/v5"
    "github.com/jackc/pgx/v5/pgxpool"
)

type UserProfile struct {
    UserID     string
    Email      string
    Nickname   string
    TotalScore int32
    BestScores map[string]int32
    Lamps      int32
    Tickets    int32
    UpdatedAt  time.Time
}

type ProfileRepository interface {
    GetProfile(ctx context.Context, userID string) (*UserProfile, error)
    UpsertProfile(ctx context.Context, profile *UserProfile) error
}

type PostgresProfileRepo struct {
    db *pgxpool.Pool
}

func NewPostgresProfileRepo(db *pgxpool.Pool) *PostgresProfileRepo {
    return &PostgresProfileRepo{db: db}
}

func (r *PostgresProfileRepo) GetProfile(ctx context.Context, userID string) (*UserProfile, error) {
    var profile UserProfile
    var bestScoresJSON []byte

    query := `SELECT user_id, email, nickname, total_score, best_scores, lamps, tickets, updated_at
              FROM user_profiles WHERE user_id = $1`

    err := r.db.QueryRow(ctx, query, userID).Scan(
        &profile.UserID,
        &profile.Email,
        &profile.Nickname,
        &profile.TotalScore,
        &bestScoresJSON,
        &profile.Lamps,
        &profile.Tickets,
        &profile.UpdatedAt,
    )

    if err != nil {
        if err == pgx.ErrNoRows {
            return nil, nil
        }
        return nil, err
    }

    if err := json.Unmarshal(bestScoresJSON, &profile.BestScores); err != nil {
        profile.BestScores = make(map[string]int32)
    }

    return &profile, nil
}

func (r *PostgresProfileRepo) UpsertProfile(ctx context.Context, profile *UserProfile) error {
    bestScoresJSON, err := json.Marshal(profile.BestScores)
    if err != nil {
        return err
    }

    query := `INSERT INTO user_profiles (user_id, email, nickname, total_score, best_scores, lamps, tickets, updated_at)
              VALUES ($1, $2, $3, $4, $5, $6, $7, NOW())
              ON CONFLICT (user_id) DO UPDATE
              SET email = EXCLUDED.email,
                  nickname = EXCLUDED.nickname,
                  total_score = EXCLUDED.total_score,
                  best_scores = EXCLUDED.best_scores,
                  lamps = EXCLUDED.lamps,
                  tickets = EXCLUDED.tickets,
                  updated_at = NOW()`

    _, err = r.db.Exec(ctx, query,
        profile.UserID,
        profile.Email,
        profile.Nickname,
        profile.TotalScore,
        bestScoresJSON,
        profile.Lamps,
        profile.Tickets,
    )

    return err
}
