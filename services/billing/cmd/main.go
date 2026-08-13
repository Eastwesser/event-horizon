package main

import (
	"context"
	"log"
	"os"

	"github.com/Eastwesser/event-horizon/services/billing/internal/app"
)

func main() {
	ctx := context.Background()
	application, err := app.New(ctx)
	if err != nil {
		log.Fatalf("failed to init billing app: %v", err)
	}

	if err := application.RunUntilSignal(ctx); err != nil {
		log.Printf("billing stopped with error: %v", err)
		os.Exit(1)
	}
}
