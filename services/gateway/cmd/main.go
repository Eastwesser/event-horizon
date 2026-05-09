package main

import (
    "context"
    "encoding/json"
    "log"
    "net/http"
    "os"
    "os/signal"
    "syscall"
    "time"

    "github.com/gin-gonic/gin"
    "github.com/gorilla/websocket"
    "github.com/nats-io/nats.go"
    "google.golang.org/grpc"
    "google.golang.org/grpc/credentials/insecure"

    "event_horizon/services/gateway/internal/client"
    authPb "event_horizon/services/auth/proto"
    gamePb "event_horizon/services/game/proto"
)

var upgrader = websocket.Upgrader{
    CheckOrigin: func(r *http.Request) bool { return true },
}

// WebSocket клиент
type WSClient struct {
    hub  *Hub
    conn *websocket.Conn
    send chan []byte
}

// Hub управляет всеми WebSocket соединениями
type Hub struct {
    clients    map[*WSClient]bool
    broadcast  chan []byte
    register   chan *WSClient
    unregister chan *WSClient
}

func NewHub() *Hub {
    return &Hub{
        clients:    make(map[*WSClient]bool),
        broadcast:  make(chan []byte),
        register:   make(chan *WSClient),
        unregister: make(chan *WSClient),
    }
}

func (h *Hub) Run() {
    for {
        select {
        case client := <-h.register:
            h.clients[client] = true
            log.Printf("🟢 WebSocket client connected. Total: %d", len(h.clients))

        case client := <-h.unregister:
            if _, ok := h.clients[client]; ok {
                delete(h.clients, client)
                close(client.send)
            }
            log.Printf("🔴 WebSocket client disconnected. Total: %d", len(h.clients))

        case message := <-h.broadcast:
            for client := range h.clients {
                select {
                case client.send <- message:
                default:
                    close(client.send)
                    delete(h.clients, client)
                }
            }
        }
    }
}

func (h *Hub) Broadcast(message []byte) {
    h.broadcast <- message
}

func (c *WSClient) readPump() {
    defer func() {
        c.hub.unregister <- c
        c.conn.Close()
    }()

    for {
        _, _, err := c.conn.ReadMessage()
        if err != nil {
            break
        }
        // Пока игнорируем сообщения от клиента (только для keep-alive)
    }
}

func (c *WSClient) writePump() {
    defer c.conn.Close()

    for {
        select {
        case message, ok := <-c.send:
            if !ok {
                c.conn.WriteMessage(websocket.CloseMessage, []byte{})
                return
            }
            c.conn.WriteMessage(websocket.TextMessage, message)
        }
    }
}

