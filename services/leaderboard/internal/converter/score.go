package converter

import (
	"github.com/Eastwesser/event-horizon/services/leaderboard/internal/model"
	pb "github.com/Eastwesser/event-horizon/services/leaderboard/proto"
)

func ScoreEntryToProto(e model.ScoreEntry) *pb.ScoreEntry {
	return &pb.ScoreEntry{
		Rank:      e.Rank,
		UserId:    e.UserID,
		Nickname:  e.Nickname,
		Score:     e.Score,
		UpdatedAt: e.UpdatedAt,
	}
}

func ScoreEntriesToProto(entries []model.ScoreEntry) []*pb.ScoreEntry {
	out := make([]*pb.ScoreEntry, len(entries))
	for i, e := range entries {
		out[i] = ScoreEntryToProto(e)
	}
	return out
}
