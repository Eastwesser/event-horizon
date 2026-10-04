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

// EvaluateAchievements derives unlock codes from best_scores / total_score / optional event level.
// totalScore should be the sum of best_scores (already maintained by the score.updated consumer).
func EvaluateAchievements(bestScores map[string]int32, totalScore int32, gameID string, level int) []string {
	codes := make([]string, 0, 16)
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
	n, err := s.achievements.Unlock(ctx, userID, codes)
	if err != nil {
		return fmt.Errorf("unlock achievements: %w", err)
	}
	if n > 0 {
		log.Printf("achievements unlocked user=%s count=%d", userID, n)
	}
	return nil
}

func (s *profileService) ListAchievements(ctx context.Context, userID string) ([]repository.Achievement, error) {
	if s.achievements == nil {
		return nil, nil
	}
	return s.achievements.ListUnlocked(ctx, userID)
}
