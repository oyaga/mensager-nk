# Makefile for Chatwoot-Go

.PHONY: help dev build test clean docker-up docker-down

help: ## Show this help
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | sort | awk 'BEGIN {FS = ":.*?## "}; {printf "\033[36m%-20s\033[0m %s\n", $$1, $$2}'

dev: ## Start development environment
	docker-compose up -d

dev-backend: ## Run backend in development mode
	cd backend && air

dev-frontend: ## Run frontend in development mode
	cd frontend && npm run dev

build: ## Build production images
	docker-compose -f docker-compose.yml build

build-backend: ## Build backend binary
	cd backend && CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o main ./cmd/server

build-frontend: ## Build frontend for production
	cd frontend && npm run build

test: ## Run tests
	cd backend && go test ./...

test-coverage: ## Run tests with coverage
	cd backend && go test -coverprofile=coverage.out ./...
	cd backend && go tool cover -html=coverage.out

lint: ## Run linters
	cd backend && golangci-lint run
	cd frontend && npm run lint

clean: ## Clean build artifacts
	rm -rf backend/tmp backend/main
	rm -rf frontend/dist frontend/node_modules

docker-up: ## Start all services with Docker Compose
	docker-compose up -d

docker-down: ## Stop all services
	docker-compose down

docker-logs: ## Show Docker logs
	docker-compose logs -f

db-migrate: ## Run database migrations
	@echo "Migrations run automatically on startup"

db-reset: ## Reset database (WARNING: destroys all data)
	docker-compose down -v
	docker-compose up -d postgres
	@echo "Database reset complete"

.DEFAULT_GOAL := help
