package converter

import (
	"testing"
	"time"

	"github.com/Eastwesser/event-horizon/services/auth/internal/model"
)

func TestUserToProtoWithScores(t *testing.T) {
	u := &model.User{ID: "u1", Email: "a@b.c", Nickname: "neo", Role: "author", CreatedAt: time.Now()}
	resp := UserToProtoWithScores(u, map[string]int32{"hexagon": 10}, 10)
	if resp == nil || resp.UserId != "u1" || resp.Role != "author" || resp.BestScores["hexagon"] != 10 {
		t.Fatalf("unexpected: %+v", resp)
	}
}
