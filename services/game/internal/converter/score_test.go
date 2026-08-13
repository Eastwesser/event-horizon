package converter

import (
	"testing"

	pb "github.com/Eastwesser/event-horizon/services/game/proto"
)

func TestSubmitScoreFromProto(t *testing.T) {
	m := SubmitScoreFromProto(&pb.SubmitScoreRequest{UserId: "u1", GameId: "hexagon", Level: 2, Score: 100})
	if m.UserID != "u1" || m.GameID != "hexagon" || m.Score != 100 {
		t.Fatalf("%+v", m)
	}
}
