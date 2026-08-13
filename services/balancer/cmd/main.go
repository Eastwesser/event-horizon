package main

import (
	"context"
	"log"
	"os"

	"github.com/Eastwesser/event-horizon/services/balancer/internal/app"
)

func main() {
	ctx := context.Background()
	application, err := app.New(ctx)
	if err != nil {
		log.Fatalf("failed to init balancer app: %v", err)
	}
	if err := application.RunUntilSignal(ctx); err != nil {
		log.Printf("balancer stopped with error: %v", err)
		os.Exit(1)
	}
}
