package handler

import (
    "context"

    "google.golang.org/grpc/codes"
    "google.golang.org/grpc/status"

    pb "github.com/Eastwesser/event-horizon/services/profile/proto"
    "github.com/Eastwesser/event-horizon/services/profile/internal/service"
	"github.com/Eastwesser/event-horizon/services/profile/internal/repository"
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

    return &pb.GetProfileResponse{
        UserId:     profile.UserID,
        Email:      profile.Email,
        Nickname:   profile.Nickname,
        TotalScore: profile.TotalScore,
        BestScores: profile.BestScores,
        Lamps:      profile.Lamps,
        Tickets:    profile.Tickets,
    }, nil
}

func (h *ProfileHandler) UpdateProfile(ctx context.Context, req *pb.UpdateProfileRequest) (*pb.UpdateProfileResponse, error) {
    profile := &repository.UserProfile{
        UserID:     req.UserId,
        Nickname:   req.Nickname,
        BestScores: req.BestScores,
    }

    // Проверяем, что TotalScore передан
    if req.TotalScore != nil {
        profile.TotalScore = *req.TotalScore
    }

    if err := h.profileService.UpdateProfile(ctx, profile); err != nil {
        return &pb.UpdateProfileResponse{Success: false, Message: err.Error()}, nil
    }

    return &pb.UpdateProfileResponse{Success: true, Message: "profile updated"}, nil
}
