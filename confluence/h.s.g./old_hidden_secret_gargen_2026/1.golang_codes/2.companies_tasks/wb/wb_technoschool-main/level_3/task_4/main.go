package main

import (
	"context"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/wb-go/wbf/ginext"
	"github.com/wb-go/wbf/zlog"
	"github.com/wb_technoschool/level_3/task_4/config"
	"github.com/wb_technoschool/level_3/task_4/handlers"
	"github.com/wb_technoschool/level_3/task_4/processor"
	"github.com/wb_technoschool/level_3/task_4/queue"
	"github.com/wb_technoschool/level_3/task_4/storage"
	"github.com/wb_technoschool/level_3/task_4/worker"
)

func main() {
	zlog.InitConsole()
	zlog.Logger.Info().Msg("Starting ImageProcessor service...")

	cfg := config.Load()

	store := storage.NewInMemoryStorage()
	zlog.Logger.Info().Msg("Storage initialized")

	fileStore, err := storage.NewFileStorage(cfg.StoragePath, store)
	if err != nil {
		zlog.Logger.Fatal().Err(err).Msg("Failed to initialize file storage")
	}
	zlog.Logger.Info().Str("path", cfg.StoragePath).Msg("File storage initialized")

	var queueInstance queue.Queue
	kafkaQueue, err := queue.NewKafkaQueue(cfg.KafkaBrokers, cfg.KafkaTopic, cfg.KafkaGroupID)
	if err != nil {
		zlog.Logger.Warn().Err(err).Msg("Failed to connect to Kafka, using in-memory queue")
		queueInstance = queue.NewInMemoryQueue()
	} else {
		zlog.Logger.Info().Msg("Kafka connected")
		queueInstance = kafkaQueue
		defer queueInstance.Close()
	}

	imageProcessor := processor.NewProcessor()
	workerInstance := worker.NewWorker(store, fileStore, queueInstance, imageProcessor)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	go func() {
		if err := workerInstance.Start(ctx); err != nil {
			zlog.Logger.Error().Err(err).Msg("Worker error")
		}
	}()
	zlog.Logger.Info().Msg("Background worker started")

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
		c.Writer.Header().Set("Access-Control-Allow-Methods", "GET, POST, DELETE, OPTIONS")
		c.Writer.Header().Set("Access-Control-Allow-Headers", "Content-Type")
		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(204)
			return
		}
		c.Next()
	})

	handler := handlers.NewHandler(store, fileStore, queueInstance)

	router.POST("/upload", handler.Upload)
	router.GET("/image/:id", handler.GetImage)
	router.GET("/image/:id/file", handler.GetImageFile)
	router.DELETE("/image/:id", handler.DeleteImage)
	router.GET("/images", handler.GetAllImages)

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

