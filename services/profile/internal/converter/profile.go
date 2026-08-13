package converter

import (
	"github.com/Eastwesser/event-horizon/services/profile/internal/model"
	pb "github.com/Eastwesser/event-horizon/services/profile/proto"
)

func ProfileToProto(p *model.Profile) *pb.GetProfileResponse {
	if p == nil {
		return nil
	}
	return &pb.GetProfileResponse{
		UserId:     p.UserID,
		Email:      p.Email,
		Nickname:   p.Nickname,
		TotalScore: p.TotalScore,
		BestScores: p.BestScores,
		Lamps:      p.Lamps,
		Tickets:    p.Tickets,
	}
}

func ProfileFromUpdateRequest(req *pb.UpdateProfileRequest) *model.Profile {
	if req == nil {
		return nil
	}
	p := &model.Profile{
		UserID:     req.UserId,
		Nickname:   req.Nickname,
		BestScores: req.BestScores,
	}
	if req.TotalScore != nil {
		p.TotalScore = *req.TotalScore
	}
	return p
}
