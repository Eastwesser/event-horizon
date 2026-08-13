package repository

import (
	"context"
	"errors"
	"strings"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/Eastwesser/event-horizon/pkg/sqb"
	"github.com/Eastwesser/event-horizon/services/auth/internal/model"
)

type User = model.User

//go:generate mockery --name=UserRepository --dir=. --output=mocks --outpkg=mocks --filename=user_repository_mock.go --with-expecter
type UserRepository interface {
	Create(ctx context.Context, email, passwordHash, role string) (string, error)
	GetByEmail(ctx context.Context, email string) (*User, error)
	GetByID(ctx context.Context, id string) (*User, error)
	UpdateNickname(ctx context.Context, userID, nickname string) error
	UpdateRole(ctx context.Context, userID, role string) error
	GetUserScores(ctx context.Context, userID string) (map[string]int32, int32, error)
}

type PostgresUserRepo struct {
	db *pgxpool.Pool
}

func NewPostgresUserRepo(db *pgxpool.Pool) *PostgresUserRepo {
	return &PostgresUserRepo{db: db}
}

func (r *PostgresUserRepo) Create(ctx context.Context, email, passwordHash, role string) (string, error) {
	nickname := strings.Split(email, "@")[0]
	query, args, err := sqb.Insert("users").
		Columns("email", "password_hash", "nickname", "role").
		Values(email, passwordHash, nickname, role).
		Suffix("RETURNING id").
		ToSql()
	if err != nil {
		return "", err
	}
	var userID string
	if err := r.db.QueryRow(ctx, query, args...).Scan(&userID); err != nil {
		return "", err
	}
	return userID, nil
}

func (r *PostgresUserRepo) GetByEmail(ctx context.Context, email string) (*User, error) {
	query, args, err := sqb.Select("id", "email", "password_hash", "nickname", "role", "created_at").
		From("users").
		Where("email = $1", email).
		ToSql()
	if err != nil {
		return nil, err
	}
	user := &User{}
	err = r.db.QueryRow(ctx, query, args...).Scan(
		&user.ID, &user.Email, &user.PasswordHash, &user.Nickname, &user.Role, &user.CreatedAt,
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
	query, args, err := sqb.Select("id", "email", "password_hash", "nickname", "role", "created_at").
		From("users").
		Where("id = $1", id).
		ToSql()
	if err != nil {
		return nil, err
	}
	user := &User{}
	err = r.db.QueryRow(ctx, query, args...).Scan(
		&user.ID, &user.Email, &user.PasswordHash, &user.Nickname, &user.Role, &user.CreatedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}
	return user, nil
}

func (r *PostgresUserRepo) UpdateNickname(ctx context.Context, userID, nickname string) error {
	query, args, err := sqb.Update("users").
		Set("nickname", nickname).
		SetRaw("updated_at = CURRENT_TIMESTAMP").
		Where("id = $1", userID).
		ToSql()
	if err != nil {
		return err
	}
	_, err = r.db.Exec(ctx, query, args...)
	return err
}

func (r *PostgresUserRepo) UpdateRole(ctx context.Context, userID, role string) error {
	query, args, err := sqb.Update("users").
		Set("role", role).
		SetRaw("updated_at = CURRENT_TIMESTAMP").
		Where("id = $1", userID).
		ToSql()
	if err != nil {
		return err
	}
	tag, err := r.db.Exec(ctx, query, args...)
	if err != nil {
		return err
	}
	if tag.RowsAffected() == 0 {
		return errors.New("user not found")
	}
	return nil
}

func (r *PostgresUserRepo) GetUserScores(ctx context.Context, userID string) (map[string]int32, int32, error) {
	query, args, err := sqb.Select("game_id", "MAX(score) as best_score").
		From("scores").
		Where("user_id = $1", userID).
		GroupBy("game_id").
		ToSql()
	if err != nil {
		return nil, 0, err
	}
	rows, err := r.db.Query(ctx, query, args...)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	bestScores := make(map[string]int32)
	var totalScore int32
	for rows.Next() {
		var gameID string
		var best int32
		if err := rows.Scan(&gameID, &best); err != nil {
			return nil, 0, err
		}
		bestScores[gameID] = best
		totalScore += best
	}
	return bestScores, totalScore, rows.Err()
}
