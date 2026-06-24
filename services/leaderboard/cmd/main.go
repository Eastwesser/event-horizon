package main

import (
    "context"
    "encoding/json"
    "log"
    "net"
    "net/http"
    "os"
    "os/signal"
    "strings"
    "syscall"
    "time"

    "github.com/nats-io/nats.go"
    "github.com/prometheus/client_golang/prometheus/promhttp"
    "google.golang.org/grpc"
    "google.golang.org/grpc/reflection"
    "go.opentelemetry.io/contrib/instrumentation/google.golang.org/grpc/otelgrpc"
    "go.opentelemetry.io/otel"
    "go.opentelemetry.io/otel/attribute"
    "go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracegrpc"
    "go.opentelemetry.io/otel/propagation"
    "go.opentelemetry.io/otel/sdk/resource"
    "go.opentelemetry.io/otel/sdk/trace"
    semconv "go.opentelemetry.io/otel/semconv/v1.26.0"

    "github.com/Eastwesser/event-horizon/services/leaderboard/internal/config"
    "github.com/Eastwesser/event-horizon/services/leaderboard/internal/handler"
    "github.com/Eastwesser/event-horizon/services/leaderboard/internal/repository"
    "github.com/Eastwesser/event-horizon/services/leaderboard/internal/service"
    pb "github.com/Eastwesser/event-horizon/services/leaderboard/proto"
)

// ScoreEvent структура для входящих сообщений из NATS
type ScoreEvent struct {
    UserID    string `json:"user_id"`
    GameID    string `json:"game_id"`
    UserEmail string `json:"user_email"`
    Nickname  string `json:"nickname"`
    Score     int    `json:"score"`
}

// Инициализация OpenTelemetry для Jaeger
func initTracer(ctx context.Context) (func(context.Context) error, error) {
    // Читаем эндпоинт из переменной окружения
    endpoint := os.Getenv("JAEGER_ENDPOINT")
    if endpoint == "" {
        endpoint = "localhost:4317"
    }
    log.Printf("🔄 Initializing Jaeger tracer with endpoint: %s", endpoint)

    exporter, err := otlptracegrpc.New(ctx,
        otlptracegrpc.WithEndpoint(endpoint),
        otlptracegrpc.WithInsecure(),
    )
    if err != nil {
        log.Printf("❌ Failed to create exporter: %v", err)
        return nil, err
    }
    log.Println("✅ Jaeger exporter created")

    tp := trace.NewTracerProvider(
        trace.WithBatcher(exporter),
        trace.WithResource(resource.NewWithAttributes(
            semconv.SchemaURL,
            semconv.ServiceNameKey.String("leaderboard"),
            attribute.String("environment", "development"),
        )),
    )
    otel.SetTracerProvider(tp)
    otel.SetTextMapPropagator(propagation.NewCompositeTextMapPropagator(
        propagation.TraceContext{},
        propagation.Baggage{},
    ))

    log.Println("✅ Jaeger tracer initialized")
    return tp.Shutdown, nil
}

func main() {
    cfg := config.Load()
    ctx := context.Background()

    // Инициализация Jaeger
    shutdown, err := initTracer(ctx)
    if err != nil {
        log.Fatalf("Failed to initialize tracer: %v", err)
    }
    defer shutdown(ctx)

    // Подключаемся к Redis
    redisRepo := repository.NewRedisLeaderboardRepo(cfg.RedisAddr, cfg.RedisDB)

    // Создаём сервис
    leaderboardService := service.NewLeaderboardService(redisRepo)

    // Подключаемся к NATS
    var nc *nats.Conn
    var lastErr error
    for i := 0; i < 30; i++ {
        nc, lastErr = nats.Connect(cfg.NATSUrl)
        if lastErr == nil {
            break
        }
        log.Printf("Failed to connect to NATS (attempt %d/30): %v", i+1, lastErr)
        time.Sleep(1 * time.Second)
    }
    if lastErr != nil {
        log.Fatalf("Failed to connect to NATS after 30 attempts: %v", lastErr)
    }

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
        if strings.Contains(err.Error(), "stream name already in use") {
            log.Println("✅ Stream already exists")
        } else {
            log.Fatalf("Failed to create stream: %v", err)
        }
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

        // Создаём контекст с таймаутом 5 секунд
        ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
        defer cancel()

        /*
            Что это даёт: 
            Если Redis или PostgreSQL зависнут, запрос не будет висеть вечно — через 5 секунд он прервётся.
        */

        // Сохраняем информацию о пользователе
        if err := leaderboardService.SaveUserInfo(ctx, event.GameID, event.UserID, event.UserEmail, event.Nickname); err != nil {
            log.Printf("Failed to save user info: %v", err)
        }

        // Быстрое обновление без ранга
        if err := leaderboardService.UpdateScoreOnly(ctx, event.GameID, event.UserID, event.UserEmail, event.Score); err != nil {
            log.Printf("Failed to update score: %v", err)
        }

        ackChan <- msg
    }, nats.Durable("leaderboard-durable"), nats.ManualAck())

    if err != nil {
        log.Fatalf("Failed to subscribe: %v", err)
    } else {
        log.Println("📡 Subscribed to NATS: score.updated")
    }

    // Создаём gRPC хендлер
    leaderboardHandler := handler.NewLeaderboardHandler(leaderboardService)

    // Настраиваем gRPC сервер с интерсепторами для трейсинга
    grpcServer := grpc.NewServer(
        grpc.UnaryInterceptor(otelgrpc.UnaryServerInterceptor()),
        grpc.StreamInterceptor(otelgrpc.StreamServerInterceptor()),
    )
    
    pb.RegisterLeaderboardServiceServer(grpcServer, leaderboardHandler)
    reflection.Register(grpcServer)

    // Метрики
    go func() {
        http.Handle("/metrics", promhttp.Handler())
        log.Printf("📊 Metrics endpoint: http://localhost:9094/metrics")
        if err := http.ListenAndServe(":9094", nil); err != nil {
            log.Printf("Leaderboard metrics server error: %v", err)
        }
    }()

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
