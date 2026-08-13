package leaderboard

import (
	"fmt"
	"strings"
)

func (m *GetTopScoresRequest) Validate() error {
	if m == nil || strings.TrimSpace(m.GameId) == "" {
		return fmt.Errorf("game_id is required")
	}
	if m.Limit < 0 || m.Limit > 100 {
		return fmt.Errorf("limit must be 0-100")
	}
	return nil
}

func (m *GetPlayerRankRequest) Validate() error {
	if m == nil || strings.TrimSpace(m.GameId) == "" {
		return fmt.Errorf("game_id is required")
	}
	if strings.TrimSpace(m.UserId) == "" {
		return fmt.Errorf("user_id is required")
	}
	return nil
}

func (m *UpdateScoreRequest) Validate() error {
	if m == nil || strings.TrimSpace(m.GameId) == "" {
		return fmt.Errorf("game_id is required")
	}
	if strings.TrimSpace(m.UserId) == "" {
		return fmt.Errorf("user_id is required")
	}
	if m.Score < 0 || m.Score > 100000 {
		return fmt.Errorf("score must be 0-100000")
	}
	return nil
}
