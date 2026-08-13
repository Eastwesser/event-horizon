package repository

import (
    "context"
    "fmt"
    "log"
    "strings"
    "time"

    "github.com/jackc/pgx/v5/pgxpool"
    "github.com/redis/go-redis/v9"
)

type ScoreEntry struct {
	Rank	  int
    UserID    string
    UserEmail string
    Nickname  string
    Score     int
    UpdatedAt int64
}

type LeaderboardRepository interface {
    UpdateScore(ctx context.Context, gameID, userID, userEmail, nickname string, score int) (int, error)
    UpdateScoreOnly(ctx context.Context, gameID, userID, userEmail string, score int) error
    GetTopScores(ctx context.Context, gameID string, limit int) ([]ScoreEntry, error)
    GetPlayerRank(ctx context.Context, gameID, userID string) (int, int, error)
    SaveUserInfo(ctx context.Context, gameID, userID, userEmail, nickname string) error
}

// type RedisLeaderboardRepo struct {
//     client *redis.Client
// }

// func NewRedisLeaderboardRepo(addr string, db int) *RedisLeaderboardRepo {
//     client := redis.NewClient(&redis.Options{
//         Addr: addr,
//         DB:   db,
//     })
//     return &RedisLeaderboardRepo{client: client}
// }

type RedisLeaderboardRepo struct {
    client *redis.Client
    db     *pgxpool.Pool
}

func NewRedisLeaderboardRepo(addr string, db int, dbPool *pgxpool.Pool) *RedisLeaderboardRepo {
    client := redis.NewClient(&redis.Options{
        Addr: addr,
        DB:   db,
    })
    return &RedisLeaderboardRepo{
        client: client,
        db:     dbPool,
    }
}

func (r *RedisLeaderboardRepo) UpdateScore(ctx context.Context, gameID, userID, userEmail, nickname string, score int) (int, error) {
    key := fmt.Sprintf("leaderboard:%s", gameID)
    infoKey := fmt.Sprintf("leaderboard:%s:info", gameID)
    
    // Сохраняем email и nickname отдельными полями
    if userEmail != "" {
        r.client.HSet(ctx, infoKey, userID+"_email", userEmail)
    }
    if nickname != "" {
        r.client.HSet(ctx, infoKey, userID+"_nickname", nickname)
    } else if userEmail != "" {
        // fallback на email
        defaultNick := userEmail
        if idx := strings.Index(defaultNick, "@"); idx > 0 {
            defaultNick = defaultNick[:idx]
        }
        r.client.HSet(ctx, infoKey, userID+"_nickname", defaultNick)
    }

    // Keep the highest score, never sum. Game publishes a highscore, not a delta.
    if err := r.client.ZAddArgs(ctx, key, redis.ZAddArgs{
        GT:      true,
        Members: []redis.Z{{Score: float64(score), Member: userID}},
    }).Err(); err != nil {
        return 0, err
    }

    if r.db != nil {
        if _, err := r.db.Exec(ctx, `
            INSERT INTO leaderboard_backup (game_id, user_id, score, user_email, updated_at)
            VALUES ($1, $2, $3, $4, NOW())
            ON CONFLICT (game_id, user_id) DO UPDATE
            SET score = GREATEST(leaderboard_backup.score, EXCLUDED.score),
                user_email = COALESCE(NULLIF(EXCLUDED.user_email, ''), leaderboard_backup.user_email),
                updated_at = NOW()
        `, gameID, userID, score, userEmail); err != nil {
            log.Printf("⚠️ Failed to persist leaderboard backup: %v", err)
        }
    }

    rank, err := r.client.ZRevRank(ctx, key, userID).Result()
    if err != nil {
        return 0, err
    }
    
    return int(rank) + 1, nil
}

