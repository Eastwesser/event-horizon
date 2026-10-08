package service

import (
	"context"
	"fmt"
	"log"

	"github.com/Eastwesser/event-horizon/services/profile/internal/repository"
)

// mainGames — first_play_* catalog (Wave 2 #5).
var mainGames = []string{
	"flappy", "hexagon", "memory", "towers", "hanoi", "twenty48", "gears", "companion",
}

// Per-game score tiers: любитель / профессионал / герой (новичок = first_play_*).
var gameScoreTiers = map[string][3]int32{
	"flappy":    {25, 100, 500},
	"hexagon":   {50, 200, 1000},
	"memory":    {200, 500, 900},
	"towers":    {10, 30, 80},
	"hanoi":     {100, 500, 2000},
	"twenty48":  {512, 2048, 8192},
	"gears":     {50, 200, 500},
	"companion": {1000, 5000, 15000},
}

// EvaluateAchievements derives unlock codes from best_scores / total_score / optional event level.
// totalScore should be the sum of best_scores (already maintained by the score.updated consumer).
func EvaluateAchievements(bestScores map[string]int32, totalScore int32, gameID string, level int) []string {
	codes := make([]string, 0, 40)
	if bestScores == nil {
		bestScores = map[string]int32{}
	}

	for _, g := range mainGames {
		if _, ok := bestScores[g]; ok {
			codes = append(codes, "first_play_"+g)
		}
	}
	// First play from this event even if not yet a stored best (edge: non-record).
	if gameID != "" {
		codes = append(codes, "first_play_"+gameID)
	}

	for game, tiers := range gameScoreTiers {
		score := bestScores[game]
		if score >= tiers[0] {
			codes = append(codes, game+"_amateur")
		}
		if score >= tiers[1] {
			codes = append(codes, game+"_pro")
		}
		if score >= tiers[2] {
			codes = append(codes, game+"_hero")
		}
	}

	// Legacy flappy thresholds (kept for already-unlocked badges).
	if bestScores["flappy"] >= 100 {
		codes = append(codes, "flappy_score_100")
	}
	if bestScores["flappy"] >= 500 {
		codes = append(codes, "flappy_score_500")
	}

	if gameID == "flappy" {
		if level >= 5 {
			codes = append(codes, "flappy_level_5")
		}
		if level >= 10 {
			codes = append(codes, "flappy_level_10")
		}
	}

	if totalScore >= 1000 {
		codes = append(codes, "total_score_1k")
	}
	if totalScore >= 5000 {
		codes = append(codes, "total_score_5k")
	}
	if totalScore >= 10000 {
		codes = append(codes, "total_score_10k")
	}

	return uniqueStrings(codes)
}

func uniqueStrings(in []string) []string {
	seen := make(map[string]struct{}, len(in))
	out := make([]string, 0, len(in))
	for _, s := range in {
		if s == "" {
			continue
		}
		if _, ok := seen[s]; ok {
			continue
		}
		seen[s] = struct{}{}
		out = append(out, s)
	}
	return out
}

func (s *profileService) EvaluateAndUnlock(ctx context.Context, userID string, bestScores map[string]int32, totalScore int32, gameID string, level int) error {
	if s.achievements == nil || userID == "" {
		return nil
	}
	codes := EvaluateAchievements(bestScores, totalScore, gameID, level)
	return s.UnlockCodes(ctx, userID, codes)
}

// UnlockCodes inserts catalog codes for the user (idempotent via ON CONFLICT).
func (s *profileService) UnlockCodes(ctx context.Context, userID string, codes []string) error {
	if s.achievements == nil || userID == "" || len(codes) == 0 {
		return nil
	}
	n, err := s.achievements.Unlock(ctx, userID, codes)
	if err != nil {
		return fmt.Errorf("unlock achievements: %w", err)
	}
	if n > 0 {
		log.Printf("achievements unlocked user=%s count=%d codes=%v", userID, n, codes)
	}
	return nil
}

func (s *profileService) ListAchievements(ctx context.Context, userID string) ([]repository.Achievement, error) {
	if s.achievements == nil {
		return nil, nil
	}
	return s.achievements.ListUnlocked(ctx, userID)
}
