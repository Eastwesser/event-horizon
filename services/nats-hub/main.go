package main

import (
    "log"
    "net/http"
    "os"
    "os/signal"
    "syscall"
    "time"

    "github.com/nats-io/nats.go"
)

func main() {
    natsURL := os.Getenv("NATS_URL")
    if natsURL == "" {
        natsURL = "nats://localhost:4222"
    }

    log.Printf("🚀 Starting NATS Hub with URL: %s", natsURL)

    nc, err := nats.Connect(natsURL)
    if err != nil {
        log.Fatalf("Failed to connect to NATS: %v", err)
    }
    defer nc.Drain()

    js, err := nc.JetStream()
    if err != nil {
        log.Fatalf("Failed to create JetStream context: %v", err)
    }

    // Создаём Stream для событий
    streamConfig := &nats.StreamConfig{
        Name:     "EVENTS",
        Subjects: []string{
            "event.>",
            "score.updated",
            "user.registered",
            "shop.purchased",
            "payment.completed",
            "inventory.item.created",
            "inventory.item.updated",
            "inventory.item.deleted",
        },
        Storage:  nats.FileStorage,
        MaxAge:   7 * 24 * time.Hour,
        // MaxMsg:   1_000_000,
        MaxBytes: 1024 * 1024 * 1024, // 1 GB
    }

    info, err := js.AddStream(streamConfig)
    if err != nil {
        log.Printf("⚠️ Stream might already exist: %v", err)
        info, err = js.StreamInfo("EVENTS")
        if err != nil {
            log.Printf("❌ Failed to get stream info: %v", err)
        } else {
            log.Printf("✅ Stream EVENTS exists: %d messages", info.State.Msgs)
        }
    } else {
        log.Printf("✅ Stream EVENTS created")
    }

    // Подписываемся на все события для логирования (опционально)
    _, err = js.Subscribe("event.>", func(msg *nats.Msg) {
        log.Printf("📡 Event: %s | Size: %d bytes", msg.Subject, len(msg.Data))
        msg.Ack()
    }, nats.Durable("nats-hub-listener"))
    if err != nil {
        log.Printf("⚠️ Failed to subscribe to events: %v", err)
    } else {
        log.Println("✅ Subscribed to all events for logging")
    }

    // HTTP сервер для health check и метрик
    go func() {
        http.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
            if nc.IsConnected() {
                w.WriteHeader(http.StatusOK)
                w.Write([]byte(`{"status":"ok","service":"nats-hub"}`))
            } else {
                w.WriteHeader(http.StatusServiceUnavailable)
                w.Write([]byte(`{"status":"degraded","service":"nats-hub"}`))
            }
        })
        http.HandleFunc("/metrics", func(w http.ResponseWriter, r *http.Request) {
            w.WriteHeader(http.StatusOK)
            w.Write([]byte(
                `# HELP nats_hub_connected NATS connection status (1=connected)\n` +
                    `# TYPE nats_hub_connected gauge\n` +
                    `nats_hub_connected 1\n`,
            ))
        })
        log.Println("📊 Health check: http://localhost:9097/health")
        if err := http.ListenAndServe(":9097", nil); err != nil {
            log.Printf("Health check server error: %v", err)
        }
    }()

    // Graceful shutdown
    quit := make(chan os.Signal, 1)
    signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)

    log.Println("📡 NATS Hub is running")
    log.Println("   Subjects: event.>, score.updated, user.registered, shop.purchased,")
    log.Println("             payment.completed, inventory.item.*")
    log.Println("   Press Ctrl+C to stop")

    <-quit
    log.Println("🛑 Shutting down NATS Hub gracefully...")

    time.Sleep(2 * time.Second)
    nc.Drain()

    log.Println("👋 NATS Hub stopped")
}