package service

import (
    "context"
    "encoding/json"
    "fmt"
    "log"
    "strings"
    "time"

    "github.com/google/uuid"
    "github.com/nats-io/nats.go"
    "google.golang.org/grpc/codes"
    "google.golang.org/grpc/status"

    billingPb "github.com/Eastwesser/event-horizon/services/billing/proto"
    "github.com/Eastwesser/event-horizon/services/game/internal/repository"
    hexagonValidator "github.com/Eastwesser/event-horizon/services/game/games/hexagons"
    "github.com/Eastwesser/event-horizon/services/game/games/memory"
)

const (
    boostCostLamps = 10
)

// boostAllowedGames — games that can arm a lamp boost (boosted runs are not ranked).
var boostAllowedGames = map[string]struct{}{
    "hexagon":   {},
    "flappy":    {},
    "memory":    {},
    "towers":    {},
    "hanoi":     {},
    "twenty48":  {},
    "gears":     {},
    "companion": {},
}

type SubmitScoreRequest struct {
    UserID    string
    GameID    string
    Level     int
    Score     int
    UserEmail string
    Nickname  string
    Seed      string
    Moves     []hexagonValidator.Move
    BoostID   string
}

type SubmitScoreResponse struct {
    Success       bool
    NewHighscore  int
    Rank          int
    Message       string
    LampsEarned   int
    TicketsEarned int
    Ranked        bool
}

type StartBoostRequest struct {
    UserID string
    GameID string
}

type StartBoostResponse struct {
    BoostID    string
    Boosted    bool
    Cost       int
    NewBalance int
    Message    string
}

type GameInfo struct {
    GameID      string
    Name        string
    Description string
    Levels      []LevelInfo
}

type LevelInfo struct {
    Level         int
    TargetScore   int
    RewardLamps   int
    RewardTickets int
}

type GameService interface {
    SubmitScore(ctx context.Context, req *SubmitScoreRequest) (*SubmitScoreResponse, error)
    GetGameInfo(ctx context.Context, gameID string) (*GameInfo, error)
    StartBoost(ctx context.Context, req *StartBoostRequest) (*StartBoostResponse, error)
}

type gameService struct {
    repo          repository.GameRepository
    js            nats.JetStreamContext
    validator     *hexagonValidator.Validator
    billingClient billingPb.BillingServiceClient
}

func NewGameService(repo repository.GameRepository, js nats.JetStreamContext, billingClient billingPb.BillingServiceClient) GameService {
    return &gameService{
        repo:          repo,
        js:            js,
        validator:     hexagonValidator.NewValidator(),
        billingClient: billingClient,
    }
}

