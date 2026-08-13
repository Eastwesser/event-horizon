package main

import (
	"context"
	"log"
	"os"

	"github.com/Eastwesser/event-horizon/services/leaderboard/internal/app"
)

func main() {
	ctx := context.Background()
	application, err := app.New(ctx)
	if err != nil {
		log.Fatalf("failed to init leaderboard app: %v", err)
	}

	if err := application.RunUntilSignal(ctx); err != nil {
		log.Printf("leaderboard stopped with error: %v", err)
		os.Exit(1)
	}
}
