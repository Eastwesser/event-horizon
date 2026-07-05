package repository

import (
	"context"
	"errors"
	"time"
	
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type User struct {
    ID           string
    Email        string
    PasswordHash string
    CreatedAt    time.Time
	Nickname	 string
}

type UserRepository interface {
	Create(ctx context.Context, email, passwordHash string) (string, error)
	GetByEmail(ctx context.Context, email string) (*User, error)
	GetByID(ctx context.Context, id string) (*User, error)
	UpdateNickname(ctx context.Context, userID, nickname string) error
	GetUserScores(ctx context.Context, userID string) (map[string]int32, int32, error)
}

type PostgresUserRepo struct {
	db *pgxpool.Pool
}

func NewPostgresUserRepo(db *pgxpool.Pool) *PostgresUserRepo {
	return &PostgresUserRepo{db: db}
}

func (r *PostgresUserRepo) Create(ctx context.Context, email, passwordHash string) (string, error) {
	var userID string
	query := `INSERT INTO users (email, password_hash) VALUES ($1, $2) RETURNING id`
	
	err := r.db.QueryRow(ctx, query, email, passwordHash).Scan(&userID)
	if err != nil {
		return "", err
	}
	return userID, nil
}

func (r *PostgresUserRepo) GetByEmail(ctx context.Context, email string) (*User, error) {
	user := &User{}
	query := `SELECT id, email, password_hash, nickname, created_at FROM users WHERE email = $1`
	err := r.db.QueryRow(ctx, query, email).Scan(
        &user.ID, 
        &user.Email, 
        &user.PasswordHash, 
        &user.CreatedAt,  // теперь time.Time
    )
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}
	return user, nil
}

func (r *PostgresUserRepo) GetByID(ctx context.Context, id string) (*User, error) {
	user := &User{}
	query := `SELECT id, email, password_hash, nickname, created_at FROM users WHERE email = $1`
	err := r.db.QueryRow(ctx, query, id).Scan(
        &user.ID, 
        &user.Email, 
        &user.PasswordHash, 
        &user.CreatedAt,  // теперь time.Time
    )
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}
	return user, nil
}

// UpdateNickname обновляет никнейм пользователя
func (r *PostgresUserRepo) UpdateNickname(ctx context.Context, userID, nickname string) error {
    query := `UPDATE users SET nickname = $1 WHERE id = $2`
    _, err := r.db.Exec(ctx, query, nickname, userID)
    return err
}

// GetUserScores возвращает рекорды пользователя по играм
func (r *PostgresUserRepo) GetUserScores(ctx context.Context, userID string) (map[string]int32, int32, error) {
    // Получаем лучшие рекорды по играм
    rows, err := r.db.Query(ctx, `
        SELECT game_id, MAX(score) as best_score
        FROM scores
        WHERE user_id = $1
        GROUP BY game_id
    `, userID)
    if err != nil {
        return nil, 0, err
    }
    defer rows.Close()

    bestScores := make(map[string]int32)
    var totalScore int32 = 0

    for rows.Next() {
        var gameID string
        var best int32
        if err := rows.Scan(&gameID, &best); err != nil {
            return nil, 0, err
        }
        bestScores[gameID] = best
        totalScore += best
    }

    if err := rows.Err(); err != nil {
        return nil, 0, err
    }

    return bestScores, totalScore, nil
}
