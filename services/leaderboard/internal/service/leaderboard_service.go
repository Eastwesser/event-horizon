package service

import (
	"context"
	"fmt"

	"github.com/Eastwesser/event-horizon/services/leaderboard/internal/repository"
)

type LeaderboardService interface {
	GetTopScores(ctx context.Context, gameID string, limit, level int) ([]repository.ScoreEntry, error)
	GetPlayerRank(ctx context.Context, gameID, userID string, level int) (int, int, error)
	UpdateScore(ctx context.Context, gameID, userID, userEmail, nickname string, score, level int) (int, error)
	UpdateScoreOnly(ctx context.Context, gameID, userID, userEmail string, score, level int) error
	SaveUserInfo(ctx context.Context, gameID, userID, userEmail, nickname string, level int) error
}

type leaderboardService struct {
	repo repository.LeaderboardRepository
}

func NewLeaderboardService(repo repository.LeaderboardRepository) LeaderboardService {
	return &leaderboardService{repo: repo}
}

func (s *leaderboardService) GetTopScores(ctx context.Context, gameID string, limit, level int) ([]repository.ScoreEntry, error) {
	if limit <= 0 || limit > 100 {
		limit = 10
	}
	return s.repo.GetTopScores(ctx, gameID, limit, level)
}

func (s *leaderboardService) GetPlayerRank(ctx context.Context, gameID, userID string, level int) (int, int, error) {
	return s.repo.GetPlayerRank(ctx, gameID, userID, level)
}

func (s *leaderboardService) UpdateScore(ctx context.Context, gameID, userID, userEmail, nickname string, score, level int) (int, error) {
	if gameID == "" || userID == "" {
		return 0, fmt.Errorf("game_id and user_id are required")
	}
	return s.repo.UpdateScore(ctx, gameID, userID, userEmail, nickname, score, level)
}

func (s *leaderboardService) UpdateScoreOnly(ctx context.Context, gameID, userID, userEmail string, score, level int) error {
	return s.repo.UpdateScoreOnly(ctx, gameID, userID, userEmail, score, level)
}

func (s *leaderboardService) SaveUserInfo(ctx context.Context, gameID, userID, userEmail, nickname string, level int) error {
	return s.repo.SaveUserInfo(ctx, gameID, userID, userEmail, nickname, level)
}
