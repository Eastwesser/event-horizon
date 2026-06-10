package service

import (
    "context"
    "encoding/json"
    "fmt"
    "log"
    "time"

    "github.com/nats-io/nats.go"

    "event_horizon/services/game/internal/repository"
    hexagonValidator "event_horizon/services/game/games/hexagons"
    "event_horizon/services/game/games/memory"
)

type SubmitScoreRequest struct {
    UserID    string
    GameID    string
    Level     int
    Score     int
    UserEmail string
    Nickname  string
    Seed      string
    Moves     []hexagonValidator.Move
}

type SubmitScoreResponse struct {
    Success       bool
    NewHighscore  int
    Rank          int
    Message       string
    LampsEarned   int
    TicketsEarned int
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
}

type gameService struct {
    repo      repository.GameRepository
    js        nats.JetStreamContext
    validator *hexagonValidator.Validator
}

func NewGameService(repo repository.GameRepository, js nats.JetStreamContext) GameService {
    return &gameService{
        repo:      repo,
        js:        js,
        validator: hexagonValidator.NewValidator(),
    }
}

func (s *gameService) SubmitScore(ctx context.Context, req *SubmitScoreRequest) (*SubmitScoreResponse, error) {
    var lampsEarned, ticketsEarned int
    validatedScore := req.Score

    log.Printf("📥 Service received: game_id=%s, user=%s, score=%d", req.GameID, req.UserID, req.Score)

    // В зависимости от игры — своя логика наград
    switch req.GameID {
    case "hexagon":
        validatedScore = req.Score
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
        validatedScore = req.Score
        lampsEarned = 5
        ticketsEarned = req.Score / 10
        if ticketsEarned > 100 {
            ticketsEarned = 100
        }
    case "towers":
        validatedScore = req.Score
        lampsEarned = 5
        ticketsEarned = req.Score / 20
        if ticketsEarned > 50 {
            ticketsEarned = 50
        }    
    default:
        return &SubmitScoreResponse{
            Success: false,
            Message: fmt.Sprintf("unknown game_id: %s", req.GameID),
        }, nil
    }

    log.Printf("📥 Validated score: %d", validatedScore)

    // Получаем текущий рекорд
    currentHighscore, err := s.repo.GetHighscore(ctx, req.UserID, req.GameID)
    if err != nil {
        log.Printf("Failed to get highscore: %v", err)
    }

    isNewRecord := validatedScore > currentHighscore

    // Сохраняем рекорд если это новый рекорд
    if isNewRecord {
        err = s.repo.SaveHighscore(ctx, req.UserID, req.GameID, validatedScore)
        if err != nil {
            log.Printf("Failed to save highscore: %v", err)
        }
    }

    log.Printf("🎯 isNewRecord=%v, validatedScore=%d, lamps=%d, tickets=%d",
        isNewRecord, validatedScore, lampsEarned, ticketsEarned)

    // Публикуем событие в NATS
    event := map[string]interface{}{
        "user_id":         req.UserID,
        "game_id":         req.GameID,
        "user_email":      req.UserEmail,
        "nickname":        req.Nickname,
        "score":           validatedScore,
        "is_record":       isNewRecord,
        "level":           req.Level,
        "lamps_earned":    lampsEarned,
        "tickets_earned":  ticketsEarned,
        "timestamp":       time.Now().Unix(),
    }
    eventData, _ := json.Marshal(event)

    _, err = s.js.Publish("score.updated", eventData)
    if err != nil {
        log.Printf("Failed to publish to NATS: %v", err)
    } else {
        log.Printf("📡 Published score.updated: user=%s, game=%s, score=%d, is_record=%v, lamps=%d, tickets=%d",
            req.UserID, req.GameID, validatedScore, isNewRecord, lampsEarned, ticketsEarned)
    }

    return &SubmitScoreResponse{
        Success:       true,
        NewHighscore:  validatedScore,
        Rank:          0,
        Message:       "score submitted successfully",
        LampsEarned:   lampsEarned,
        TicketsEarned: ticketsEarned,
    }, nil
}

func (s *gameService) GetGameInfo(ctx context.Context, gameID string) (*GameInfo, error) {
    switch gameID {
    case "hexagon":
        return &GameInfo{
            GameID:      "hexagon",
            Name:        "Никуся — Блинопёк",
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
        return &GameInfo{
            GameID:      "flappy",
            Name:        "Flappy Bird",
            Description: "Трубы, птичка, полёт",
            Levels: []LevelInfo{
                {Level: 1, TargetScore: 10, RewardLamps: 5, RewardTickets: 0},
                {Level: 2, TargetScore: 20, RewardLamps: 10, RewardTickets: 5},
                {Level: 3, TargetScore: 35, RewardLamps: 15, RewardTickets: 10},
            },
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
    default:
        return nil, fmt.Errorf("game not found: %s", gameID)
    }
}
