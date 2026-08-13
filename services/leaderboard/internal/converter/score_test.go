package converter

import (
	"testing"

	"github.com/Eastwesser/event-horizon/services/leaderboard/internal/model"
)

func TestScoreEntryToProto(t *testing.T) {
	p := ScoreEntryToProto(model.ScoreEntry{Rank: 1, UserID: "u", Nickname: "n", Score: 9})
	if p.Rank != 1 || p.UserId != "u" {
		t.Fatalf("%+v", p)
	}
}
