package config

import (
	"os"
)

type Config struct {
	// Database
	DatabaseURL string
	RedisURL    string
	RabbitMQURL string

	// Storage
	MinioEndpoint  string
	MinioAccessKey string
	MinioSecretKey string
	MinioUseSSL    bool

	// JWT
	JWTSecret string

	// Server
	Port        string
	FrontendURL string
	GoEnv       string

	// Features
	EnableWebhooks bool
	EnableEmail    bool
}

func New() *Config {
	return &Config{
		// Database
		DatabaseURL: getEnv("DATABASE_URL", "postgresql://chatwoot:chatwoot123@localhost:5432/chatwoot_go?sslmode=disable"),
		RedisURL:    getEnv("REDIS_URL", "redis://localhost:6379"),
		RabbitMQURL: getEnv("RABBITMQ_URL", "amqp://chatwoot:chatwoot123@localhost:5672/"),

		// Storage
		MinioEndpoint:  getEnv("MINIO_ENDPOINT", "localhost:9000"),
		MinioAccessKey: getEnv("MINIO_ACCESS_KEY", "minioadmin"),
		MinioSecretKey: getEnv("MINIO_SECRET_KEY", "minioadmin123"),
		MinioUseSSL:    getEnv("MINIO_USE_SSL", "false") == "true",

		// JWT
		JWTSecret: getEnv("JWT_SECRET", "your-super-secret-jwt-key-change-in-production"),

		// Server
		Port:        getEnv("PORT", "8080"),
		FrontendURL: getEnv("FRONTEND_URL", "http://localhost:5173"),
		GoEnv:       getEnv("GO_ENV", "development"),

		// Features
		EnableWebhooks: getEnv("ENABLE_WEBHOOKS", "true") == "true",
		EnableEmail:    getEnv("ENABLE_EMAIL", "false") == "true",
	}
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}
