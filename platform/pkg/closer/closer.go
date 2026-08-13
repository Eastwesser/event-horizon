package closer

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"
)

const defaultShutdownTimeout = 10 * time.Second

// Closer runs registered cleanup funcs in reverse order (LIFO) on shutdown.
type Closer struct {
	mu     sync.Mutex
	once   sync.Once
	funcs  []namedFunc
	logger *slog.Logger
}

type namedFunc struct {
	name string
	fn   func(context.Context) error
}

func New(logger *slog.Logger) *Closer {
	if logger == nil {
		logger = slog.Default()
	}
	return &Closer{logger: logger}
}

// Add registers an unnamed cleanup function.
func (c *Closer) Add(fn func(context.Context) error) {
	c.AddNamed("", fn)
}

// AddNamed registers a cleanup function with a label for logs.
func (c *Closer) AddNamed(name string, fn func(context.Context) error) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.funcs = append(c.funcs, namedFunc{name: name, fn: fn})
}

// WaitSignal blocks until SIGINT/SIGTERM, then CloseAll with timeout.
func (c *Closer) WaitSignal(parent context.Context) error {
	ch := make(chan os.Signal, 1)
	signal.Notify(ch, syscall.SIGINT, syscall.SIGTERM)
	defer signal.Stop(ch)

	select {
	case sig := <-ch:
		c.logger.Info("shutdown signal received", "signal", sig.String())
	case <-parent.Done():
		c.logger.Info("parent context cancelled, shutting down")
	}

	ctx, cancel := context.WithTimeout(context.Background(), defaultShutdownTimeout)
	defer cancel()
	return c.CloseAll(ctx)
}

// CloseAll runs cleanup funcs once, newest first.
func (c *Closer) CloseAll(ctx context.Context) error {
	var result error
	c.once.Do(func() {
		c.mu.Lock()
		funcs := c.funcs
		c.funcs = nil
		c.mu.Unlock()

		for i := len(funcs) - 1; i >= 0; i-- {
			f := funcs[i]
			name := f.name
			if name == "" {
				name = "resource"
			}
			start := time.Now()
			if err := f.fn(ctx); err != nil {
				c.logger.Error("close failed", "name", name, "err", err, "took", time.Since(start))
				result = errors.Join(result, fmt.Errorf("%s: %w", name, err))
				continue
			}
			c.logger.Info("closed", "name", name, "took", time.Since(start))
		}
	})
	return result
}
