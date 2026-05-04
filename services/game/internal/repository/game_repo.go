package repository

import (
    "context"
)

type GameRepository interface {
    GetHighscore(ctx context.Context, userID, gameID string) (int, error)
    SaveHighscore(ctx context.Context, userID, gameID string, score int) error
}

type PostgresGameRepo struct {
    // TODO: добавить подключение к PostgreSQL
}

func NewPostgresGameRepo() *PostgresGameRepo {
    return &PostgresGameRepo{}
}

func (r *PostgresGameRepo) GetHighscore(ctx context.Context, userID, gameID string) (int, error) {
    // TODO: реализовать
    return 0, nil
}

func (r *PostgresGameRepo) SaveHighscore(ctx context.Context, userID, gameID string, score int) error {
    // TODO: реализовать
    return nil
}
