package main

import (
	"context"
	"log"
	"os"

	"github.com/Eastwesser/event-horizon/services/nats-hub/internal/app"
)

func main() {
	ctx := context.Background()
	application, err := app.New(ctx)
	if err != nil {
		log.Fatalf("failed to init nats-hub app: %v", err)
	}
	if err := application.RunUntilSignal(ctx); err != nil {
		log.Printf("nats-hub stopped with error: %v", err)
		os.Exit(1)
	}
}
