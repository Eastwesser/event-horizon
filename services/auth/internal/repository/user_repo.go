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
}

type UserRepository interface {
	Create(ctx context.Context, email, passwordHash string) (string, error)
	GetByEmail(ctx context.Context, email string) (*User, error)
	GetByID(ctx context.Context, id string) (*User, error)
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
	query := `SELECT id, email, password_hash, created_at FROM users WHERE email = $1`
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
	query := `SELECT id, email, password_hash, created_at FROM users WHERE id = $1`
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
