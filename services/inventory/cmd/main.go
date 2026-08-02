package main

import (
    "context"
    "database/sql"
    "fmt"
    "log"
    "net"
    "net/http"
    "os"
    "os/signal"
    "syscall"
    "time"

    "github.com/Eastwesser/event-horizon/services/inventory/internal/config"
    "github.com/Eastwesser/event-horizon/services/inventory/internal/handler"
    "github.com/Eastwesser/event-horizon/services/inventory/internal/repository"
    "github.com/Eastwesser/event-horizon/services/inventory/internal/service"
    "github.com/Eastwesser/event-horizon/services/inventory/internal/worker"
    pb "github.com/Eastwesser/event-horizon/services/inventory/proto"

    _ "github.com/lib/pq"
    "github.com/nats-io/nats.go"
    "github.com/prometheus/client_golang/prometheus/promhttp"
    "go.mongodb.org/mongo-driver/mongo"
    "go.mongodb.org/mongo-driver/mongo/options"
    "google.golang.org/grpc"
    "google.golang.org/grpc/health"
    "google.golang.org/grpc/health/grpc_health_v1"
    "google.golang.org/grpc/reflection"
)

func main() {
    // Загружаем конфиг
    cfg, err := config.Load()
    if err != nil {
        log.Fatalf("Failed to load config: %v", err)
    }

    ctx := context.Background()

    // Выбираем драйвер
    var repo repository.InventoryRepository
    var db *sql.DB
    var mongoClient *mongo.Client

    switch cfg.Driver {
    case "postgres":
        log.Println("Using PostgreSQL driver")
        repo, db, err = initPostgres(cfg)
    case "mongo":
        log.Println("Using MongoDB driver")
        repo, mongoClient, err = initMongo(cfg)
    default:
        log.Fatalf("Unknown driver: %s (use 'postgres' or 'mongo')", cfg.Driver)
    }

    if err != nil {
        log.Fatalf("Failed to initialize repository: %v", err)
    }

    // Инициализируем NATS
    nc, err := nats.Connect(cfg.NATSURL)
    if err != nil {
        log.Fatalf("Failed to connect to NATS: %v", err)
    }
    defer nc.Close()

    js, err := nc.JetStream()
    if err != nil {
        log.Fatalf("Failed to create JetStream context: %v", err)
    }

    // Создаём сервис с кешем и NATS
    cache := repository.NewRedisCacheRepo(cfg.RedisAddr, 5*time.Minute)
    svc := service.NewInventoryService(repo, cache, js)

    // Создаём gRPC сервер
    grpcServer := grpc.NewServer(
        grpc.MaxRecvMsgSize(10*1024*1024),
        grpc.MaxSendMsgSize(10*1024*1024),
    )

    // Регистрируем сервис (передаём svc!)
    h := handler.NewGRPCHandler(svc)
    pb.RegisterInventoryServiceServer(grpcServer, h)

    // Health check
    healthServer := health.NewServer()
    grpc_health_v1.RegisterHealthServer(grpcServer, healthServer)
    healthServer.SetServingStatus("", grpc_health_v1.HealthCheckResponse_SERVING)

    // Reflection для отладки
    reflection.Register(grpcServer)

    // HTTP сервер для health check и метрик
    mux := http.NewServeMux()
    
    mux.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
        status := "ok"
        httpStatus := http.StatusOK
        
        if cfg.Driver == "postgres" && db != nil {
            if err := db.Ping(); err != nil {
                status = "degraded"
                httpStatus = http.StatusServiceUnavailable
                log.Printf("Health check: DB ping failed: %v", err)
            }
        }
        
        if cfg.Driver == "mongo" && mongoClient != nil {
            ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
            defer cancel()
            if err := mongoClient.Ping(ctx, nil); err != nil {
                status = "degraded"
                httpStatus = http.StatusServiceUnavailable
                log.Printf("Health check: MongoDB ping failed: %v", err)
            }
        }
        
        w.Header().Set("Content-Type", "application/json")
        w.WriteHeader(httpStatus)
        fmt.Fprintf(w, `{"status":"%s","service":"inventory","driver":"%s"}`, status, cfg.Driver)
    })
    
    mux.Handle("/metrics", promhttp.Handler())
    
    go func() {
        metricsAddr := fmt.Sprintf(":%d", cfg.MetricsPort)
        log.Printf("Health check and metrics: http://localhost%s/health", metricsAddr)
        if err := http.ListenAndServe(metricsAddr, mux); err != nil {
            log.Printf("HTTP server error: %v", err)
        }
    }()

    // Запускаем OutboxWorker (только для PostgreSQL)
    if cfg.Driver == "postgres" {
        outboxWorker := worker.NewOutboxWorker(db, js)
        go outboxWorker.Start(ctx)
        log.Println("✅ Outbox worker started")
    }

    // Запускаем gRPC сервер
    addr := fmt.Sprintf(":%d", cfg.GRPCPort)
    lis, err := net.Listen("tcp", addr)
    if err != nil {
        log.Fatalf("Failed to listen: %v", err)
    }

    log.Printf("Inventory Service listening on %s (driver: %s)", addr, cfg.Driver)

    // Graceful shutdown
    go func() {
        sigCh := make(chan os.Signal, 1)
        signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)
        <-sigCh
        log.Println("Shutting down gracefully...")
        grpcServer.GracefulStop()
        
        if db != nil {
            db.Close()
        }
        if mongoClient != nil {
            ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
            defer cancel()
            mongoClient.Disconnect(ctx)
        }
        nc.Drain()
    }()

    if err := grpcServer.Serve(lis); err != nil {
        log.Fatalf("Failed to serve: %v", err)
    }
}

func initPostgres(cfg *config.Config) (repository.InventoryRepository, *sql.DB, error) {
    db, err := sql.Open("postgres", cfg.PGDSN())
    if err != nil {
        return nil, nil, fmt.Errorf("failed to open PostgreSQL: %w", err)
    }

    db.SetMaxOpenConns(25)
    db.SetMaxIdleConns(5)
    db.SetConnMaxLifetime(5 * time.Minute)

    if err := db.Ping(); err != nil {
        return nil, nil, fmt.Errorf("failed to ping PostgreSQL: %w", err)
    }

    log.Println("PostgreSQL connection established")
    return repository.NewPostgresRepo(db), db, nil
}

func initMongo(cfg *config.Config) (repository.InventoryRepository, *mongo.Client, error) {
    ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
    defer cancel()
    
    client, err := mongo.Connect(ctx, options.Client().ApplyURI(cfg.MongoURI))
    if err != nil {
        return nil, nil, fmt.Errorf("failed to connect to MongoDB: %w", err)
    }

    if err := client.Ping(ctx, nil); err != nil {
        return nil, nil, fmt.Errorf("failed to ping MongoDB: %w", err)
    }

    log.Println("MongoDB connection established")
    db := client.Database(cfg.MongoDBName)
    return repository.NewMongoRepo(db), client, nil
}
