package handler

import (
	"context"
	"time"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"github.com/Eastwesser/event-horizon/services/profile/internal/repository"
	"github.com/Eastwesser/event-horizon/services/profile/internal/service"
	pb "github.com/Eastwesser/event-horizon/services/profile/proto"
)

type ProfileHandler struct {
	pb.UnimplementedProfileServiceServer
	profileService service.ProfileService
}

func NewProfileHandler(profileService service.ProfileService) *ProfileHandler {
	return &ProfileHandler{
		profileService: profileService,
	}
}

func (h *ProfileHandler) GetProfile(ctx context.Context, req *pb.GetProfileRequest) (*pb.GetProfileResponse, error) {
	if req.UserId == "" {
		return nil, status.Error(codes.InvalidArgument, "user_id required")
	}

	profile, err := h.profileService.GetProfile(ctx, req.UserId)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	if profile == nil {
		return nil, status.Error(codes.NotFound, "profile not found")
	}

	// Backfill score-based achievements from stored best_scores / total_score
	// (sum of best_scores). Level achievements need a live score.updated event.
	_ = h.profileService.EvaluateAndUnlock(ctx, profile.UserID, profile.BestScores, profile.TotalScore, "", 0)

	achievements, err := h.profileService.ListAchievements(ctx, profile.UserID)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	pbAch := make([]*pb.Achievement, 0, len(achievements))
	for _, a := range achievements {
		pbAch = append(pbAch, &pb.Achievement{
			Code:        a.Code,
			Title:       a.Title,
			Description: a.Description,
			Icon:        a.Icon,
			UnlockedAt:  a.UnlockedAt.UTC().Format(time.RFC3339),
		})
	}

	return &pb.GetProfileResponse{
		UserId:       profile.UserID,
		Email:        profile.Email,
		Nickname:     profile.Nickname,
		TotalScore:   profile.TotalScore,
		BestScores:   profile.BestScores,
		Lamps:        profile.Lamps,
		Tickets:      profile.Tickets,
		Achievements: pbAch,
	}, nil
}

func (h *ProfileHandler) UpdateProfile(ctx context.Context, req *pb.UpdateProfileRequest) (*pb.UpdateProfileResponse, error) {
	profile := &repository.UserProfile{
		UserID:     req.UserId,
		Nickname:   req.Nickname,
		BestScores: req.BestScores,
	}

	if req.TotalScore != nil {
		profile.TotalScore = *req.TotalScore
	}

	if err := h.profileService.UpdateProfile(ctx, profile); err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	return &pb.UpdateProfileResponse{Success: true, Message: "profile updated"}, nil
}
