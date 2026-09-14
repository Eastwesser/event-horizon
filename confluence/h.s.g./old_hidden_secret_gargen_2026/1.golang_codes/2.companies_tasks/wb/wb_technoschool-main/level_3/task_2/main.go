package main

import (
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/wb-go/wbf/ginext"
	"github.com/wb-go/wbf/zlog"
	"github.com/wb_technoschool/level_3/task_2/cache"
	"github.com/wb_technoschool/level_3/task_2/config"
	"github.com/wb_technoschool/level_3/task_2/handlers"
	"github.com/wb_technoschool/level_3/task_2/service"
	"github.com/wb_technoschool/level_3/task_2/storage"
)

func main() {
	zlog.InitConsole()
	zlog.Logger.Info().Msg("Starting URL Shortener service...")

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

	baseURL := "http://localhost:" + cfg.Port
	if baseURLEnv := os.Getenv("BASE_URL"); baseURLEnv != "" {
		baseURL = baseURLEnv
	}

	shortenerService := service.NewShortenerService(store, cacheInstance, baseURL)

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
		c.Writer.Header().Set("Access-Control-Allow-Methods", "GET, POST, OPTIONS")
		c.Writer.Header().Set("Access-Control-Allow-Headers", "Content-Type")
		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(204)
			return
		}
		c.Next()
	})

	handler := handlers.NewHandler(shortenerService)

	router.POST("/shorten", handler.Shorten)
	router.GET("/s/:short_url", handler.Redirect)
	router.GET("/analytics/:short_url", handler.Analytics)

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
	time.Sleep(1 * time.Second)
	zlog.Logger.Info().Msg("Server stopped")
}
