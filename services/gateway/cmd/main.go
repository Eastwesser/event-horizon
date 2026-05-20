package main

import (
    "bytes"
    "context"
    "encoding/json"
    "io"
    "log"
    "net/http"
    "os"
    "os/signal"
    "strconv"
    "syscall"
    "time"

    "github.com/gin-gonic/gin"
    "github.com/gorilla/websocket"
    "github.com/nats-io/nats.go"
    "google.golang.org/grpc"
    "google.golang.org/grpc/credentials/insecure"

    "event_horizon/services/gateway/internal/cache"
    "event_horizon/services/gateway/internal/client"
    authPb "event_horizon/services/auth/proto"
    gamePb "event_horizon/services/game/proto"
    leaderboardPb "event_horizon/services/leaderboard/proto"
)

var upgrader = websocket.Upgrader{
    CheckOrigin: func(r *http.Request) bool { return true },
}

type WSClient struct {
    hub  *Hub
    conn *websocket.Conn
    send chan []byte
}

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
    hub := NewHub()
    go hub.Run()

    authClient, err := client.NewAuthClient("localhost:50051")
    if err != nil {
        log.Fatalf("Failed to connect to auth: %v", err)
    }
    defer authClient.Close()

    gameConn, err := grpc.NewClient("localhost:50052", grpc.WithTransportCredentials(insecure.NewCredentials()))
    if err != nil {
        log.Fatalf("Failed to connect to game: %v", err)
    }
    defer gameConn.Close()
    gameClient := gamePb.NewGameServiceClient(gameConn)

    nc, err := nats.Connect("nats://localhost:4222")
    if err != nil {
        log.Fatalf("Failed to connect to NATS: %v", err)
    }
    defer nc.Drain()

    js, err := nc.JetStream()
    if err != nil {
        log.Fatalf("Failed to create JetStream context: %v", err)
    }

    _, err = js.AddStream(&nats.StreamConfig{
        Name:     "EVENTS",
        Subjects: []string{"event.>", "score.updated"},
        Storage:  nats.FileStorage,
    })
    if err != nil {
        log.Printf("Stream might already exist: %v", err)
    }

    _, err = js.Subscribe("score.updated", func(msg *nats.Msg) {
        hub.Broadcast(msg.Data)
        log.Printf("📡 Broadcasted score update to WebSocket clients")
    }, nats.Durable("gateway-websocket"))
    if err != nil {
        log.Printf("Failed to subscribe to score.updated: %v", err)
    }

    r := gin.Default()

    r.GET("/health", func(c *gin.Context) {
        c.JSON(http.StatusOK, gin.H{"status": "ok"})
    })

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
        c.JSON(http.StatusOK, gin.H{
            "access_token": resp.AccessToken,
            "token_type":   resp.TokenType,
            "expires_in":   resp.ExpiresIn,
            "user_id":      resp.UserId,
        })
    })

    scoreCache := cache.NewScoreCache(2 * time.Second)

    r.GET("/api/billing/balance/all", func(c *gin.Context) {
        token := c.GetHeader("Authorization")
        if token == "" {
            c.JSON(http.StatusUnauthorized, gin.H{"error": "Missing authorization header"})
            return
        }
        c.JSON(http.StatusOK, gin.H{"lamps": 0, "tickets": 0})
    })

    r.GET("/api/leaderboard", func(c *gin.Context) {
        gameID := c.Query("game_id")
        limit := c.Query("limit")
        
        // Вызываем leaderboard через gRPC
        conn, err := grpc.Dial("localhost:50054", grpc.WithTransportCredentials(insecure.NewCredentials()))
        if err != nil {
            c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
            return
        }
        defer conn.Close()
        
        leaderboardClient := leaderboardPb.NewLeaderboardServiceClient(conn)
        
        // Преобразуем limit в int32
        var limitInt int32 = 10
        if limit != "" {
            if l, err := strconv.Atoi(limit); err == nil {
                limitInt = int32(l)
            }
        }
        
        resp, err := leaderboardClient.GetTopScores(c.Request.Context(), &leaderboardPb.GetTopScoresRequest{
            GameId: gameID,
            Limit:  limitInt,
        })
        if err != nil {
            c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
            return
        }
        
        c.JSON(http.StatusOK, gin.H{"entries": resp.Entries})
    })

    r.POST("/api/game/submit", func(c *gin.Context) {
        body, _ := c.GetRawData()
        c.Request.Body = io.NopCloser(bytes.NewBuffer(body))

        cacheKey := string(body)
        if cached, ok := scoreCache.Get(cacheKey); ok {
            c.Data(http.StatusOK, "application/json", cached)
            return
        }

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
            UserId: req.UserID,
            GameId: req.GameID,
            Level:  req.Level,
            Seed:   req.Seed,
            Moves:  moves,
        })
        if err != nil {
            c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
            return
        }

        respJSON, _ := json.Marshal(resp)
        scoreCache.Set(cacheKey, respJSON)
        c.JSON(http.StatusOK, resp)
    })

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