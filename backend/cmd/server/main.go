package main

import (
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
	"github.com/nakamura/chatwoot-go/internal/config"
	"github.com/nakamura/chatwoot-go/internal/database"
	"github.com/nakamura/chatwoot-go/internal/middleware"
	"github.com/nakamura/chatwoot-go/internal/routes"
	"github.com/nakamura/chatwoot-go/internal/storage"
	"github.com/nakamura/chatwoot-go/internal/websocket"
)

func main() {
	// Load environment variables
	if err := godotenv.Load(); err != nil {
		log.Println("No .env file found, using system environment variables")
	}

	// Initialize configuration
	cfg := config.New()

	// Initialize database
	db, err := database.NewPostgresDB(cfg.DatabaseURL)
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}

	// Run migrations
	if err := database.RunMigrations(db); err != nil {
		log.Fatalf("Failed to run migrations: %v", err)
	}

	// Initialize Redis
	redisClient := database.NewRedisClient(cfg.RedisURL)

	// Initialize WebSocket hub
	wsHub := websocket.NewHub()
	go wsHub.Run()

	// Initialize Minio
	minioService, err := storage.NewMinioService(cfg)
	if err != nil {
		// Log warning but continue; upload feature will fail gracefully if used
		log.Printf("⚠️ Warning: Failed to initialize Minio: %v. Uploads will not work.", err)
	} else {
		log.Println("✅ Minio storage initialized")
	}

	// Setup Gin router
	if cfg.GoEnv == "production" {
		gin.SetMode(gin.ReleaseMode)
	}

	router := gin.Default()

	// CORS middleware - Permissive for MVP/Production ease
	router.Use(cors.New(cors.Config{
		AllowAllOrigins:  true,
		AllowMethods:     []string{"GET", "POST", "PUT", "PATCH", "DELETE", "OPTIONS"},
		AllowHeaders:     []string{"Origin", "Content-Type", "Authorization", "Accept"},
		ExposeHeaders:    []string{"Content-Length"},
		AllowCredentials: true,
	}))

	// Request logging middleware
	router.Use(middleware.Logger())

	// Recovery middleware
	router.Use(gin.Recovery())

	// Health check
	router.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{
			"status":  "ok",
			"service": "chatwoot-go",
		})
	})

	// Setup routes
	routes.SetupRoutes(router, db, redisClient, wsHub, minioService, cfg)

	// Serve static frontend files (SPA)
	distPath := "./dist"
	if _, err := os.Stat(distPath); err == nil {
		log.Println("📦 Serving frontend from ./dist")

		// Serve static assets (js, css, images, etc.)
		router.Static("/assets", filepath.Join(distPath, "assets"))

		// Serve other static files from dist root
		router.StaticFile("/favicon.ico", filepath.Join(distPath, "favicon.ico"))
		router.StaticFile("/manifest.json", filepath.Join(distPath, "manifest.json"))

		// SPA fallback: serve index.html for all non-API routes
		router.NoRoute(func(c *gin.Context) {
			path := c.Request.URL.Path

			// Don't serve index.html for API routes or WebSocket
			if strings.HasPrefix(path, "/api/") || strings.HasPrefix(path, "/cable") {
				c.JSON(http.StatusNotFound, gin.H{"error": "Not found"})
				return
			}

			// Serve index.html for SPA routing
			c.File(filepath.Join(distPath, "index.html"))
		})
	} else {
		log.Println("⚠️ No ./dist folder found - frontend will not be served")
	}

	// Start server
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	log.Printf("🚀 Server starting on port %s", port)
	if err := router.Run(":" + port); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}
