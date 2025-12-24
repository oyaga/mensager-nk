package main

import (
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/compress"
	"github.com/gofiber/fiber/v2/middleware/cors"
	"github.com/gofiber/fiber/v2/middleware/logger"
	"github.com/gofiber/fiber/v2/middleware/recover"
	"github.com/joho/godotenv"
	"github.com/nakamura/chatwoot-go/services/message-service/internal/database"
	"github.com/nakamura/chatwoot-go/services/message-service/internal/handlers"
)

func main() {
	// Carregar variáveis de ambiente
	if err := godotenv.Load(); err != nil {
		log.Println("No .env file found, using system environment variables")
	}

	// Obter configurações do ambiente
	databaseURL := os.Getenv("SUPABASE_DB_URL")
	if databaseURL == "" {
		log.Fatal("SUPABASE_DB_URL environment variable is required")
	}

	port := os.Getenv("PORT")
	if port == "" {
		port = "3001"
	}

	goEnv := os.Getenv("GO_ENV")
	if goEnv == "" {
		goEnv = "development"
	}

	// Conectar ao banco de dados
	db, err := database.NewPostgresConnection(databaseURL)
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}
	defer db.Close()

	// Inicializar Fiber com configurações otimizadas
	app := fiber.New(fiber.Config{
		AppName:               "Message Service v1.0",
		ServerHeader:          "Message-Service",
		DisableStartupMessage: goEnv == "production",
		Prefork:               goEnv == "production", // Multi-process para produção
		ReduceMemoryUsage:     true,
		CaseSensitive:         true,
		StrictRouting:         true,
		CompressedFileSuffix:  ".gz",
	})

	// Middlewares
	app.Use(recover.New())
	app.Use(logger.New(logger.Config{
		Format: "[${time}] ${status} - ${method} ${path} (${latency})\n",
	}))
	app.Use(cors.New(cors.Config{
		AllowOrigins: "*",
		AllowHeaders: "Origin, Content-Type, Accept, Authorization",
		AllowMethods: "GET, POST, PUT, DELETE, OPTIONS",
	}))
	app.Use(compress.New(compress.Config{
		Level: compress.LevelBestSpeed,
	}))

	// Inicializar handlers
	messageHandler := handlers.NewMessageHandler(db)

	// Health check
	app.Get("/health", func(c *fiber.Ctx) error {
		return c.JSON(fiber.Map{
			"status":  "ok",
			"service": "message-service",
		})
	})

	// Rotas da API
	api := app.Group("/api/v1")

	// Messages endpoints
	messages := api.Group("/messages")
	messages.Post("/", messageHandler.CreateMessage)
	messages.Get("/:id", messageHandler.GetMessage)

	// Conversations messages
	conversations := api.Group("/conversations")
	conversations.Get("/:conversation_id/messages", messageHandler.GetMessages)

	// Graceful shutdown
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM)

	go func() {
		<-c
		log.Println("Shutting down gracefully...")
		_ = app.Shutdown()
	}()

	// Start server
	log.Printf("🚀 Message Service starting on port %s (env: %s)", port, goEnv)
	if err := app.Listen(":" + port); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}
