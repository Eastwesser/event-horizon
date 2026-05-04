package service

import (
    "context"
    "fmt"
    
    "event_horizon/services/leaderboard/internal/repository"
)

type LeaderboardService interface {
    GetTopScores(ctx context.Context, gameID string, limit int) ([]repository.ScoreEntry, error)
    GetPlayerRank(ctx context.Context, gameID, userID string) (int, int, error)
    UpdateScore(ctx context.Context, gameID, userID, userEmail string, score int) (int, error)
}

type leaderboardService struct {
    repo repository.LeaderboardRepository
}

func NewLeaderboardService(repo repository.LeaderboardRepository) LeaderboardService {
    return &leaderboardService{repo: repo}
}

func (s *leaderboardService) GetTopScores(ctx context.Context, gameID string, limit int) ([]repository.ScoreEntry, error) {
    if limit <= 0 || limit > 100 {
        limit = 10
    }
    return s.repo.GetTopScores(ctx, gameID, limit)
}

func (s *leaderboardService) GetPlayerRank(ctx context.Context, gameID, userID string) (int, int, error) {
    return s.repo.GetPlayerRank(ctx, gameID, userID)
}

func (s *leaderboardService) UpdateScore(ctx context.Context, gameID, userID, userEmail string, score int) (int, error) {
    if gameID == "" || userID == "" {
        return 0, fmt.Errorf("game_id and user_id are required")
    }
    return s.repo.UpdateScore(ctx, gameID, userID, userEmail, score)
}