func (s *gameService) SubmitScore(ctx context.Context, req *SubmitScoreRequest) (*SubmitScoreResponse, error) {
    var lampsEarned, ticketsEarned int
    validatedScore := req.Score

    log.Printf("📥 Service received: game_id=%s, user=%s, score=%d", req.GameID, req.UserID, req.Score)

    // В зависимости от игры — своя логика наград
    switch req.GameID {
    case "hexagon":
        valid, valScore, err := s.validator.ValidateMoves(req.Seed, req.Moves, req.Score)
        if err != nil {
            log.Printf("❌ Hexagon validation error: %v", err)
            return &SubmitScoreResponse{Success: false, Message: "validation error"}, nil
        }
        if !valid {
            log.Printf("⚠️ Invalid hexagon game state for user %s", req.UserID)
            return &SubmitScoreResponse{Success: false, Message: "invalid game state or moves"}, nil
        }
        validatedScore = valScore
        lampsEarned = 10
        ticketsEarned = 0
        if validatedScore > 0 {
            ticketsEarned = validatedScore / 100
            if ticketsEarned > 100 {
                ticketsEarned = 100
            }
        }

    case "memory":
        game := memory.NewMemoryGame(req.Seed)
        var memoryMoves []memory.Move
        for _, m := range req.Moves {
            memoryMoves = append(memoryMoves, memory.Move{
                CardIndex1: int(m.FromX),
                CardIndex2: int(m.ToX),
            })
        }

        valid, valScore, err := game.ValidateMoves(req.Seed, memoryMoves, req.Score)
        if err != nil {
            log.Printf("❌ Memory validation error: %v", err)
            return &SubmitScoreResponse{
                Success: false,
                Message: "validation error",
            }, nil
        }
        if !valid {
            log.Printf("⚠️ Invalid memory game state for user %s", req.UserID)
            return &SubmitScoreResponse{
                Success: false,
                Message: "invalid game state or moves",
            }, nil
        }

        validatedScore = valScore
        lampsEarned, ticketsEarned = game.CalculateRewards(validatedScore)

    case "flappy":
        if req.Score < 0 || req.Score > 10000 {
            return &SubmitScoreResponse{Success: false, Message: "score out of allowed range"}, nil
        }
        if req.Level < 1 || req.Level > 10 {
            return &SubmitScoreResponse{Success: false, Message: "flappy level must be 1-10"}, nil
        }
        validatedScore = req.Score
        lampsEarned = 5 + (req.Level - 1)
        ticketsEarned = req.Score/10 + (req.Level-1)*5
        if ticketsEarned > 100 {
            ticketsEarned = 100
        }
    case "towers":
        if req.Score < 0 || req.Score > 10000 {
            return &SubmitScoreResponse{Success: false, Message: "score out of allowed range"}, nil
        }
        validatedScore = req.Score
        lampsEarned = 5
        ticketsEarned = req.Score / 20
        if ticketsEarned > 50 {
            ticketsEarned = 50
        }

    case "hanoi":
        // Score is computed client-side as 1000 - (excess moves × 20), floored at 100
        // (same formula as Memory), so the valid range is always [100, 1000].
        if req.Score < 100 || req.Score > 1000 {
            return &SubmitScoreResponse{Success: false, Message: "score out of allowed range"}, nil
        }
        validatedScore = req.Score
        lampsEarned = 5
        ticketsEarned = 5
        if validatedScore > 500 {
            ticketsEarned += (validatedScore - 500) / 100
        }
        if ticketsEarned > 20 {
            ticketsEarned = 20
        }
        if validatedScore >= 900 {
            lampsEarned += 5
        }

    case "twenty48":
        if req.Score < 0 || req.Score > 100000 {
            return &SubmitScoreResponse{Success: false, Message: "score out of allowed range"}, nil
        }
        validatedScore = req.Score
        lampsEarned = 5
        ticketsEarned = req.Score / 200
        if ticketsEarned > 80 {
            ticketsEarned = 80
        }

    case "gears":
        if req.Score < 0 || req.Score > 50000 {
            return &SubmitScoreResponse{Success: false, Message: "score out of allowed range"}, nil
        }
        validatedScore = req.Score
        lampsEarned = 5
        ticketsEarned = req.Score / 50
        if ticketsEarned > 60 {
            ticketsEarned = 60
        }

    case "companion":
        if req.Score < 0 || req.Score > 10000 {
            return &SubmitScoreResponse{Success: false, Message: "score out of allowed range"}, nil
        }
        validatedScore = req.Score
        lampsEarned = 3
        ticketsEarned = req.Score / 100
        if ticketsEarned > 30 {
            ticketsEarned = 30
        }

    default:
        return &SubmitScoreResponse{
            Success: false,
            Message: fmt.Sprintf("unknown game_id: %s", req.GameID),
        }, nil
    }

    log.Printf("📥 Validated score: %d", validatedScore)

    // Kids-safe: any unconsumed boost for this user+game skips leaderboard + rewards.
    boostID := strings.TrimSpace(req.BoostID)
    if boostID == "" {
        if activeID, ok, lookupErr := s.repo.GetActiveRunBoost(ctx, req.UserID, req.GameID); lookupErr != nil {
            log.Printf("lookup boost failed: %v", lookupErr)
        } else if ok {
            boostID = activeID
        }
    }
    consumedID, boosted, err := s.repo.ConsumeActiveRunBoost(ctx, req.UserID, req.GameID, boostID)
    if err != nil {
        log.Printf("consume boost failed: %v", err)
    }
    // If client sent a stale boost_id, still block on any remaining active boost.
    if !boosted {
        if activeID, ok, lookupErr := s.repo.GetActiveRunBoost(ctx, req.UserID, req.GameID); lookupErr == nil && ok {
            if cid, cok, cerr := s.repo.ConsumeActiveRunBoost(ctx, req.UserID, req.GameID, activeID); cerr == nil && cok {
                consumedID, boosted = cid, true
            }
        }
    }
    if boosted {
        log.Printf("🚫 Boosted run not ranked: user=%s game=%s boost=%s score=%d",
            req.UserID, req.GameID, consumedID, validatedScore)
        return &SubmitScoreResponse{
            Success:       true,
            NewHighscore:  validatedScore,
            Rank:          0,
            Message:       "boosted run — not ranked",
            LampsEarned:   0,
            TicketsEarned: 0,
            Ranked:        false,
        }, nil
    }

    level := req.Level
    if level < 1 {
        level = 1
    }

    // Получаем текущий рекорд (per-level)
    currentHighscore, err := s.repo.GetHighscore(ctx, req.UserID, req.GameID, level)
    if err != nil {
        log.Printf("Failed to get highscore: %v", err)
    }

    isNewRecord := validatedScore > currentHighscore

    log.Printf("🎯 isNewRecord=%v, validatedScore=%d, level=%d, lamps=%d, tickets=%d",
        isNewRecord, validatedScore, level, lampsEarned, ticketsEarned)

    // Event for Leaderboard / Billing (same payload shape as before).
    event := map[string]interface{}{
        "user_id":         req.UserID,
        "game_id":         req.GameID,
        "user_email":      req.UserEmail,
        "nickname":        req.Nickname,
        "score":           validatedScore,
        "is_record":       isNewRecord,
        "level":           level,
        "lamps_earned":    lampsEarned,
        "tickets_earned":  ticketsEarned,
        "timestamp":       time.Now().Unix(),
    }
    eventData, _ := json.Marshal(event)

    // Prefer transactional outbox (reliable). Fall back to legacy direct NATS publish
    // if outbox insert fails (e.g. migration not yet applied).
    if isNewRecord {
        if err := s.repo.SaveHighscoreAndEnqueueOutbox(ctx, req.UserID, req.GameID, validatedScore, level, "score.updated", eventData); err != nil {
            log.Printf("outbox+highscore failed, falling back to SaveHighscore+Publish: %v", err)
            if saveErr := s.repo.SaveHighscore(ctx, req.UserID, req.GameID, validatedScore, level); saveErr != nil {
                log.Printf("Failed to save highscore: %v", saveErr)
            }
            s.publishScoreUpdatedDirect(eventData, req.UserID, req.GameID, validatedScore, isNewRecord, lampsEarned, ticketsEarned)
        } else {
            log.Printf("📬 Enqueued score.updated (with highscore): user=%s game=%s level=%d score=%d", req.UserID, req.GameID, level, validatedScore)
        }
    } else {
        if err := s.repo.EnqueueOutbox(ctx, "score.updated", eventData); err != nil {
            log.Printf("outbox enqueue failed, falling back to direct Publish: %v", err)
            s.publishScoreUpdatedDirect(eventData, req.UserID, req.GameID, validatedScore, isNewRecord, lampsEarned, ticketsEarned)
        } else {
            log.Printf("📬 Enqueued score.updated: user=%s game=%s level=%d score=%d", req.UserID, req.GameID, level, validatedScore)
        }
    }

    return &SubmitScoreResponse{
        Success:       true,
        NewHighscore:  validatedScore,
        Rank:          0,
        Message:       "score submitted successfully",
        LampsEarned:   lampsEarned,
        TicketsEarned: ticketsEarned,
        Ranked:        true,
    }, nil
}

