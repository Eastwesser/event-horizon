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
)

type SubmitScoreRequest struct {
    UserID string
    GameID string
    Level  int
    Score  int
    UserEmail string  // 👈 добавляем
    Seed   string
    Moves  []hexagonValidator.Move
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

// func (s *gameService) SubmitScore(ctx context.Context, req *SubmitScoreRequest) (*SubmitScoreResponse, error) {
//     log.Printf("📥 Service received score: %d", req.Score)
    
//     validatedScore := req.Score  // временно

//     // 1. Валидация игры — сервер сам вычисляет счёт (техдолг)
//     // valid, validatedScore, err := s.validator.ValidateMoves(req.Seed, req.Moves, 0) // finalScore не нужен
//     // if err != nil {
//     //     return nil, fmt.Errorf("validation error: %w", err)
//     // }
//     // if !valid {
//     //     return &SubmitScoreResponse{
//     //         Success: false,
//     //         Message: "invalid game state or moves",
//     //     }, nil
//     // }

//     // 2. Проверка, что игрок не играл сегодня (через Redis)
//     // TODO: реализовать проверку daily limit

//     // 3. Получаем текущий рекорд игрока
//     currentHighscore, err := s.repo.GetHighscore(ctx, req.UserID, req.GameID)
//     if err != nil {
//         log.Printf("Failed to get highscore: %v", err)
//     }

//     isNewRecord := validatedScore > currentHighscore

//     // 4. Сохраняем новый рекорд, если нужно
//     if isNewRecord {
//         err = s.repo.SaveHighscore(ctx, req.UserID, req.GameID, validatedScore)
//         if err != nil {
//             log.Printf("Failed to save highscore: %v", err)
//         }
//     }

//     // 5. Публикуем событие в NATS (всегда, даже если не рекорд?)
//     event := map[string]interface{}{
//         "user_id":    req.UserID,
//         "game_id":    req.GameID,
//         "user_email": req.UserEmail,
//         "score":      validatedScore,
//         "is_record":  isNewRecord,
//         "level":      req.Level,
//         "lamps_earned":  lampsEarned,   // 👈 добавить
//         "tickets_earned": ticketsEarned, // 👈 добавить
//         "timestamp":  time.Now().Unix(),
//     }
//     eventData, _ := json.Marshal(event)
    
//     _, err = s.js.Publish("score.updated", eventData)
//     if err != nil {
//         log.Printf("Failed to publish to NATS: %v", err)
//     } else {
//         log.Printf("📡 Published score.updated: user=%s, game=%s, score=%d, is_record=%v",
//             req.UserID, req.GameID, validatedScore, isNewRecord)
//     }

//     // 6. Получаем ранг из leaderboard (опционально, через gRPC)
//     rank := 0
//     // TODO: вызвать leaderboard.GetPlayerRank

//     // 7. Начисляем награды
//     lampsEarned := 10  // базовая награда за игру
//     ticketsEarned := 0
//     if isNewRecord {
//         ticketsEarned = validatedScore / 100
//         if ticketsEarned > 100 {
//             ticketsEarned = 100
//         }
//     }

//     return &SubmitScoreResponse{
//         Success:       true,
//         NewHighscore:  validatedScore,
//         Rank:          rank,
//         Message:       "score submitted successfully",
//         LampsEarned:   lampsEarned,
//         TicketsEarned: ticketsEarned,
//     }, nil
// }

func (s *gameService) SubmitScore(ctx context.Context, req *SubmitScoreRequest) (*SubmitScoreResponse, error) {
    // Включаем валидацию
    valid, validatedScore, err := s.validator.ValidateMoves(req.Seed, req.Moves, 0)
    if err != nil {
        return nil, fmt.Errorf("validation error: %w", err)
    }
    if !valid {
        return &SubmitScoreResponse{
            Success: false,
            Message: "invalid game state or moves",
        }, nil
    }    

    log.Printf("📥 Service received score: %d", req.Score)
    
    validatedScore := req.Score

    // Получаем текущий рекорд
    currentHighscore, err := s.repo.GetHighscore(ctx, req.UserID, req.GameID)
    if err != nil {
        log.Printf("Failed to get highscore: %v", err)
    }

    isNewRecord := validatedScore > currentHighscore

    // Сохраняем рекорд
    if isNewRecord {
        err = s.repo.SaveHighscore(ctx, req.UserID, req.GameID, validatedScore)
        if err != nil {
            log.Printf("Failed to save highscore: %v", err)
        }
    }

    // 👇 НАЧИСЛЯЕМ НАГРАДЫ (до публикации)
    lampsEarned := 10
    ticketsEarned := 0
    if isNewRecord {
        ticketsEarned = validatedScore / 100
        if ticketsEarned > 100 {
            ticketsEarned = 100
        }
    }

    log.Printf("🎯 isNewRecord=%v, validatedScore=%d, lamps=%d, tickets=%d", 
    isNewRecord, validatedScore, lampsEarned, ticketsEarned)

    // 👇 ПУБЛИКУЕМ СОБЫТИЕ (после того как награды вычислены)
    event := map[string]interface{}{
        "user_id":         req.UserID,
        "game_id":         req.GameID,
        "user_email":      req.UserEmail,
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

    // Ответ
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
            Name:        "Гексагональный пазл",
            Description: "Складывай плитки с едой, чтобы они исчезали и приносили очки",
            Levels: []LevelInfo{
                {Level: 1, TargetScore: 100, RewardLamps: 10, RewardTickets: 0},
                {Level: 2, TargetScore: 200, RewardLamps: 15, RewardTickets: 0},
                {Level: 3, TargetScore: 350, RewardLamps: 20, RewardTickets: 5},
                {Level: 4, TargetScore: 550, RewardLamps: 30, RewardTickets: 10},
                {Level: 5, TargetScore: 800, RewardLamps: 50, RewardTickets: 20},
            },
        }, nil
    case "flappy":
        return &GameInfo{
            GameID:      "flappy",
            Name:        "Flappy Bird",
            Description: "Трубы, птичка, полёт",
            Levels: []LevelInfo{
                {Level: 1, TargetScore: 10, RewardLamps: 10, RewardTickets: 0},
                {Level: 2, TargetScore: 20, RewardLamps: 15, RewardTickets: 5},
                {Level: 3, TargetScore: 35, RewardLamps: 20, RewardTickets: 10},
            },
        }, nil
    default:
        return nil, fmt.Errorf("game not found: %s", gameID)
    }
}
