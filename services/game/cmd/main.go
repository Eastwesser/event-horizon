package main

import (
    "log"
    "net"
    "os"
    "os/signal"
    "syscall"

    "github.com/nats-io/nats.go"
    "google.golang.org/grpc"
    "google.golang.org/grpc/reflection"

    "event_horizon/services/game/internal/config"
    "event_horizon/services/game/internal/handler"
    "event_horizon/services/game/internal/repository"
    "event_horizon/services/game/internal/service"
    pb "event_horizon/services/game/proto"
)

func main() {
    cfg := config.Load()

    // Подключаемся к NATS
    nc, err := nats.Connect(cfg.NATSUrl)
    if err != nil {
        log.Fatalf("Failed to connect to NATS: %v", err)
    }
    defer nc.Close()

    js, err := nc.JetStream()
    if err != nil {
        log.Fatalf("Failed to create JetStream context: %v", err)
    }

    // Создаём репозиторий и сервис
    gameRepo := repository.NewPostgresGameRepo()
    gameService := service.NewGameService(gameRepo, js)
    gameHandler := handler.NewGameHandler(gameService)

    // gRPC сервер
    grpcServer := grpc.NewServer()
    pb.RegisterGameServiceServer(grpcServer, gameHandler)
    reflection.Register(grpcServer)

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

    lis, err := net.Listen("tcp", ":"+cfg.GRPCPort)
    if err != nil {
        log.Fatalf("Failed to listen: %v", err)
    }

    log.Printf("🎮 Game service listening on :%s", cfg.GRPCPort)
    log.Printf("   NATS: %s", cfg.NATSUrl)

    if err := grpcServer.Serve(lis); err != nil {
        log.Fatalf("Failed to serve: %v", err)
    }
}
