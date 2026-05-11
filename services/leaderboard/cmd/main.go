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

    // Batch ack channel (Канал для batch ack)
    ackChan := make(chan *nats.Msg, 1000)
    ticker := time.NewTicker(100 * time.Millisecond)

    go func() {
        batch := make([]*nats.Msg, 0, 100)
        for {
            select {
            case msg := <-ackChan:
                batch = append(batch, msg)
                if len(batch) >= 100 {
                    for _, m := range batch {
                        m.Ack()
                    }
                    batch = batch[:0]
                }
            case <-ticker.C:
                if len(batch) > 0 {
                    for _, m := range batch {
                        m.Ack()
                    }
                    batch = batch[:0]
                }
            }
        }
    }()

    // Единая подписка
    _, err = js.Subscribe("score.updated", func(msg *nats.Msg) {
        var event ScoreEvent
        if err := json.Unmarshal(msg.Data, &event); err != nil {
            log.Printf("Failed to unmarshal: %v", err)
            return
        }

        log.Printf("📡 Received score: game=%s user=%s score=%d", event.GameID, event.UserID, event.Score)

        // Быстрое обновление без ранга
        if err := leaderboardService.UpdateScoreOnly(context.Background(), event.GameID, event.UserID, event.UserEmail, event.Score); err != nil {
            log.Printf("Failed to update score: %v", err)
        }

        ackChan <- msg
    }, nats.Durable("leaderboard-durable"), nats.ManualAck())

    if err != nil {
        log.Printf("Warning: failed to subscribe: %v", err)
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