// UpdateScoreOnly обновляет счёт без получения ранга (быстрее)
func (r *RedisLeaderboardRepo) UpdateScoreOnly(ctx context.Context, gameID, userID, userEmail string, score int) error {
    key := fmt.Sprintf("leaderboard:%s", gameID)
    emailKey := fmt.Sprintf("leaderboard:%s:emails", gameID)
    
    if userEmail != "" {
        r.client.HSet(ctx, emailKey, userID, userEmail)
    }

    if err := r.client.ZAddArgs(ctx, key, redis.ZAddArgs{
        GT:      true,
        Members: []redis.Z{{Score: float64(score), Member: userID}},
    }).Err(); err != nil {
        return err
    }

    if r.db != nil {
        if _, err := r.db.Exec(ctx, `
            INSERT INTO leaderboard_backup (game_id, user_id, score, user_email, updated_at)
            VALUES ($1, $2, $3, $4, NOW())
            ON CONFLICT (game_id, user_id) DO UPDATE
            SET score = GREATEST(leaderboard_backup.score, EXCLUDED.score),
                user_email = COALESCE(NULLIF(EXCLUDED.user_email, ''), leaderboard_backup.user_email),
                updated_at = NOW()
        `, gameID, userID, score, userEmail); err != nil {
            log.Printf("⚠️ Failed to persist leaderboard backup: %v", err)
        }
    }

    return nil
}

func (r *RedisLeaderboardRepo) GetTopScores(ctx context.Context, gameID string, limit int) ([]ScoreEntry, error) {
    key := fmt.Sprintf("leaderboard:%s", gameID)
    infoKey := fmt.Sprintf("leaderboard:%s:info", gameID)
    
    results, err := r.client.ZRevRangeWithScores(ctx, key, 0, int64(limit-1)).Result()
    if err != nil {
        return nil, err
    }
    
    entries := make([]ScoreEntry, 0, len(results))
    for i, result := range results {
        userID := result.Member.(string)
        score := int(result.Score)
        
        // Получаем email и nickname через HGet (а не HGetAll)
        email, _ := r.client.HGet(ctx, infoKey, userID+"_email").Result()
        nickname, _ := r.client.HGet(ctx, infoKey, userID+"_nickname").Result()
        
        if nickname == "" {
            if email != "" {
                nickname = email
                if idx := strings.Index(nickname, "@"); idx > 0 {
                    nickname = nickname[:idx]
                }
            } else {
                nickname = userID[:8] // fallback на часть userID
            }
        }
        
        entries = append(entries, ScoreEntry{
            Rank:      i + 1,
            UserID:    userID,
            UserEmail: email,
            Nickname:  nickname,
            Score:     score,
            UpdatedAt: time.Now().Unix(),
        })
    }
    
    return entries, nil
}

// GetPlayerRank отдельный вызов, когда нужен ранг
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

// SaveUserInfo сохраняет email и nickname пользователя
func (r *RedisLeaderboardRepo) SaveUserInfo(ctx context.Context, gameID, userID, userEmail, nickname string) error {
    infoKey := fmt.Sprintf("leaderboard:%s:info", gameID)
    
    if userEmail != "" {
        if err := r.client.HSet(ctx, infoKey, userID+"_email", userEmail).Err(); err != nil {
            return err
        }
    }
    
    if nickname != "" {
        if err := r.client.HSet(ctx, infoKey, userID+"_nickname", nickname).Err(); err != nil {
            return err
        }
    } else if userEmail != "" {
        // fallback на email
        defaultNick := userEmail
        if idx := strings.Index(defaultNick, "@"); idx > 0 {
            defaultNick = defaultNick[:idx]
        }
        if err := r.client.HSet(ctx, infoKey, userID+"_nickname", defaultNick).Err(); err != nil {
            return err
        }
    }
    
    log.Printf("💾 Saved user info: user=%s, email=%s, nickname=%s", userID, userEmail, nickname)
    return nil
}

// RestoreFromPostgres загружает все рекорды из PostgreSQL в Redis
func (r *RedisLeaderboardRepo) RestoreFromPostgres(ctx context.Context, gameID string) error {
    // Запрос к PostgreSQL (через gRPC или прямой доступ к БД)
    rows, err := r.db.Query(ctx, `
        SELECT user_id, score
        FROM leaderboard_backup
        WHERE game_id = $1
        ORDER BY score DESC
    `, gameID)
    if err != nil {
        return err
    }
    defer rows.Close()

    pipe := r.client.Pipeline()
    for rows.Next() {
        var userID string
        var score int64
        if err := rows.Scan(&userID, &score); err != nil {
            continue
        }
        pipe.ZAdd(ctx, "leaderboard:"+gameID, redis.Z{
            Score:  float64(score),
            Member: userID,
        })
    }
    _, err = pipe.Exec(ctx)
    return err
}
