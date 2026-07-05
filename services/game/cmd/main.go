package main

import (
    "context"
    "database/sql"
    "log"
    "net"
    "net/http"
    _ "net/http/pprof"
    "os"
    "os/signal"
    "syscall"
    "time"

    _ "github.com/lib/pq"
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

    "github.com/Eastwesser/event-horizon/services/game/internal/config"
    "github.com/Eastwesser/event-horizon/services/game/internal/handler"
    "github.com/Eastwesser/event-horizon/services/game/internal/repository"
    "github.com/Eastwesser/event-horizon/services/game/internal/service"
    pb "github.com/Eastwesser/event-horizon/services/game/proto"
)

// Инициализация OpenTelemetry для Jaeger
func initTracer(ctx context.Context) (func(context.Context) error, error) {
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
            semconv.ServiceNameKey.String("game"),
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

    // 1. Инициализация Jaeger
    shutdown, err := initTracer(ctx)
    if err != nil {
        log.Fatalf("Failed to initialize tracer: %v", err)
    }
    defer shutdown(ctx)

    // 2. Подключение к PostgreSQL
    dbURL := "postgres://" + cfg.DBUser + ":" + cfg.DBPassword + "@" + cfg.DBHost + ":" + cfg.DBPort + "/" + cfg.DBName + "?sslmode=disable"

    db, err := sql.Open("postgres", dbURL)
    if err != nil {
        log.Fatalf("Unable to connect to database: %v", err)
    }
    defer db.Close()

    // 3. Подключение к NATS
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

    js, err := nc.JetStream()
    if err != nil {
        log.Fatalf("Failed to create JetStream context: %v", err)
    }

    // 4. Репозиторий и сервис
    gameRepo := repository.NewPostgresGameRepo(db)
    gameService := service.NewGameService(gameRepo, js)
    gameHandler := handler.NewGameHandler(gameService)

    // 5. gRPC сервер
    grpcServer := grpc.NewServer(
        grpc.UnaryInterceptor(otelgrpc.UnaryServerInterceptor()),
        grpc.StreamInterceptor(otelgrpc.StreamServerInterceptor()),
    )
    pb.RegisterGameServiceServer(grpcServer, gameHandler)
    reflection.Register(grpcServer)

    // 6. Метрики
    go func() {
        http.Handle("/metrics", promhttp.Handler())
        log.Printf("📊 Metrics endpoint: http://localhost:9092/metrics")
        if err := http.ListenAndServe(":9092", nil); err != nil {
            log.Printf("Game metrics server error: %v", err)
        }
    }()

    // 7. Listener
    lis, err := net.Listen("tcp", ":"+cfg.GRPCPort)
    if err != nil {
        log.Fatalf("Failed to listen: %v", err)
    }

    go func() {
        log.Printf("🎮 Game service listening on :%s", cfg.GRPCPort)
        log.Printf("   NATS: %s", cfg.NATSUrl)
        if err := grpcServer.Serve(lis); err != nil {
            log.Fatalf("Failed to serve: %v", err)
        }
    }()

    // 8. Graceful shutdown
    quit := make(chan os.Signal, 1)
    signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
    <-quit

    log.Println("Shutting down game service gracefully...")

    grpcServer.GracefulStop()
    nc.Drain()

    log.Println("Game service stopped gracefully")
}