package main

import (
	"context"
	"log"
	"os"

	"github.com/Eastwesser/event-horizon/services/auth/internal/app"
	"github.com/Eastwesser/event-horizon/services/auth/internal/config"
)

func main() {
	cfg := config.Load()
	if cfg.JWTSecret == "your-secret-key-change-me" {
		log.Println("WARNING: JWT_SECRET is using the insecure default — set JWT_SECRET before production")
	}

	ctx := context.Background()
	application, err := app.New(ctx)
	if err != nil {
		log.Fatalf("failed to init auth app: %v", err)
	}

	if err := application.RunUntilSignal(ctx); err != nil {
		log.Printf("auth stopped with error: %v", err)
		os.Exit(1)
	}
}
