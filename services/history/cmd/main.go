package main

import (
	"context"
	"log"
	"os"

	"github.com/Eastwesser/event-horizon/services/history/internal/app"
)

func main() {
	ctx := context.Background()
	application, err := app.New(ctx)
	if err != nil {
		log.Fatalf("failed to init history app: %v", err)
	}
	if err := application.RunUntilSignal(ctx); err != nil {
		log.Printf("history stopped: %v", err)
		os.Exit(1)
	}
}
