package main

import (
	"context"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/wb-go/wbf/ginext"
	"github.com/wb-go/wbf/zlog"
	"github.com/wb_technoschool/level_3/task_1/cache"
	"github.com/wb_technoschool/level_3/task_1/config"
	"github.com/wb_technoschool/level_3/task_1/handlers"
	"github.com/wb_technoschool/level_3/task_1/notifiers"
	"github.com/wb_technoschool/level_3/task_1/queue"
	"github.com/wb_technoschool/level_3/task_1/service"
	"github.com/wb_technoschool/level_3/task_1/storage"
	"github.com/wb_technoschool/level_3/task_1/worker"
)

func main() {
	zlog.InitConsole()
	zlog.Logger.Info().Msg("Starting DelayedNotifier service...")

	cfg := config.Load()

	store := storage.NewInMemoryStorage()
	zlog.Logger.Info().Msg("Storage initialized")

	var cacheInstance cache.Cache
	redisCache, err := cache.NewRedisCache(cfg.RedisAddr, cfg.RedisPassword, cfg.RedisDB)
	if err != nil {
		zlog.Logger.Warn().Err(err).Msg("Failed to connect to Redis, using NoOp cache")
		cacheInstance = cache.NewNoOpCache()
	} else {
		zlog.Logger.Info().Msg("Redis cache connected")
		cacheInstance = redisCache
		defer redisCache.Close()
	}

	var queueInstance queue.Queue
	rabbitQueue, err := queue.NewRabbitMQQueue(cfg.RabbitMQURL)
	if err != nil {
		zlog.Logger.Warn().Err(err).Msg("Failed to connect to RabbitMQ, using in-memory queue")
		queueInstance = queue.NewInMemoryQueue()
	} else {
		zlog.Logger.Info().Msg("RabbitMQ connected")
		queueInstance = rabbitQueue
	}
	defer queueInstance.Close()

	notifierFactory := notifiers.NewNotifierFactory()
	
	notifierFactory.Register(notifiers.NewConsoleNotifier())
	zlog.Logger.Info().Msg("Console notifier registered")
	
	if cfg.SMTPUser != "" && cfg.SMTPPassword != "" {
		notifierFactory.Register(notifiers.NewEmailNotifier(
			cfg.SMTPHost, cfg.SMTPPort, cfg.SMTPUser, cfg.SMTPPassword, cfg.SMTPFrom,
		))
		zlog.Logger.Info().Msg("Email notifier registered")
	} else {
		zlog.Logger.Warn().Msg("Email notifier not configured")
	}
	
	if cfg.TelegramToken != "" {
		notifierFactory.Register(notifiers.NewTelegramNotifier(cfg.TelegramToken))
		zlog.Logger.Info().Msg("Telegram notifier registered")
	} else {
		zlog.Logger.Warn().Msg("Telegram notifier not configured")
	}

	notificationService := service.NewNotificationService(store, cacheInstance, queueInstance)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	workerInstance := worker.NewWorker(store, cacheInstance, queueInstance, notifierFactory)
	go func() {
		if err := workerInstance.Start(ctx); err != nil {
			zlog.Logger.Error().Err(err).Msg("Worker error")
		}
	}()
	zlog.Logger.Info().Msg("Background worker started")

	router := ginext.New("")

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
		c.Writer.Header().Set("Access-Control-Allow-Methods", "GET, POST, DELETE, OPTIONS")
		c.Writer.Header().Set("Access-Control-Allow-Headers", "Content-Type")
		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(204)
			return
		}
		c.Next()
	})

	handler := handlers.NewHandler(notificationService)
	
	router.POST("/notify", handler.CreateNotification)
	router.GET("/notify/:id", handler.GetNotification)
	router.DELETE("/notify/:id", handler.DeleteNotification)
	router.GET("/notify", handler.GetAllNotifications)

	router.StaticFile("/", "./static/index.html")
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