func (s *gameService) StartBoost(ctx context.Context, req *StartBoostRequest) (*StartBoostResponse, error) {
    if req == nil {
        return nil, status.Error(codes.InvalidArgument, "request is required")
    }
    gameID := strings.TrimSpace(req.GameID)
    userID := strings.TrimSpace(req.UserID)
    if userID == "" || gameID == "" {
        return nil, status.Error(codes.InvalidArgument, "user_id and game_id are required")
    }
    if _, ok := boostAllowedGames[gameID]; !ok {
        return nil, status.Errorf(codes.FailedPrecondition, "boosts not available for game %s", gameID)
    }
    if s.billingClient == nil {
        return nil, status.Error(codes.Unavailable, "billing unavailable")
    }

    // Reuse an unpaid-for active boost (crash / reconnect) without double-charging.
    if existing, ok, err := s.repo.GetActiveRunBoost(ctx, userID, gameID); err != nil {
        return nil, status.Errorf(codes.Internal, "lookup boost: %v", err)
    } else if ok {
        return &StartBoostResponse{
            BoostID:    existing,
            Boosted:    true,
            Cost:       0,
            NewBalance: -1,
            Message:    "existing boost reused",
        }, nil
    }

    boostID := uuid.NewString()
    spendResp, err := s.billingClient.SpendCurrency(ctx, &billingPb.SpendCurrencyRequest{
        UserId:      userID,
        Currency:    billingPb.CurrencyType_LAMPS,
        Amount:      boostCostLamps,
        Reason:      "game_boost",
        ReferenceId: "boost:" + boostID,
    })
    if err != nil {
        if st, ok := status.FromError(err); ok && st.Code() == codes.FailedPrecondition {
            return nil, status.Error(codes.FailedPrecondition, st.Message())
        }
        if strings.Contains(err.Error(), "insufficient") {
            return nil, status.Error(codes.FailedPrecondition, err.Error())
        }
        return nil, status.Errorf(codes.Internal, "spend lamps: %v", err)
    }
    if spendResp != nil && !spendResp.GetSuccess() {
        msg := spendResp.GetMessage()
        if msg == "" {
            msg = "failed to spend lamps"
        }
        return nil, status.Error(codes.FailedPrecondition, msg)
    }

    if err := s.repo.CreateRunBoost(ctx, boostID, userID, gameID); err != nil {
        return nil, status.Errorf(codes.Internal, "record boost: %v", err)
    }

    newBalance := 0
    if spendResp != nil {
        newBalance = int(spendResp.GetNewBalance())
    }
    return &StartBoostResponse{
        BoostID:    boostID,
        Boosted:    true,
        Cost:       boostCostLamps,
        NewBalance: newBalance,
        Message:    "boost armed — this run will not be ranked",
    }, nil
}

