package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"

	"github.com/gin-gonic/gin"
	"github.com/wb-go/wb_technoschool/level_3/task_6/config"
	"github.com/wb-go/wb_technoschool/level_3/task_6/handlers"
	"github.com/wb-go/wb_technoschool/level_3/task_6/service"
	"github.com/wb-go/wb_technoschool/level_3/task_6/storage"
	"github.com/wb-go/wbf/dbpg"
	"github.com/wb-go/wbf/zlog"
)

func main() {
	// Initialize logger
	zlog.Init()
	logger := zlog.Logger

	// Load config
	cfg, err := config.Load()
	if err != nil {
		logger.Fatal().Err(err).Msg("Failed to load config")
	}

	// Get database DSN
	dsn := cfg.Database.DSN
	if dsn == "" {
		dsn = os.Getenv("DATABASE_DSN")
	}
	if dsn == "" {
		dsn = "postgres://postgres:8832@localhost:5442/sales_tracker?sslmode=disable"
	}

	logger.Info().Str("dsn", dsn).Msg("Connecting to database")

	// Create database connection using WBF
	db, err := dbpg.New(dsn, []string{}, &dbpg.Options{
		MaxOpenConns: 10,
		MaxIdleConns: 5,
	})
	if err != nil {
		logger.Fatal().Err(err).Msg("Failed to connect to database")
	}
	defer func() {
		if db.Master != nil {
			_ = db.Master.Close()
		}
		for _, slave := range db.Slaves {
			if slave != nil {
				_ = slave.Close()
			}
		}
	}()

	// Initialize storage
	st := storage.New(db)
	if err := st.InitSchema(context.Background()); err != nil {
		logger.Fatal().Err(err).Msg("Failed to initialize schema")
	}

	// Initialize service and handlers
	svc := service.New(st)
	h := handlers.New(svc)

	// Setup Gin router
	gin.SetMode(gin.ReleaseMode)
	r := gin.Default()

	// CORS middleware
	r.Use(corsMiddleware())

	// Serve static files
	r.Static("/static", "./static")
	r.StaticFile("/", "./static/index.html")

	// API routes
	api := r.Group("/api")
	{
		// CRUD operations
		api.POST("/items", h.CreateTransaction)
		api.GET("/items", h.ListTransactions)
		api.GET("/items/:id", h.GetTransaction)
		api.PUT("/items/:id", h.UpdateTransaction)
		api.DELETE("/items/:id", h.DeleteTransaction)

		// Analytics
		api.GET("/analytics", h.GetAnalytics)
		api.GET("/analytics/categories", h.GetCategoryAnalytics)

		// Export
		api.GET("/export/csv", h.ExportCSV)
	}

	// Health check
	r.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "ok"})
	})

	// Start server
	addr := fmt.Sprintf("%s:%d", cfg.Server.Host, cfg.Server.Port)
	logger.Info().Str("addr", addr).Msg("Starting server")
	log.Printf("Server running at http://%s\n", addr)
	log.Printf("Open http://%s in your browser\n", addr)

	if err := r.Run(addr); err != nil {
		logger.Fatal().Err(err).Msg("Server failed")
	}
}

func corsMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Writer.Header().Set("Access-Control-Allow-Origin", "*")
		c.Writer.Header().Set("Access-Control-Allow-Credentials", "true")
		c.Writer.Header().Set("Access-Control-Allow-Headers", "Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token, Authorization, accept, origin, Cache-Control, X-Requested-With")
		c.Writer.Header().Set("Access-Control-Allow-Methods", "POST, OPTIONS, GET, PUT, DELETE")

		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(204)
			return
		}

		c.Next()
	}
}

