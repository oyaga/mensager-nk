package main

import (
	"log"
	"os"
	"path/filepath"
	"strings"

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

	// CORS middleware - Inline (como evolution-go)
	router.Use(func(c *gin.Context) {
		c.Writer.Header().Set("Access-Control-Allow-Origin", "*")
		c.Writer.Header().Set("Access-Control-Allow-Credentials", "true")
		c.Writer.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, PATCH, DELETE, OPTIONS")
		c.Writer.Header().Set("Access-Control-Allow-Headers", "Origin, Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token, Authorization, Accept, Cache-Control, X-Requested-With")
		c.Writer.Header().Set("Access-Control-Expose-Headers", "Content-Length")

		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(200)
			return
		}

		c.Next()
	})

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

	// Setup API routes (ANTES das rotas estáticas)
	routes.SetupRoutes(router, db, redisClient, wsHub, minioService, cfg)

	// Serve static frontend files (SPA) - Padrão Evolution-Go
	distPath := "./dist"
	log.Println("📦 Configuring static files from ./dist")

	// Verificar se o diretório dist existe
	if _, err := os.Stat(distPath); os.IsNotExist(err) {
		log.Printf("⚠️  WARNING: dist directory not found at %s", distPath)
		log.Printf("⚠️  Frontend will NOT be available. Please build frontend first.")
	} else {
		log.Printf("✅ Found dist directory at %s", distPath)

		// Verificar se index.html existe
		indexPath := filepath.Join(distPath, "index.html")
		if _, err := os.Stat(indexPath); os.IsNotExist(err) {
			log.Printf("⚠️  WARNING: index.html not found at %s", indexPath)
		} else {
			log.Printf("✅ Found index.html - Frontend ready to serve")
		}
	}

	// Rota raiz explícita
	router.GET("/", func(c *gin.Context) {
		indexPath := filepath.Join(distPath, "index.html")
		if _, err := os.Stat(indexPath); os.IsNotExist(err) {
			c.JSON(500, gin.H{
				"error": "Frontend not found",
				"message": "The frontend build files are missing. Please rebuild the Docker image.",
			})
			return
		}
		c.File(indexPath)
	})

	// Arquivos estáticos específicos
	router.GET("/assets/*filepath", func(c *gin.Context) {
		fp := c.Param("filepath")
		fullPath := filepath.Join(distPath, "assets", fp)

		if _, err := os.Stat(fullPath); os.IsNotExist(err) {
			log.Printf("⚠️  Asset not found: %s", fullPath)
			c.JSON(404, gin.H{"error": "Asset not found"})
			return
		}
		c.File(fullPath)
	})

	router.GET("/favicon.ico", func(c *gin.Context) {
		faviconPath := filepath.Join(distPath, "favicon.ico")
		if _, err := os.Stat(faviconPath); os.IsNotExist(err) {
			c.Status(404)
			return
		}
		c.File(faviconPath)
	})

	router.GET("/vite.svg", func(c *gin.Context) {
		vitePath := filepath.Join(distPath, "vite.svg")
		if _, err := os.Stat(vitePath); os.IsNotExist(err) {
			c.Status(404)
			return
		}
		c.File(vitePath)
	})

	// SPA Fallback (NoRoute) - Para rotas do React Router
	router.NoRoute(func(c *gin.Context) {
		path := c.Request.URL.Path

		// Não serve index.html para API ou WebSocket
		if strings.HasPrefix(path, "/api/") || strings.HasPrefix(path, "/cable") {
			c.JSON(404, gin.H{"error": "Not found"})
			return
		}

		// Fallback para index.html (SPA routing)
		c.File(filepath.Join(distPath, "index.html"))
	})

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
