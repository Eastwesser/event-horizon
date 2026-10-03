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
	Rank      int
	UserID    string
	UserEmail string
	Nickname  string
	Score     int
	UpdatedAt int64
	Level     int
}

type LeaderboardRepository interface {
	UpdateScore(ctx context.Context, gameID, userID, userEmail, nickname string, score, level int) (int, error)
	UpdateScoreOnly(ctx context.Context, gameID, userID, userEmail string, score, level int) error
	GetTopScores(ctx context.Context, gameID string, limit, level int) ([]ScoreEntry, error)
	GetPlayerRank(ctx context.Context, gameID, userID string, level int) (int, int, error)
	SaveUserInfo(ctx context.Context, gameID, userID, userEmail, nickname string, level int) error
}

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

func normalizeLevel(level int) int {
	if level < 1 {
		return 1
	}
	if level > 100 {
		return 100
	}
	return level
}

func boardKey(gameID string, level int) string {
	return fmt.Sprintf("leaderboard:%s:%d", gameID, normalizeLevel(level))
}

func infoKey(gameID string, level int) string {
	return boardKey(gameID, level) + ":info"
}

func emailKey(gameID string, level int) string {
	return boardKey(gameID, level) + ":emails"
}

func (r *RedisLeaderboardRepo) UpdateScore(ctx context.Context, gameID, userID, userEmail, nickname string, score, level int) (int, error) {
	level = normalizeLevel(level)
	key := boardKey(gameID, level)
	ik := infoKey(gameID, level)

	if userEmail != "" {
		r.client.HSet(ctx, ik, userID+"_email", userEmail)
	}
	if nickname != "" {
		r.client.HSet(ctx, ik, userID+"_nickname", nickname)
	} else if userEmail != "" {
		defaultNick := userEmail
		if idx := strings.Index(defaultNick, "@"); idx > 0 {
			defaultNick = defaultNick[:idx]
		}
		r.client.HSet(ctx, ik, userID+"_nickname", defaultNick)
	}

	if err := r.client.ZAddArgs(ctx, key, redis.ZAddArgs{
		GT:      true,
		Members: []redis.Z{{Score: float64(score), Member: userID}},
	}).Err(); err != nil {
		return 0, err
	}

	if r.db != nil {
		if _, err := r.db.Exec(ctx, `
            INSERT INTO leaderboard_backup (game_id, level, user_id, score, user_email, updated_at)
            VALUES ($1, $2, $3, $4, $5, NOW())
            ON CONFLICT (game_id, level, user_id) DO UPDATE
            SET score = GREATEST(leaderboard_backup.score, EXCLUDED.score),
                user_email = COALESCE(NULLIF(EXCLUDED.user_email, ''), leaderboard_backup.user_email),
                updated_at = NOW()
        `, gameID, level, userID, score, userEmail); err != nil {
			log.Printf("⚠️ Failed to persist leaderboard backup: %v", err)
		}
	}

	rank, err := r.client.ZRevRank(ctx, key, userID).Result()
	if err != nil {
		return 0, err
	}

	return int(rank) + 1, nil
}

func (r *RedisLeaderboardRepo) UpdateScoreOnly(ctx context.Context, gameID, userID, userEmail string, score, level int) error {
	level = normalizeLevel(level)
	key := boardKey(gameID, level)
	ek := emailKey(gameID, level)

	if userEmail != "" {
		r.client.HSet(ctx, ek, userID, userEmail)
	}

	if err := r.client.ZAddArgs(ctx, key, redis.ZAddArgs{
		GT:      true,
		Members: []redis.Z{{Score: float64(score), Member: userID}},
	}).Err(); err != nil {
		return err
	}

	if r.db != nil {
		if _, err := r.db.Exec(ctx, `
            INSERT INTO leaderboard_backup (game_id, level, user_id, score, user_email, updated_at)
            VALUES ($1, $2, $3, $4, $5, NOW())
            ON CONFLICT (game_id, level, user_id) DO UPDATE
            SET score = GREATEST(leaderboard_backup.score, EXCLUDED.score),
                user_email = COALESCE(NULLIF(EXCLUDED.user_email, ''), leaderboard_backup.user_email),
                updated_at = NOW()
        `, gameID, level, userID, score, userEmail); err != nil {
			log.Printf("⚠️ Failed to persist leaderboard backup: %v", err)
		}
	}

	return nil
}

