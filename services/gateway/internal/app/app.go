package app

import (
	"context"

	"github.com/Eastwesser/event-horizon/platform/pkg/closer"
	"github.com/Eastwesser/event-horizon/platform/pkg/logger"
	"log/slog"
)

// App is the API gateway process. HTTP route wiring lives in gateway.go
// (extracted from the former cmd/main monolith). New/RunUntilSignal match other services.
type App struct {
	log    *slog.Logger
	closer *closer.Closer
}

// New prepares logging/closer; servers start in RunUntilSignal via runGateway.
func New(ctx context.Context) (*App, error) {
	_ = ctx
	log := logger.NewFrom("info", "text")
	return &App{
		log:    log,
		closer: closer.New(log),
	}, nil
}

// RunUntilSignal blocks in the gateway HTTP server until process signal.
func (a *App) RunUntilSignal(ctx context.Context) error {
	_ = ctx
	a.log.Info("starting gateway bootstrap")
	runGateway()
	return nil
}

// Close is mostly a no-op; runGateway owns its shutdown path today.
func (a *App) Close(ctx context.Context) error {
	if a.closer == nil {
		return nil
	}
	return a.closer.CloseAll(ctx)
}
