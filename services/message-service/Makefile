.PHONY: help build run dev docker-build docker-run test clean deps

help: ## Mostra ajuda
	@echo "Comandos disponíveis:"
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | awk 'BEGIN {FS = ":.*?## "}; {printf "  \033[36m%-15s\033[0m %s\n", $$1, $$2}'

deps: ## Instala dependências
	go mod download
	go mod tidy

build: ## Build do binário
	CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build \
		-ldflags="-s -w -X main.Version=1.0.0" \
		-trimpath \
		-o message-service \
		./cmd/main.go
	@ls -lh message-service

run: ## Executa o serviço localmente
	go run cmd/main.go

dev: ## Executa com hot reload (necessita air)
	air

docker-build: ## Build da imagem Docker (scratch)
	docker build -t message-service:latest .
	@docker images message-service:latest

docker-build-alpine: ## Build da imagem Docker (alpine)
	docker build --target runtime-alpine -t message-service:alpine .
	@docker images message-service:alpine

docker-run: ## Executa container Docker
	docker run -p 3001:3001 --env-file .env message-service:latest

docker-compose-up: ## Inicia com docker-compose
	docker-compose up -d

docker-compose-down: ## Para docker-compose
	docker-compose down

docker-compose-logs: ## Mostra logs do docker-compose
	docker-compose logs -f

test: ## Executa testes
	go test -v ./...

test-coverage: ## Executa testes com coverage
	go test -v -coverprofile=coverage.out ./...
	go tool cover -html=coverage.out

bench: ## Executa benchmarks
	go test -bench=. -benchmem ./...

lint: ## Executa linter
	golangci-lint run

clean: ## Limpa arquivos gerados
	rm -f message-service
	rm -f coverage.out
	docker-compose down -v

size: ## Mostra tamanho do binário
	@echo "Tamanho do binário:"
	@ls -lh message-service 2>/dev/null || echo "Execute 'make build' primeiro"
	@echo "\nTamanho das imagens Docker:"
	@docker images message-service --format "table {{.Repository}}\t{{.Tag}}\t{{.Size}}"

health: ## Verifica health do serviço
	@curl -s http://localhost:3001/health | jq .

.DEFAULT_GOAL := help
