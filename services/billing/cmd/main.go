package main

import (
    "context"
    "encoding/json"
    "log"
    "net"
    "os"
    "os/signal"
    "syscall"
    // "time" — УДАЛИТЬ, не используется

    "github.com/jackc/pgx/v5/pgxpool"
    "github.com/nats-io/nats.go"
    "google.golang.org/grpc"
    "google.golang.org/grpc/reflection"

    "event_horizon/services/billing/internal/config"
    "event_horizon/services/billing/internal/handler"
    "event_horizon/services/billing/internal/repository"
    "event_horizon/services/billing/internal/service"
    pb "event_horizon/services/billing/proto"
)

// ScoreEvent структура для парсинга NATS сообщений
type ScoreEvent struct {
    UserID        string `json:"user_id"`
    GameID        string `json:"game_id"`
    Score         int    `json:"score"`
    IsRecord      bool   `json:"is_record"`
    Level         int    `json:"level"`
    LampsEarned   int    `json:"lamps_earned"`
    TicketsEarned int    `json:"tickets_earned"`
    Timestamp     int64  `json:"timestamp"`
}

func main() {
    cfg := config.Load()

    // Подключение к PostgreSQL
    dbURL := "postgres://" + cfg.DBUser + ":" + cfg.DBPassword + "@" + cfg.DBHost + ":" + cfg.DBPort + "/" + cfg.DBName
    dbpool, err := pgxpool.New(context.Background(), dbURL)
    if err != nil {
        log.Fatalf("Unable to connect to database: %v", err)
    }
    defer dbpool.Close()

    // Подключение к Redis
    redisRepo := repository.NewRedisBillingRepo(cfg.RedisAddr, cfg.RedisDB)

    // Подключение к NATS
    nc, err := nats.Connect(cfg.NATSUrl)
    if err != nil {
        log.Fatalf("Failed to connect to NATS: %v", err)
    }
    defer nc.Drain()

    js, err := nc.JetStream()
    if err != nil {
        log.Fatalf("Failed to create JetStream context: %v", err)
    }

    // Репозитории и сервис
    pgRepo := repository.NewPostgresBillingRepo(dbpool)
    billingService := service.NewBillingService(pgRepo, redisRepo)
    billingHandler := handler.NewBillingHandler(billingService)

    // gRPC сервер
    grpcServer := grpc.NewServer()
    pb.RegisterBillingServiceServer(grpcServer, billingHandler)
    reflection.Register(grpcServer)

    // Подписка на NATS (начисление валюты за рекорды)
    _, err = js.Subscribe("score.updated", func(msg *nats.Msg) {
        var event ScoreEvent
        if err := json.Unmarshal(msg.Data, &event); err != nil {
            log.Printf("Failed to unmarshal score event: %v", err)
            return
        }

        log.Printf("📡 Received score event for user %s, lamps=%d, tickets=%d",
            event.UserID, event.LampsEarned, event.TicketsEarned)

        // Начисление валюты (используем context.Background() вместо msg.Context)
        if event.LampsEarned > 0 {
            _, err := billingService.AddCurrency(context.Background(), event.UserID,
                repository.Lamps, event.LampsEarned, "game_reward", msg.Header.Get("Nats-Msg-Id"))
            if err != nil {
                log.Printf("Failed to add lamps: %v", err)
            } else {
                log.Printf("💰 Added %d lamps to user %s", event.LampsEarned, event.UserID)
            }
        }

        if event.TicketsEarned > 0 {
            _, err := billingService.AddCurrency(context.Background(), event.UserID,
                repository.Tickets, event.TicketsEarned, "game_reward", msg.Header.Get("Nats-Msg-Id"))
            if err != nil {
                log.Printf("Failed to add tickets: %v", err)
            } else {
                log.Printf("🎫 Added %d tickets to user %s", event.TicketsEarned, event.UserID)
            }
        }

        msg.Ack()
    }, nats.Durable("billing-durable"), nats.ManualAck())

    if err != nil {
        log.Printf("Warning: failed to subscribe to score.updated: %v", err)
    } else {
        log.Println("📡 Subscribed to NATS: score.updated (for rewards)")
    }

    // Listener
    lis, err := net.Listen("tcp", ":"+cfg.GRPCPort)
    if err != nil {
        log.Fatalf("Failed to listen: %v", err)
    }

    go func() {
        log.Printf("💰 Billing service listening on :%s", cfg.GRPCPort)
        log.Printf("   PostgreSQL: %s:%s/%s", cfg.DBHost, cfg.DBPort, cfg.DBName)
        log.Printf("   Redis: %s (DB %d)", cfg.RedisAddr, cfg.RedisDB)
        log.Printf("   NATS: %s", cfg.NATSUrl)
        if err := grpcServer.Serve(lis); err != nil {
            log.Fatalf("Failed to serve: %v", err)
        }
    }()

    // Graceful shutdown
    quit := make(chan os.Signal, 1)
    signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
    <-quit

    log.Println("Shutting down billing service gracefully...")

    grpcServer.GracefulStop()
    dbpool.Close()
    nc.Drain()

    log.Println("Billing service stopped gracefully")
}