func main() {
    // Создаём WebSocket Hub
    hub := NewHub()
    go hub.Run()

    // Подключаемся к Auth gRPC
    authClient, err := client.NewAuthClient("localhost:50051")
    if err != nil {
        log.Fatalf("Failed to connect to auth: %v", err)
    }
    defer authClient.Close()

    // Подключаемся к Game gRPC
    gameConn, err := grpc.NewClient("localhost:50052", grpc.WithTransportCredentials(insecure.NewCredentials()))
    if err != nil {
        log.Fatalf("Failed to connect to game: %v", err)
    }
    defer gameConn.Close()
    gameClient := gamePb.NewGameServiceClient(gameConn)

    // Подключаемся к NATS
    nc, err := nats.Connect("nats://localhost:4222")
    if err != nil {
        log.Fatalf("Failed to connect to NATS: %v", err)
    }
    defer nc.Drain()

    js, err := nc.JetStream()
    if err != nil {
        log.Fatalf("Failed to create JetStream context: %v", err)
    }

    // Создаём поток для событий
    _, err = js.AddStream(&nats.StreamConfig{
        Name:     "EVENTS",
        Subjects: []string{"event.>", "score.updated"},
        Storage:  nats.FileStorage,
    })
    if err != nil {
        log.Printf("Stream might already exist: %v", err)
    }

    // Подписываемся на NATS для обновления топа
    _, err = js.Subscribe("score.updated", func(msg *nats.Msg) {
        hub.Broadcast(msg.Data)
        log.Printf("📡 Broadcasted score update to WebSocket clients")
    }, nats.Durable("gateway-websocket"))
    if err != nil {
        log.Printf("Failed to subscribe to score.updated: %v", err)
    }

    // Gin сервер
    r := gin.Default()

    r.GET("/health", func(c *gin.Context) {
        c.JSON(http.StatusOK, gin.H{"status": "ok"})
    })

    // WebSocket эндпоинт
    r.GET("/ws/leaderboard", func(c *gin.Context) {
        conn, err := upgrader.Upgrade(c.Writer, c.Request, nil)
        if err != nil {
            log.Printf("WebSocket upgrade failed: %v", err)
            return
        }

        client := &WSClient{
            hub:  hub,
            conn: conn,
            send: make(chan []byte, 256),
        }
        hub.register <- client

        go client.writePump()
        go client.readPump()
    })

    // Регистрация
    r.POST("/api/auth/register", func(c *gin.Context) {
        var req authPb.RegisterRequest
        if err := c.ShouldBindJSON(&req); err != nil {
            c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
            return
        }

        resp, err := authClient.GetClient().Register(c.Request.Context(), &req)
        if err != nil {
            c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
            return
        }

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

    // Логин
    r.POST("/api/auth/login", func(c *gin.Context) {
        var req authPb.LoginRequest
        if err := c.ShouldBindJSON(&req); err != nil {
            c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
            return
        }

        resp, err := authClient.GetClient().Login(c.Request.Context(), &req)
        if err != nil {
            c.JSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
            return
        }

        eventData := map[string]interface{}{
            "event": "user.logged_in",
            "email": req.Email,
        }
        eventJSON, _ := json.Marshal(eventData)
        js.Publish("event.user.logged_in", eventJSON)

        c.JSON(http.StatusOK, resp)
    })

    // Submit score
    r.POST("/api/game/submit", func(c *gin.Context) {
        var req struct {
            UserID string `json:"user_id"`
            GameID string `json:"game_id"`
            Level  int32  `json:"level"`
            Seed   string `json:"seed"`
            Moves  []struct {
                FromX     int32 `json:"fromX"`
                FromY     int32 `json:"fromY"`
                ToX       int32 `json:"toX"`
                ToY       int32 `json:"toY"`
                Timestamp int32 `json:"timestamp"`
            } `json:"moves"`
        }

        if err := c.ShouldBindJSON(&req); err != nil {
            c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
            return
        }

        moves := make([]*gamePb.Move, len(req.Moves))
        for i, m := range req.Moves {
            moves[i] = &gamePb.Move{
                FromX:     m.FromX,
                FromY:     m.FromY,
                ToX:       m.ToX,
                ToY:       m.ToY,
                Timestamp: m.Timestamp,
            }
        }

        resp, err := gameClient.SubmitScore(c.Request.Context(), &gamePb.SubmitScoreRequest{
            UserId:    req.UserID,
            GameId:    req.GameID,
            Level:     req.Level,
            Seed:      req.Seed,
            Moves:     moves,
        })
        if err != nil {
            c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
            return
        }

        c.JSON(http.StatusOK, resp)
    })

    // HTTP сервер
    srv := &http.Server{
        Addr:    ":8080",
        Handler: r,
    }

    go func() {
        if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
            log.Fatalf("Failed to start server: %v", err)
        }
    }()

    log.Println("🚀 Gateway listening on :8080 (with NATS JetStream & WebSocket)")
    log.Println("   WebSocket endpoint: ws://localhost:8080/ws/leaderboard")

    // Graceful shutdown
    quit := make(chan os.Signal, 1)
    signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
    <-quit

    log.Println("Shutting down gateway gracefully...")

    ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
    defer cancel()

    if err := srv.Shutdown(ctx); err != nil {
        log.Printf("HTTP server shutdown error: %v", err)
    }

    authClient.Close()
    gameConn.Close()
    nc.Drain()

    log.Println("Gateway stopped gracefully")
}
