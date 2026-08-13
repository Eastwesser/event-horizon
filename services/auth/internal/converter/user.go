package converter

import (
	pb "github.com/Eastwesser/event-horizon/services/auth/proto"
	"github.com/Eastwesser/event-horizon/services/auth/internal/model"
)

// UserToProto maps a domain user to GetUserResponse fields (without scores).
func UserToProto(u *model.User) *pb.GetUserResponse {
	if u == nil {
		return nil
	}
	return &pb.GetUserResponse{
		UserId:   u.ID,
		Email:    u.Email,
		Nickname: u.Nickname,
		Role:     u.Role,
	}
}

// UserToProtoWithScores attaches best scores to the response.
func UserToProtoWithScores(u *model.User, bestScores map[string]int32, totalScore int32) *pb.GetUserResponse {
	resp := UserToProto(u)
	if resp == nil {
		return nil
	}
	resp.BestScores = bestScores
	resp.TotalScore = totalScore
	return resp
}
