package main

import (
    "context"
    "encoding/json"
    "fmt"
    "log"
    "net"
    "net/http"
    "os"
    "os/signal"
    "syscall"
    "time"

    "github.com/jackc/pgx/v5/pgxpool"
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

    "github.com/Eastwesser/event-horizon/services/billing/internal/config"
    "github.com/Eastwesser/event-horizon/services/billing/internal/handler"
    "github.com/Eastwesser/event-horizon/services/billing/internal/repository"
    "github.com/Eastwesser/event-horizon/services/billing/internal/service"
    pb "github.com/Eastwesser/event-horizon/services/billing/proto"
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
            semconv.ServiceNameKey.String("billing"),
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

    // gRPC сервер с интерсепторами для трейсинга
    grpcServer := grpc.NewServer(
        grpc.UnaryInterceptor(otelgrpc.UnaryServerInterceptor()),
        grpc.StreamInterceptor(otelgrpc.StreamServerInterceptor()),
    )
    pb.RegisterBillingServiceServer(grpcServer, billingHandler)
    reflection.Register(grpcServer)

    // Метрики
    go func() {
        http.Handle("/metrics", promhttp.Handler())
        log.Printf("📊 Metrics endpoint: http://localhost:9093/metrics")
        if err := http.ListenAndServe(":9093", nil); err != nil {
            log.Printf("Billing metrics server error: %v", err)
        }
    }()

    // Подписка на NATS (начисление валюты за рекорды)
    _, err = js.Subscribe("score.updated", func(msg *nats.Msg) {

        var event ScoreEvent
        if err := json.Unmarshal(msg.Data, &event); err != nil {
            log.Printf("Failed to unmarshal score event: %v", err)
            return
        }
        log.Printf("📦 1 Full event: %+v", event)
        log.Printf("📡 Received score event for user %s, lamps=%d, tickets=%d",
            event.UserID, event.LampsEarned, event.TicketsEarned)

        referenceID := msg.Header.Get("Nats-Msg-Id")
        if referenceID == "" {
            referenceID = fmt.Sprintf("%s-%d", event.UserID, time.Now().UnixNano())
        }

        /*
            Проблема: один referenceID для двух валют

            Ты используешь один и тот же referenceID и для лампочек, и для тикетов.
            Если лампочки уже записали транзакцию с этим ID, то тикеты не могут записать ту же самую.
        */

        // Уникальный ID для лампочек
        lampsRefID := fmt.Sprintf("%s-lamps-%d", event.UserID, time.Now().UnixNano())

        // Уникальный ID для тикетов
        ticketsRefID := fmt.Sprintf("%s-tickets-%d", event.UserID, time.Now().UnixNano())

        // Начисление валюты LAMPS (используем context.Background() вместо msg.Context)
        if event.LampsEarned > 0 {
            _, err := billingService.AddCurrency(context.Background(), event.UserID,
                repository.Lamps, event.LampsEarned, "game_reward", lampsRefID)
            if err != nil {
                log.Printf("Failed to add lamps: %v", err)
            } else {
                log.Printf("💰 Added %d lamps to user %s", event.LampsEarned, event.UserID)
            }
        }
        log.Printf("📦 Lamps event: %+v", event)

        // Начисление валюты TICKETS (используем context.Background() вместо msg.Context)
        if event.TicketsEarned > 0 {
            _, err := billingService.AddCurrency(context.Background(), event.UserID,
                repository.Tickets, event.TicketsEarned, "game_reward", ticketsRefID)
            if err != nil {
                log.Printf("Failed to add tickets: %v", err)
            } else {
                log.Printf("🎫 Added %d tickets to user %s", event.TicketsEarned, event.UserID)
            }
        }
        log.Printf("📦 Tickets event: %+v", event)

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
