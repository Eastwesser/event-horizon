package main

import (
	"context"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/wb-go/wbf/ginext"
	"github.com/wb-go/wbf/zlog"
	"github.com/wb_technoschool/level_3/task_5/config"
	"github.com/wb_technoschool/level_3/task_5/handlers"
	"github.com/wb_technoschool/level_3/task_5/scheduler"
	"github.com/wb_technoschool/level_3/task_5/service"
	"github.com/wb_technoschool/level_3/task_5/storage"
)

func main() {
	zlog.InitConsole()
	zlog.Logger.Info().Msg("Starting EventBooker service...")

	cfg := config.Load()

	zlog.Logger.Info().
		Str("database_url", cfg.DatabaseURL).
		Str("port", cfg.Port).
		Int("timeout", cfg.DefaultTimeoutMin).
		Msg("Loaded configuration")

	store, err := storage.NewPostgresStorage(cfg.DatabaseURL)
	if err != nil {
		zlog.Logger.Fatal().Err(err).Msg("Failed to connect to database")
	}
	defer store.Close()

	zlog.Logger.Info().Msg("Database connected")

	bookingService := service.NewBookingService(store, cfg.DefaultTimeoutMin)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	bookingScheduler := scheduler.NewBookingScheduler(store, 30*time.Second)
	go bookingScheduler.Start(ctx)

	router := ginext.New("")
	router.SetTrustedProxies(nil)

	router.Use(func(c *ginext.Context) {
		start := time.Now()
		path := c.Request.URL.Path
		raw := c.Request.URL.RawQuery

		c.Next()

		if raw != "" {
			path = path + "?" + raw
		}

		zlog.Logger.Info().
			Str("method", c.Request.Method).
			Str("path", path).
			Int("status", c.Writer.Status()).
			Dur("latency", time.Since(start)).
			Str("ip", c.ClientIP()).
			Msg("HTTP Request")
	})

	router.Use(func(c *ginext.Context) {
		c.Writer.Header().Set("Access-Control-Allow-Origin", "*")
		c.Writer.Header().Set("Access-Control-Allow-Methods", "GET, POST, OPTIONS")
		c.Writer.Header().Set("Access-Control-Allow-Headers", "Content-Type")
		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(204)
			return
		}
		c.Next()
	})

	handler := handlers.NewHandler(bookingService)

	router.POST("/events", handler.CreateEvent)
	router.GET("/events", handler.GetAllEvents)
	router.GET("/events/:id", handler.GetEvent)
	router.POST("/events/:id/book", handler.CreateBooking)
	router.POST("/bookings/:id/confirm", handler.ConfirmBooking)
	router.GET("/events/:id/bookings", handler.GetEventBookings)
	router.GET("/bookings", handler.GetUserBookings)

	router.StaticFile("/", "./static/index.html")
	router.StaticFile("/admin", "./static/admin.html")
	router.Static("/static", "./static")

	go func() {
		addr := ":" + cfg.Port
		zlog.Logger.Info().Str("addr", addr).Msg("Server starting")
		if err := router.Run(addr); err != nil {
			zlog.Logger.Fatal().Err(err).Msg("Failed to start server")
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	zlog.Logger.Info().Msg("Shutting down server...")
	cancel()
	time.Sleep(1 * time.Second)
	zlog.Logger.Info().Msg("Server stopped")
}
