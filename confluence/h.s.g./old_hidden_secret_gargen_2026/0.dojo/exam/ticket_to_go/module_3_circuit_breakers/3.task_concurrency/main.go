package main

import (
	"context"
	"fmt"
	"time"
)

/*
Graceful shutdown с context.

На собесе / в сервисе:
  ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
  defer stop()
  // сервер.Shutdown(ctx) / consumer.Stop()

Здесь имитируем сигнал через cancel через 200ms — тот же контракт Done().
*/
func work(ctx context.Context) {
	t := time.NewTicker(50 * time.Millisecond)
	defer t.Stop()
	for {
		select {
		case <-ctx.Done():
			fmt.Println("shutting down:", ctx.Err())
			return
		case <-t.C:
			fmt.Println("tick")
		}
	}
}

func main() {
	ctx, cancel := context.WithCancel(context.Background())

	go func() {
		time.Sleep(200 * time.Millisecond)
		fmt.Println("<< signal >>")
		cancel()
	}()

	work(ctx)
	fmt.Println("cleanup done")
}
