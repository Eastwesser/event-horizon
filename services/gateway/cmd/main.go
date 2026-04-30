package main

import (
    "encoding/json"
    "log"
    "net/http"

    "github.com/gin-gonic/gin"
    "github.com/nats-io/nats.go"

    "event_horizon/services/gateway/internal/client"
    pb "event_horizon/services/auth/proto"
)

func main() {
    // Подключаемся к Auth gRPC
    authClient, err := client.NewAuthClient("localhost:50051")
    if err != nil {
        log.Fatalf("Failed to connect to auth: %v", err)
    }
    defer authClient.Close()

    // Подключаемся к NATS
    nc, err := nats.Connect("nats://localhost:4222")
    if err != nil {
        log.Fatalf("Failed to connect to NATS: %v", err)
    }
    defer nc.Close()

    // Создаём JetStream контекст
    js, err := nc.JetStream()
    if err != nil {
        log.Fatalf("Failed to create JetStream context: %v", err)
    }

    // Создаём поток для событий (если не существует)
    _, err = js.AddStream(&nats.StreamConfig{
        Name:     "EVENTS",
        Subjects: []string{"event.>"},
        Storage:  nats.FileStorage,
    })
    if err != nil {
        log.Printf("Stream might already exist: %v", err)
    }

    r := gin.Default()

    r.GET("/health", func(c *gin.Context) {
        c.JSON(http.StatusOK, gin.H{"status": "ok"})
    })

    // Регистрация с публикацией события
    r.POST("/api/auth/register", func(c *gin.Context) {
        var req pb.RegisterRequest
        if err := c.ShouldBindJSON(&req); err != nil {
            c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
            return
        }

        resp, err := authClient.GetClient().Register(c.Request.Context(), &req)
        if err != nil {
            c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
            return
        }

        // Публикуем событие в NATS
        eventData := map[string]interface{}{
            "event":   "user.registered",
            "user_id": resp.UserId,
            "email":   resp.Email,
        }
        eventJSON, _ := json.Marshal(eventData)
        js.Publish("event.user.registered", eventJSON)
        log.Printf("📡 Published event: user.registered for %s", resp.Email)

        c.JSON(http.StatusOK, resp)
    })

    // Логин с публикацией события
    r.POST("/api/auth/login", func(c *gin.Context) {
        var req pb.LoginRequest
        if err := c.ShouldBindJSON(&req); err != nil {
            c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
            return
        }

        resp, err := authClient.GetClient().Login(c.Request.Context(), &req)
        if err != nil {
            c.JSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
            return
        }

        // Публикуем событие логина
        eventData := map[string]interface{}{
            "event": "user.logged_in",
            "email": req.Email,
        }
        eventJSON, _ := json.Marshal(eventData)
        js.Publish("event.user.logged_in", eventJSON)

        c.JSON(http.StatusOK, resp)
    })

    log.Println("🚀 Gateway listening on :8080 (with NATS JetStream)")
    if err := r.Run(":8080"); err != nil {
        log.Fatalf("Failed to start server: %v", err)
    }
}
