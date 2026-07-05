package main

import (
    "context"
    "encoding/json"
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

    "github.com/Eastwesser/event-horizon/services/profile/internal/config"
    "github.com/Eastwesser/event-horizon/services/profile/internal/handler"
    "github.com/Eastwesser/event-horizon/services/profile/internal/repository"
    "github.com/Eastwesser/event-horizon/services/profile/internal/service"
    pb "github.com/Eastwesser/event-horizon/services/profile/proto"
)

type ScoreEvent struct {
    UserID        string `json:"user_id"`
    GameID        string `json:"game_id"`
    Score         int    `json:"score"`
    IsRecord      bool   `json:"is_record"`
    LampsEarned   int    `json:"lamps_earned"`
    TicketsEarned int    `json:"tickets_earned"`
}

type UserEvent struct {
    UserID string `json:"user_id"`
    Email  string `json:"email"`
}

func contains(slice []string, item string) bool {
    for _, s := range slice {
        if s == item {
            return true
        }
    }
    return false
}

func main() {
    cfg := config.Load()
    ctx := context.Background()

    // Подключаемся к PostgreSQL с retry
    dbURL := "postgres://" + cfg.DBUser + ":" + cfg.DBPassword + "@" + cfg.DBHost + ":" + cfg.DBPort + "/" + cfg.DBName

    var dbpool *pgxpool.Pool
    var dbErr error
    for i := 0; i < 30; i++ {
        dbpool, dbErr = pgxpool.New(ctx, dbURL)
        if dbErr == nil {
            if err := dbpool.Ping(ctx); err == nil {
                break
            }
        }
        log.Printf("Failed to connect to DB (attempt %d/30): %v", i+1, dbErr)
        time.Sleep(2 * time.Second)
    }
    if dbErr != nil {
        log.Fatalf("Unable to connect to database after 30 attempts: %v", dbErr)
    }
    defer dbpool.Close()
    log.Println("✅ Connected to PostgreSQL")

    // Подключаемся к NATS с retry
    var nc *nats.Conn
    var natsErr error
    for i := 0; i < 30; i++ {
        nc, natsErr = nats.Connect(cfg.NATSUrl)
        if natsErr == nil {
            break
        }
        log.Printf("Failed to connect to NATS (attempt %d/30): %v", i+1, natsErr)
        time.Sleep(1 * time.Second)
    }
    if natsErr != nil {
        log.Fatalf("Failed to connect to NATS after 30 attempts: %v", natsErr)
    }
    defer nc.Close()
    log.Println("✅ Connected to NATS")

    js, err := nc.JetStream()
    if err != nil {
        log.Fatalf("Failed to create JetStream context: %v", err)
    }

    // Создаём Stream для событий (если не существует)
    stream, err := js.StreamInfo("EVENTS")
	if err != nil {
		// Stream не существует — создаём
		_, err = js.AddStream(&nats.StreamConfig{
			Name:     "EVENTS",
			Subjects: []string{"event.>", "score.updated", "user.registered"},
			Storage:  nats.FileStorage,
		})
		if err != nil {
			log.Printf("Failed to create stream: %v", err)
		}
	} else {
		// Stream существует — проверяем и обновляем subjects
		if !contains(stream.Config.Subjects, "user.registered") {
			newSubjects := append(stream.Config.Subjects, "user.registered")
			_, err = js.UpdateStream(&nats.StreamConfig{
				Name:     "EVENTS",
				Subjects: newSubjects,
				Storage:  nats.FileStorage,
			})
			if err != nil {
				log.Printf("Failed to update stream: %v", err)
			} else {
				log.Println("✅ Updated EVENTS stream with user.registered")
			}
		}
	}

    // Инициализируем слои
    profileRepo := repository.NewPostgresProfileRepo(dbpool)
    profileService := service.NewProfileService(profileRepo)
    profileHandler := handler.NewProfileHandler(profileService)

    // gRPC сервер
    grpcServer := grpc.NewServer()
    pb.RegisterProfileServiceServer(grpcServer, profileHandler)
    reflection.Register(grpcServer)

    // Метрики
    go func() {
        http.Handle("/metrics", promhttp.Handler())
        log.Printf("📊 Metrics endpoint: http://localhost:%s/metrics", cfg.MetricsPort)
        if err := http.ListenAndServe(":"+cfg.MetricsPort, nil); err != nil {
            log.Printf("Metrics server error: %v", err)
        }
    }()

    // // Подписка на NATS (обновление профиля при регистрации)
    // _, err = js.Subscribe("user.registered", func(msg *nats.Msg) {
    //     var event UserEvent
    //     if err := json.Unmarshal(msg.Data, &event); err != nil {
    //         log.Printf("Failed to unmarshal user.registered: %v", err)
    //         return
    //     }

    //     profile := &repository.UserProfile{
    //         UserID:     event.UserID,
    //         Email:      event.Email,
    //         Nickname:   event.Email[:8],
    //         TotalScore: 0,
    //         BestScores: make(map[string]int32),
    //         Lamps:      0,
    //         Tickets:    0,
    //     }

    //     if err := profileRepo.UpsertProfile(ctx, profile); err != nil {
    //         log.Printf("Failed to upsert profile for user %s: %v", event.UserID, err)
    //     } else {
    //         log.Printf("✅ Profile created for user: %s", event.UserID)
    //     }

    //     msg.Ack()
    // }, nats.Durable("profile-user-registered"), nats.ManualAck())

    // if err != nil {
    //     log.Printf("Warning: failed to subscribe to user.registered: %v", err)
    // }

	// Создаём или обновляем Stream для событий
	_, err = js.AddStream(&nats.StreamConfig{
		Name:     "EVENTS",
		Subjects: []string{"event.>", "score.updated", "user.registered"},
		Storage:  nats.FileStorage,
	})
	if err != nil {
		// Если Stream уже существует, пытаемся обновить его
		if err.Error() == "nats: stream name already in use" {
			log.Println("✅ Stream EVENTS already exists, updating...")
			
			// Получаем текущую информацию о Stream
			streamInfo, _ := js.StreamInfo("EVENTS")
			if streamInfo != nil {
				// Добавляем user.registered к существующим subject'ам
				newSubjects := streamInfo.Config.Subjects
				if !contains(newSubjects, "user.registered") {
					newSubjects = append(newSubjects, "user.registered")
					
					_, err = js.UpdateStream(&nats.StreamConfig{
						Name:     "EVENTS",
						Subjects: newSubjects,
						Storage:  nats.FileStorage,
					})
					if err != nil {
						log.Printf("❌ Failed to update stream: %v", err)
					} else {
						log.Println("✅ Stream EVENTS updated with user.registered")
					}
				}
			}
		} else {
			log.Printf("❌ Failed to create stream: %v", err)
		}
	}

    // Подписка на NATS (обновление рекордов)
    _, err = js.Subscribe("score.updated", func(msg *nats.Msg) {
        var event ScoreEvent
        if err := json.Unmarshal(msg.Data, &event); err != nil {
            log.Printf("Failed to unmarshal score.updated: %v", err)
            return
        }

        log.Printf("📡 Updating profile for user %s (game: %s, score: %d)", event.UserID, event.GameID, event.Score)

        // Получаем текущий профиль
        profile, err := profileRepo.GetProfile(ctx, event.UserID)
        if err != nil || profile == nil {
            profile = &repository.UserProfile{
                UserID:     event.UserID,
                BestScores: make(map[string]int32),
            }
        }

        // Обновляем best_scores
        if event.IsRecord {
            if profile.BestScores == nil {
                profile.BestScores = make(map[string]int32)
            }
            if current, ok := profile.BestScores[event.GameID]; !ok || int32(event.Score) > current {
                profile.BestScores[event.GameID] = int32(event.Score)
            }
        }

        // Обновляем общий счёт (суммируем все рекорды)
        var total int32
        for _, s := range profile.BestScores {
            total += s
        }
        profile.TotalScore = total

        // Обновляем лампочки и билетики
        profile.Lamps += int32(event.LampsEarned)
        profile.Tickets += int32(event.TicketsEarned)

        if err := profileRepo.UpsertProfile(ctx, profile); err != nil {
            log.Printf("Failed to update profile for user %s: %v", event.UserID, err)
        } else {
            log.Printf("✅ Profile updated for user %s (total: %d)", event.UserID, profile.TotalScore)
        }

        msg.Ack()
    }, nats.Durable("profile-score-updated"), nats.ManualAck())

    if err != nil {
        log.Printf("Warning: failed to subscribe to score.updated: %v", err)
    }

    // Запускаем gRPC сервер
    lis, err := net.Listen("tcp", ":"+cfg.GRPCPort)
    if err != nil {
        log.Fatalf("Failed to listen: %v", err)
    }

    go func() {
        log.Printf("📊 Profile service listening on :%s", cfg.GRPCPort)
        if err := grpcServer.Serve(lis); err != nil {
            log.Fatalf("Failed to serve: %v", err)
        }
    }()

    // Graceful shutdown
    quit := make(chan os.Signal, 1)
    signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
    <-quit

    log.Println("Shutting down profile service gracefully...")
    grpcServer.GracefulStop()
    nc.Drain()

    log.Println("Profile service stopped gracefully")
}
