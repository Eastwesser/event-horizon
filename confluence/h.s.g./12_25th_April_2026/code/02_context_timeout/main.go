// Context timeout: работа обрывается по ctx.Done().
package main

import (
	"context"
	"fmt"
	"time"
)

func work(ctx context.Context) error {
	for i := 0; i < 10; i++ {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-time.After(50 * time.Millisecond):
			fmt.Println("step", i)
		}
	}
	return nil
}

func main() {
	ctx, cancel := context.WithTimeout(context.Background(), 120*time.Millisecond)
	defer cancel()
	if err := work(ctx); err != nil {
		fmt.Println("stopped:", err)
	}
}
