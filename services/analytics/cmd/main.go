package main

import (
	"context"
	"log"
	"os"

	"github.com/Eastwesser/event-horizon/services/analytics/internal/app"
)

func main() {
	ctx := context.Background()
	application, err := app.New(ctx)
	if err != nil {
		log.Fatalf("failed to init analytics app: %v", err)
	}
	if err := application.RunUntilSignal(ctx); err != nil {
		log.Printf("analytics stopped: %v", err)
		os.Exit(1)
	}
}
