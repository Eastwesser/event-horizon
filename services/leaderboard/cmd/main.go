package main

import (
    "context"
    "encoding/json"
    "log"
    "net"
    "os"
    "os/signal"
    "syscall"
    "time"

    "github.com/nats-io/nats.go"
    "google.golang.org/grpc"
    "google.golang.org/grpc/reflection"

    "event_horizon/services/leaderboard/internal/config"
    "event_horizon/services/leaderboard/internal/handler"
    "event_horizon/services/leaderboard/internal/repository"
    "event_horizon/services/leaderboard/internal/service"
    pb "event_horizon/services/leaderboard/proto"
)

// ScoreEvent структура для входящих сообщений из NATS
type ScoreEvent struct {
    UserID    string `json:"user_id"`
    GameID    string `json:"game_id"`
    UserEmail string `json:"user_email"`
    Score     int    `json:"score"`
}

func main() {
    cfg := config.Load()

    // Подключаемся к Redis
    redisRepo := repository.NewRedisLeaderboardRepo(cfg.RedisAddr, cfg.RedisDB)

    // Создаём сервис
    leaderboardService := service.NewLeaderboardService(redisRepo)

    // Подключаемся к NATS
    nc, err := nats.Connect(cfg.NATSUrl)
    if err != nil {
        log.Fatalf("Failed to connect to NATS: %v", err)
    }
    defer nc.Close()

    // Создаём JetStream контекст
    js, err := nc.JetStream()
    if err != nil {
        log.Fatalf("Failed to create JetStream context: %v", err)
    }

    // Создаём поток для score событий (если не существует)
    _, err = js.AddStream(&nats.StreamConfig{
        Name:     "SCORES",
        Subjects: []string{"score.updated"},
        Storage:  nats.FileStorage,
        MaxAge:   24 * time.Hour,
    })
    if err != nil {
        log.Printf("Stream might already exist: %v", err)
    }

    // Подписываемся на события score.updated
    _, err = js.Subscribe("score.updated", func(msg *nats.Msg) {
        var event ScoreEvent
        if err := json.Unmarshal(msg.Data, &event); err != nil {
            log.Printf("Failed to unmarshal score event: %v", err)
            return
        }

        log.Printf("📡 Received score update via NATS: game=%s user=%s score=%d",
            event.GameID, event.UserID, event.Score)

        // Обновляем leaderboard (используем context.Background для NATS сообщений)
        newRank, err := leaderboardService.UpdateScore(
            context.Background(),
            event.GameID,
            event.UserID,
            event.UserEmail,
            event.Score,
        )
        if err != nil {
            log.Printf("Failed to update score: %v", err)
            return
        }

        log.Printf("✅ Score updated for %s, new rank: %d", event.UserID, newRank)

        // Подтверждаем обработку сообщения (для JetStream)
        msg.Ack()
    }, nats.Durable("leaderboard-durable"), nats.ManualAck())

    if err != nil {
        log.Printf("Warning: failed to subscribe to score.updated: %v", err)
    } else {
        log.Println("📡 Subscribed to NATS: score.updated")
    }

    // Создаём gRPC хендлер
    leaderboardHandler := handler.NewLeaderboardHandler(leaderboardService)

    // Настраиваем gRPC сервер
    grpcServer := grpc.NewServer()
    pb.RegisterLeaderboardServiceServer(grpcServer, leaderboardHandler)
    reflection.Register(grpcServer)

    // Запускаем listener
    lis, err := net.Listen("tcp", ":"+cfg.GRPCPort)
    if err != nil {
        log.Fatalf("Failed to listen: %v", err)
    }

    // Graceful shutdown
    go func() {
        sigCh := make(chan os.Signal, 1)
        signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)
        <-sigCh
        log.Println("Shutting down gracefully...")
        grpcServer.GracefulStop()
        nc.Drain()
        os.Exit(0)
    }()

    log.Printf("🚀 Leaderboard service listening on :%s", cfg.GRPCPort)
    log.Printf("   Redis: %s (DB %d)", cfg.RedisAddr, cfg.RedisDB)
    log.Printf("   NATS: %s", cfg.NATSUrl)

    if err := grpcServer.Serve(lis); err != nil {
        log.Fatalf("Failed to serve: %v", err)
    }
}