func (r *RedisLeaderboardRepo) GetTopScores(ctx context.Context, gameID string, limit, level int) ([]ScoreEntry, error) {
	level = normalizeLevel(level)
	key := boardKey(gameID, level)
	ik := infoKey(gameID, level)

	results, err := r.client.ZRevRangeWithScores(ctx, key, 0, int64(limit-1)).Result()
	if err != nil {
		return nil, err
	}

	entries := make([]ScoreEntry, 0, len(results))
	for i, result := range results {
		userID := result.Member.(string)
		score := int(result.Score)

		email, _ := r.client.HGet(ctx, ik, userID+"_email").Result()
		nickname, _ := r.client.HGet(ctx, ik, userID+"_nickname").Result()

		if nickname == "" {
			if email != "" {
				nickname = email
				if idx := strings.Index(nickname, "@"); idx > 0 {
					nickname = nickname[:idx]
				}
			} else if len(userID) >= 8 {
				nickname = userID[:8]
			} else {
				nickname = userID
			}
		}

		entries = append(entries, ScoreEntry{
			Rank:      i + 1,
			UserID:    userID,
			UserEmail: email,
			Nickname:  nickname,
			Score:     score,
			UpdatedAt: time.Now().Unix(),
			Level:     level,
		})
	}

	return entries, nil
}

func (r *RedisLeaderboardRepo) GetPlayerRank(ctx context.Context, gameID, userID string, level int) (int, int, error) {
	level = normalizeLevel(level)
	key := boardKey(gameID, level)

	rank, err := r.client.ZRevRank(ctx, key, userID).Result()
	if err != nil {
		if err == redis.Nil {
			return 0, 0, nil
		}
		return 0, 0, err
	}

	score, err := r.client.ZScore(ctx, key, userID).Result()
	if err != nil {
		return 0, 0, err
	}

	return int(rank) + 1, int(score), nil
}

func (r *RedisLeaderboardRepo) SaveUserInfo(ctx context.Context, gameID, userID, userEmail, nickname string, level int) error {
	ik := infoKey(gameID, level)

	if userEmail != "" {
		if err := r.client.HSet(ctx, ik, userID+"_email", userEmail).Err(); err != nil {
			return err
		}
	}

	if nickname != "" {
		if err := r.client.HSet(ctx, ik, userID+"_nickname", nickname).Err(); err != nil {
			return err
		}
	} else if userEmail != "" {
		defaultNick := userEmail
		if idx := strings.Index(defaultNick, "@"); idx > 0 {
			defaultNick = defaultNick[:idx]
		}
		if err := r.client.HSet(ctx, ik, userID+"_nickname", defaultNick).Err(); err != nil {
			return err
		}
	}

	log.Printf("💾 Saved user info: user=%s level=%d email=%s nickname=%s", userID, normalizeLevel(level), userEmail, nickname)
	return nil
}

// RestoreFromPostgres loads records for one game+level into Redis.
func (r *RedisLeaderboardRepo) RestoreFromPostgres(ctx context.Context, gameID string, level int) error {
	level = normalizeLevel(level)
	rows, err := r.db.Query(ctx, `
        SELECT user_id, score
        FROM leaderboard_backup
        WHERE game_id = $1 AND level = $2
        ORDER BY score DESC
    `, gameID, level)
	if err != nil {
		return err
	}
	defer rows.Close()

	pipe := r.client.Pipeline()
	key := boardKey(gameID, level)
	for rows.Next() {
		var userID string
		var score int64
		if err := rows.Scan(&userID, &score); err != nil {
			continue
		}
		pipe.ZAdd(ctx, key, redis.Z{
			Score:  float64(score),
			Member: userID,
		})
	}
	_, err = pipe.Exec(ctx)
	return err
}
