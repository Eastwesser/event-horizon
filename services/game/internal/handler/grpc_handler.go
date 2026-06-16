package handler

import (
    "context"
    "log"

    "google.golang.org/grpc/codes"
    "google.golang.org/grpc/status"
    "github.com/prometheus/client_golang/prometheus"
    "github.com/prometheus/client_golang/prometheus/promauto"

    pb "event_horizon/services/game/proto"
    "event_horizon/services/game/internal/service"
    hexagonValidator "event_horizon/services/game/games/hexagons"
)

var (
    gameSubmitsTotal = promauto.NewCounterVec(
        prometheus.CounterOpts{
            Name: "game_submits_total",
            Help: "Total number of score submissions",
        },
        []string{"game_id", "status"},
    )

    gameScoreHistogram = promauto.NewHistogramVec(
        prometheus.HistogramOpts{
            Name:    "game_score_histogram",
            Help:    "Distribution of scores by game",
            Buckets: []float64{10, 50, 100, 200, 500, 1000, 2000, 5000},
        },
        []string{"game_id"},
    )
)

type GameHandler struct {
    pb.UnimplementedGameServiceServer
    gameService service.GameService
}

func NewGameHandler(svc service.GameService) *GameHandler {
    return &GameHandler{
        gameService: svc,
    }
}

func (h *GameHandler) SubmitScore(ctx context.Context, req *pb.SubmitScoreRequest) (*pb.SubmitScoreResponse, error) {
    log.Printf("📥 HANDLER RAW: %+v", req)
    
    if req.UserId == "" || req.GameId == "" {
        return nil, status.Error(codes.InvalidArgument, "user_id and game_id are required")
    }
    
    log.Printf("📥 HANDLER: user=%s, game=%s, level=%d, score=%d, moves=%d, email=%s", 
        req.UserId, req.GameId, req.Level, req.Score, len(req.Moves), req.UserEmail)

    // Конвертируем moves из protobuf в hexagonValidator.Move
    moves := make([]hexagonValidator.Move, len(req.Moves))
    for i, m := range req.Moves {
        moves[i] = hexagonValidator.Move{
            FromX:     int(m.FromX),
            FromY:     int(m.FromY),
            ToX:       int(m.ToX),
            ToY:       int(m.ToY),
            Timestamp: m.Timestamp,
        }
    }

    resp, err := h.gameService.SubmitScore(ctx, &service.SubmitScoreRequest{
        UserID:    req.UserId,
        GameID:    req.GameId,
        Level:     int(req.Level),
        Score:     int(req.Score),
        UserEmail: req.UserEmail,
        Seed:      req.Seed,
        Moves:     moves,
    })
    if err != nil {
        log.Printf("SubmitScore error: %v", err)
        return &pb.SubmitScoreResponse{
            Success: false,
            Message: err.Error(),
        }, nil
    }

    // Метрики после успешного ответа
    if resp.Success {
        gameSubmitsTotal.WithLabelValues(req.GameId, "success").Inc()
        gameScoreHistogram.WithLabelValues(req.GameId).Observe(float64(resp.NewHighscore))
    } else {
        gameSubmitsTotal.WithLabelValues(req.GameId, "error").Inc()
    }

    return &pb.SubmitScoreResponse{
        Success:        resp.Success,
        NewHighscore:   int32(resp.NewHighscore),
        Rank:           int32(resp.Rank),
        Message:        resp.Message,
        LampsEarned:    int32(resp.LampsEarned),
        TicketsEarned:  int32(resp.TicketsEarned),
    }, nil
}

func (h *GameHandler) GetGameInfo(ctx context.Context, req *pb.GetGameInfoRequest) (*pb.GetGameInfoResponse, error) {
    if req.GameId == "" {
        return nil, status.Error(codes.InvalidArgument, "game_id is required")
    }

    info, err := h.gameService.GetGameInfo(ctx, req.GameId)
    if err != nil {
        return nil, status.Error(codes.NotFound, err.Error())
    }

    levels := make([]*pb.LevelInfo, len(info.Levels))
    for i, l := range info.Levels {
        levels[i] = &pb.LevelInfo{
            Level:         int32(l.Level),
            TargetScore:   int32(l.TargetScore),
            RewardLamps:   int32(l.RewardLamps),
            RewardTickets: int32(l.RewardTickets),
        }
    }

    return &pb.GetGameInfoResponse{
        GameId:      info.GameID,
        Name:        info.Name,
        Description: info.Description,
        Levels:      levels,
    }, nil
}
