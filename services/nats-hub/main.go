package main

import (
    "log"
    "os"
    "time"

    "github.com/nats-io/nats.go"
)

func main() {
    natsURL := os.Getenv("NATS_URL")
    if natsURL == "" {
        natsURL = "nats://localhost:4222"
    }

    nc, err := nats.Connect(natsURL)
    if err != nil {
        log.Fatalf("Failed to connect to NATS: %v", err)
    }
    defer nc.Close()

    js, err := nc.JetStream()
    if err != nil {
        log.Fatalf("Failed to create JetStream context: %v", err)
    }

    // Создаём Stream для событий
    _, err = js.AddStream(&nats.StreamConfig{
        Name:     "EVENTS",
        Subjects: []string{
            "event.>",
            "score.updated",
            "user.registered",
            "shop.purchased",
            "payment.completed",
        },
        Storage:  nats.FileStorage,
        MaxAge:   7 * 24 * time.Hour,
    })
    if err != nil {
        log.Printf("Stream might already exist: %v", err)
    } else {
        log.Println("✅ Stream EVENTS created")
    }

    // Держим сервис запущенным
    select {}
}
