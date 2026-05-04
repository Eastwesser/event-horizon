package repository

import (
    "context"
    "fmt"
    "time"

    "github.com/redis/go-redis/v9"
)

type ScoreEntry struct {
	Rank	  int
    UserID    string
    UserEmail string
    Score     int
    UpdatedAt int64
}

type LeaderboardRepository interface {
    UpdateScore(ctx context.Context, gameID, userID, userEmail string, score int) (int, error)
    GetTopScores(ctx context.Context, gameID string, limit int) ([]ScoreEntry, error)
    GetPlayerRank(ctx context.Context, gameID, userID string) (int, int, error)
}

type RedisLeaderboardRepo struct {
    client *redis.Client
}

func NewRedisLeaderboardRepo(addr string, db int) *RedisLeaderboardRepo {
    client := redis.NewClient(&redis.Options{
        Addr: addr,
        DB:   db,
    })
    return &RedisLeaderboardRepo{client: client}
}

func (r *RedisLeaderboardRepo) UpdateScore(ctx context.Context, gameID, userID, userEmail string, score int) (int, error) {
    key := fmt.Sprintf("leaderboard:%s", gameID)
    
    // Добавляем в Sorted Set (score = очки)
    member := &redis.Z{
        Score:  float64(score),
        Member: userID,
    }
    err := r.client.ZAdd(ctx, key, *member).Err()
    if err != nil {
        return 0, err
    }
    
    // Сохраняем email отдельно (для отображения в топе)
    emailKey := fmt.Sprintf("leaderboard:%s:emails", gameID)
    r.client.HSet(ctx, emailKey, userID, userEmail)
    
    // Получаем новый ранг
    rank, err := r.client.ZRevRank(ctx, key, userID).Result()
    if err != nil {
        return 0, err
    }
    
    return int(rank) + 1, nil
}

func (r *RedisLeaderboardRepo) GetTopScores(ctx context.Context, gameID string, limit int) ([]ScoreEntry, error) {
    key := fmt.Sprintf("leaderboard:%s", gameID)
    emailKey := fmt.Sprintf("leaderboard:%s:emails", gameID)
    
    // Получаем топ-N
    results, err := r.client.ZRevRangeWithScores(ctx, key, 0, int64(limit-1)).Result()
    if err != nil {
        return nil, err
    }
    
    entries := make([]ScoreEntry, 0, len(results))
    for i, result := range results {
        userID := result.Member.(string)
        score := int(result.Score)
        
        // Получаем email
        email, _ := r.client.HGet(ctx, emailKey, userID).Result()
        
        entries = append(entries, ScoreEntry{
            Rank:      i + 1,
            UserID:    userID,
            UserEmail: email,
            Score:     score,
            UpdatedAt: time.Now().Unix(),
        })
    }
    
    return entries, nil
}

func (r *RedisLeaderboardRepo) GetPlayerRank(ctx context.Context, gameID, userID string) (int, int, error) {
    key := fmt.Sprintf("leaderboard:%s", gameID)
    
    rank, err := r.client.ZRevRank(ctx, key, userID).Result()
    if err != nil {
        if err == redis.Nil {
            return 0, 0, nil // игрок не в топе
        }
        return 0, 0, err
    }
    
    score, err := r.client.ZScore(ctx, key, userID).Result()
    if err != nil {
        return 0, 0, err
    }
    
    return int(rank) + 1, int(score), nil
}
