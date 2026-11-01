package main

import (
	"context"
	"log"
	"net/http"
    _ "net/http/pprof" // <-- регистрирует /debug/pprof/
	"os"

	"github.com/Eastwesser/event-horizon/services/gateway/internal/app"
)

func main() {
	ctx := context.Background()
	application, err := app.New(ctx)
	if err != nil {
		log.Fatalf("failed to init gateway app: %v", err)
	}

	// --- PPROF SERVER ---
	go func() {
		log.Println("pprof server listening on :6060")
		if err := http.ListenAndServe(":6060", nil); err != nil {
			log.Printf("pprof server error: %v", err)
		}
	}()

	if err := application.RunUntilSignal(ctx); err != nil {
		log.Printf("gateway stopped with error: %v", err)
		os.Exit(1)
	}
}
