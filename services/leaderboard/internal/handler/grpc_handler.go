package handler

import (
    "context"
    "log"

    "google.golang.org/grpc/codes"
    "google.golang.org/grpc/status"

    pb "event_horizon/services/leaderboard/proto"
    "event_horizon/services/leaderboard/internal/service"
)

type LeaderboardHandler struct {
    pb.UnimplementedLeaderboardServiceServer
    leaderboardService service.LeaderboardService
}

func NewLeaderboardHandler(svc service.LeaderboardService) *LeaderboardHandler {
    return &LeaderboardHandler{
        leaderboardService: svc,
    }
}

func (h *LeaderboardHandler) GetTopScores(ctx context.Context, req *pb.GetTopScoresRequest) (*pb.GetTopScoresResponse, error) {
    if req.GameId == "" {
        return nil, status.Error(codes.InvalidArgument, "game_id is required")
    }

    limit := int(req.Limit)
    if limit <= 0 || limit > 100 {
        limit = 10
    }

    entries, err := h.leaderboardService.GetTopScores(ctx, req.GameId, limit)
    if err != nil {
        log.Printf("Failed to get top scores: %v", err)
        return nil, status.Error(codes.Internal, err.Error())
    }

    pbEntries := make([]*pb.ScoreEntry, 0, len(entries))
    for _, e := range entries {
        pbEntries = append(pbEntries, &pb.ScoreEntry{
            Rank:      int32(e.Rank),
            UserId:    e.UserID,
            UserEmail: e.UserEmail,
            Score:     int32(e.Score),
            UpdatedAt: e.UpdatedAt,
        })
    }

    return &pb.GetTopScoresResponse{Entries: pbEntries}, nil
}

func (h *LeaderboardHandler) GetPlayerRank(ctx context.Context, req *pb.GetPlayerRankRequest) (*pb.GetPlayerRankResponse, error) {
    if req.GameId == "" || req.UserId == "" {
        return nil, status.Error(codes.InvalidArgument, "game_id and user_id are required")
    }

    rank, score, err := h.leaderboardService.GetPlayerRank(ctx, req.GameId, req.UserId)
    if err != nil {
        log.Printf("Failed to get player rank: %v", err)
        return nil, status.Error(codes.Internal, err.Error())
    }

    return &pb.GetPlayerRankResponse{
        Rank:  int32(rank),
        Score: int32(score),
    }, nil
}

func (h *LeaderboardHandler) UpdateScore(ctx context.Context, req *pb.UpdateScoreRequest) (*pb.UpdateScoreResponse, error) {
    if req.GameId == "" || req.UserId == "" {
        return nil, status.Error(codes.InvalidArgument, "game_id and user_id are required")
    }

    newRank, err := h.leaderboardService.UpdateScore(ctx, req.GameId, req.UserId, req.UserEmail, int(req.Score))
    if err != nil {
        log.Printf("Failed to update score: %v", err)
        return &pb.UpdateScoreResponse{
            Success: false,
            Message: err.Error(),
        }, nil
    }

    return &pb.UpdateScoreResponse{
        Success:  true,
        NewRank:  int32(newRank),
        Message:  "score updated successfully",
    }, nil
}