// publishScoreUpdatedDirect is the legacy path (kept for fallback / tests).
func (s *gameService) publishScoreUpdatedDirect(eventData []byte, userID, gameID string, score int, isRecord bool, lamps, tickets int) {
    if s.js == nil {
        log.Printf("NATS JetStream nil; skip direct score.updated publish")
        return
    }
    _, err := s.js.Publish("score.updated", eventData)
    if err != nil {
        log.Printf("Failed to publish to NATS: %v", err)
        return
    }
    log.Printf("📡 Published score.updated (direct): user=%s, game=%s, score=%d, is_record=%v, lamps=%d, tickets=%d",
        userID, gameID, score, isRecord, lamps, tickets)
}

func (s *gameService) GetGameInfo(ctx context.Context, gameID string) (*GameInfo, error) {
    switch gameID {
    case "hexagon":
        return &GameInfo{
            GameID:      "hexagon",
            Name:        "Блинопёк",
            Description: "Гексагональный пазл с блинами",
            Levels: []LevelInfo{
                {Level: 1, TargetScore: 100, RewardLamps: 10, RewardTickets: 0},
                {Level: 2, TargetScore: 200, RewardLamps: 15, RewardTickets: 0},
                {Level: 3, TargetScore: 350, RewardLamps: 20, RewardTickets: 5},
                {Level: 4, TargetScore: 550, RewardLamps: 30, RewardTickets: 10},
                {Level: 5, TargetScore: 800, RewardLamps: 50, RewardTickets: 20},
            },
        }, nil
    case "memory":
        return &GameInfo{
            GameID:      "memory",
            Name:        "Меморина",
            Description: "Найди пары фруктов",
            Levels: []LevelInfo{
                {Level: 1, TargetScore: 500, RewardLamps: 5, RewardTickets: 5},
                {Level: 2, TargetScore: 800, RewardLamps: 10, RewardTickets: 10},
                {Level: 3, TargetScore: 1000, RewardLamps: 15, RewardTickets: 15},
            },
        }, nil
    case "flappy":
        flappyLevels := make([]LevelInfo, 0, 10)
        for lv := 1; lv <= 10; lv++ {
            flappyLevels = append(flappyLevels, LevelInfo{
                Level:         lv,
                TargetScore:   10 * lv,
                RewardLamps:   5 + (lv - 1),
                RewardTickets: (lv - 1) * 5,
            })
        }
        return &GameInfo{
            GameID:      "flappy",
            Name:        "Flappy Bird",
            Description: "Трубы, птичка, полёт",
            Levels:      flappyLevels,
        }, nil
    case "towers":
        return &GameInfo{
            GameID:      "towers",
            Name:        "Башенки",
            Description: "Строй башню из падающих блоков",
            Levels: []LevelInfo{
                {Level: 1, TargetScore: 100, RewardLamps: 5, RewardTickets: 0},
                {Level: 2, TargetScore: 200, RewardLamps: 10, RewardTickets: 5},
                {Level: 3, TargetScore: 350, RewardLamps: 15, RewardTickets: 10},
            },
        }, nil
    case "hanoi":
        return &GameInfo{
            GameID:      "hanoi",
            Name:        "Ханойская башня",
            Description: "Переместите все кольца на третий стержень за минимум ходов",
            Levels: []LevelInfo{
                {Level: 1, TargetScore: 500, RewardLamps: 5, RewardTickets: 5},
                {Level: 2, TargetScore: 800, RewardLamps: 8, RewardTickets: 10},
                {Level: 3, TargetScore: 1000, RewardLamps: 10, RewardTickets: 20},
            },
        }, nil
    case "twenty48":
        return &GameInfo{
            GameID:      "twenty48",
            Name:        "Горизонт 2048",
            Description: "Сдвигай плитки, собери 2048",
            Levels: []LevelInfo{
                {Level: 1, TargetScore: 500, RewardLamps: 5, RewardTickets: 5},
                {Level: 2, TargetScore: 2000, RewardLamps: 10, RewardTickets: 15},
                {Level: 3, TargetScore: 5000, RewardLamps: 15, RewardTickets: 30},
            },
        }, nil
    case "gears":
        return &GameInfo{
            GameID:      "gears",
            Name:        "Орбиты",
            Description: "Сливай шестерёнки до восьмой",
            Levels: []LevelInfo{
                {Level: 1, TargetScore: 80, RewardLamps: 5, RewardTickets: 5},
                {Level: 2, TargetScore: 200, RewardLamps: 10, RewardTickets: 15},
                {Level: 3, TargetScore: 400, RewardLamps: 15, RewardTickets: 25},
            },
        }, nil
    case "companion":
        return &GameInfo{
            GameID:      "companion",
            Name:        "Компаньон",
            Description: "Мягкий тамагочи без FOMO-смерти",
            Levels: []LevelInfo{
                {Level: 1, TargetScore: 50, RewardLamps: 3, RewardTickets: 3},
                {Level: 2, TargetScore: 150, RewardLamps: 5, RewardTickets: 8},
                {Level: 3, TargetScore: 400, RewardLamps: 8, RewardTickets: 15},
            },
        }, nil
    default:
        return nil, fmt.Errorf("game not found: %s", gameID)
    }
}
