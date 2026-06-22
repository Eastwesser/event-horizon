package main

import (
    "log"
    "net/http"
    "os"
    "os/signal"
    "syscall"

    "github.com/Eastwesser/event-horizon/services/balancer/internal/balancer"
)

func main() {
    // Gateway (3 backend instances) - используем имена сервисов в Docker-сети
    backends := []string{
        "http://gateway:8080",
        "http://gateway-2:8080",
        "http://gateway-3:8080",
    }

    lb := balancer.NewLeastConnBalancer(backends)

    srv := &http.Server{
        Addr:    ":8079",
        Handler: lb,
    }

    go func() {
        log.Printf("⚖️ Load Balancer listening on :8079")
        log.Printf("   Backends: %v", backends)
        if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
            log.Fatalf("Failed to start: %v", err)
        }
    }()

    quit := make(chan os.Signal, 1)
    signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
    <-quit

    log.Println("🛑 LB Graceful shutdown")
    srv.Shutdown(nil)
}
