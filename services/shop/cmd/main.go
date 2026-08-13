package main

import (
	"context"
	"log"
	"os"

	"github.com/Eastwesser/event-horizon/services/shop/internal/app"
)

func main() {
	ctx := context.Background()
	application, err := app.New(ctx)
	if err != nil {
		log.Fatalf("failed to init shop app: %v", err)
	}

	if err := application.RunUntilSignal(ctx); err != nil {
		log.Printf("shop stopped with error: %v", err)
		os.Exit(1)
	}
}
