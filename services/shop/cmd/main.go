package main

import (
    "database/sql"
    "log"
    "net"
    "net/http"
    "os"
    "os/signal"
    "syscall"
    "time"

    _ "github.com/lib/pq"
    "github.com/nats-io/nats.go"
    "github.com/prometheus/client_golang/prometheus/promhttp"
    "google.golang.org/grpc"
    "google.golang.org/grpc/reflection"

    "github.com/Eastwesser/event-horizon/services/shop/internal/config"
    "github.com/Eastwesser/event-horizon/services/shop/internal/handler"
    "github.com/Eastwesser/event-horizon/services/shop/internal/repository"
    "github.com/Eastwesser/event-horizon/services/shop/internal/service"
    pb "github.com/Eastwesser/event-horizon/services/shop/proto"
)

func main() {
    cfg := config.Load()

    // 1. Подключение к PostgreSQL
    dbURL := "postgres://" + cfg.DBUser + ":" + cfg.DBPassword + "@" + cfg.DBHost + ":" + cfg.DBPort + "/" + cfg.DBName + "?sslmode=disable"
    db, err := sql.Open("postgres", dbURL)
    if err != nil {
        log.Fatalf("Unable to connect to database: %v", err)
    }
    defer db.Close()

    // 2. Подключение к Redis
    redisRepo := repository.NewRedisShopRepo(cfg.RedisAddr, cfg.RedisDB)

    // 3. Подключение к NATS
    var nc *nats.Conn
    for i := 0; i < 30; i++ {
        nc, err = nats.Connect(cfg.NATSUrl)
        if err == nil {
            break
        }
        log.Printf("Failed to connect to NATS (attempt %d/30): %v", i+1, err)
        time.Sleep(1 * time.Second)
    }
    if err != nil {
        log.Fatalf("Failed to connect to NATS after 30 attempts: %v", err)
    }
    defer nc.Close()

    js, err := nc.JetStream()
    if err != nil {
        log.Fatalf("Failed to create JetStream context: %v", err)
    }

    // 4. Репозитории и сервис
    pgRepo := repository.NewPostgresShopRepo(db)
    shopService := service.NewShopService(pgRepo, redisRepo, js, cfg.BillingAddr)
    shopHandler := handler.NewShopHandler(shopService)

    // 5. gRPC сервер
    grpcServer := grpc.NewServer()
    pb.RegisterShopServiceServer(grpcServer, shopHandler)
    reflection.Register(grpcServer)

    // 6. Метрики
    go func() {
        http.Handle("/metrics", promhttp.Handler())
        log.Printf("📊 Metrics endpoint: http://localhost:%s/metrics", cfg.MetricsPort)
        if err := http.ListenAndServe(":"+cfg.MetricsPort, nil); err != nil {
            log.Printf("Metrics server error: %v", err)
        }
    }()

    // 7. Listener
    lis, err := net.Listen("tcp", ":"+cfg.GRPCPort)
    if err != nil {
        log.Fatalf("Failed to listen: %v", err)
    }

    go func() {
        log.Printf("🛒 Shop service listening on :%s", cfg.GRPCPort)
        if err := grpcServer.Serve(lis); err != nil {
            log.Fatalf("Failed to serve: %v", err)
        }
    }()

    // 8. Graceful shutdown
    quit := make(chan os.Signal, 1)
    signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
    <-quit

    log.Println("Shutting down shop service gracefully...")
    grpcServer.GracefulStop()
    nc.Drain()
    log.Println("Shop service stopped gracefully")
}
