package game

import (
	"fmt"
	"strings"
)

func (m *SubmitScoreRequest) Validate() error {
	if m == nil {
		return fmt.Errorf("request is nil")
	}
	if strings.TrimSpace(m.UserId) == "" {
		return fmt.Errorf("user_id is required")
	}
	if strings.TrimSpace(m.GameId) == "" {
		return fmt.Errorf("game_id is required")
	}
	if m.Level < 0 || m.Level > 100 {
		return fmt.Errorf("level must be 0-100")
	}
	if m.Score < 0 || m.Score > 100000 {
		return fmt.Errorf("score must be 0-100000")
	}
	if len(m.Moves) > 10000 {
		return fmt.Errorf("too many moves")
	}
	return nil
}

func (m *GetGameInfoRequest) Validate() error {
	if m == nil || strings.TrimSpace(m.GameId) == "" {
		return fmt.Errorf("game_id is required")
	}
	return nil
}
