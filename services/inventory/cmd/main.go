package main

import (
	"context"
	"log"
	"os"

	"github.com/Eastwesser/event-horizon/services/inventory/internal/app"
)

func main() {
	ctx := context.Background()
	application, err := app.New(ctx)
	if err != nil {
		log.Fatalf("failed to init inventory app: %v", err)
	}

	if err := application.RunUntilSignal(ctx); err != nil {
		log.Printf("inventory stopped with error: %v", err)
		os.Exit(1)
	}
}
