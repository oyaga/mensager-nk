package database

import (
	"context"
	"fmt"
	"log"
	"os"
	"strconv"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

// DB representa a conexão com o banco de dados
type DB struct {
	Pool *pgxpool.Pool
}

// NewPostgresConnection cria uma nova conexão otimizada com Supabase PostgreSQL
func NewPostgresConnection(databaseURL string) (*DB, error) {
	// Parse configurações do ambiente
	maxConns := getEnvAsInt("DB_MAX_CONNS", 20)
	minConns := getEnvAsInt("DB_MIN_CONNS", 5)
	maxConnLifetime := getEnvAsDuration("DB_MAX_CONN_LIFETIME", time.Hour)
	maxConnIdleTime := getEnvAsDuration("DB_MAX_CONN_IDLE_TIME", 30*time.Minute)

	// Configuração do pool de conexões
	config, err := pgxpool.ParseConfig(databaseURL)
	if err != nil {
		return nil, fmt.Errorf("unable to parse database URL: %w", err)
	}

	// Otimizações para Supabase
	config.MaxConns = int32(maxConns)
	config.MinConns = int32(minConns)
	config.MaxConnLifetime = maxConnLifetime
	config.MaxConnIdleTime = maxConnIdleTime
	config.HealthCheckPeriod = 1 * time.Minute

	// Configurações de conexão para melhor performance
	config.ConnConfig.RuntimeParams["application_name"] = "message-service"

	// Timeout de conexão
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// Criar pool de conexões
	pool, err := pgxpool.NewWithConfig(ctx, config)
	if err != nil {
		return nil, fmt.Errorf("unable to create connection pool: %w", err)
	}

	// Testar conexão
	if err := pool.Ping(ctx); err != nil {
		return nil, fmt.Errorf("unable to ping database: %w", err)
	}

	log.Printf("✅ Database connected successfully (Max: %d, Min: %d connections)", maxConns, minConns)

	return &DB{Pool: pool}, nil
}

// Close fecha a conexão com o banco de dados
func (db *DB) Close() {
	if db.Pool != nil {
		db.Pool.Close()
		log.Println("Database connection closed")
	}
}

// Health verifica o status da conexão
func (db *DB) Health(ctx context.Context) error {
	return db.Pool.Ping(ctx)
}

// Helper functions
func getEnvAsInt(key string, defaultVal int) int {
	if value := os.Getenv(key); value != "" {
		if intVal, err := strconv.Atoi(value); err == nil {
			return intVal
		}
	}
	return defaultVal
}

func getEnvAsDuration(key string, defaultVal time.Duration) time.Duration {
	if value := os.Getenv(key); value != "" {
		if duration, err := time.ParseDuration(value); err == nil {
			return duration
		}
	}
	return defaultVal
}
