# 🚀 Message Service - Chatwoot Go

Microserviço ultra-otimizado em Go para processamento de mensagens, integrado com Supabase PostgreSQL.

## 📋 Características

- ✅ **Performance**: Fiber framework + pgx (conexão direta PostgreSQL)
- ✅ **Eficiência**: Binário estático de ~15-20MB (scratch) ou ~25MB (alpine)
- ✅ **Supabase**: Integração nativa com pool de conexões otimizado
- ✅ **Produção**: Docker multi-stage, health checks e graceful shutdown
- ✅ **Escalável**: Suporte a Prefork (multi-process) em produção

## 🏗️ Arquitetura

```
message-service/
├── cmd/
│   └── main.go                 # Entry point
├── internal/
│   ├── database/
│   │   └── postgres.go         # Conexão pgx com Supabase
│   ├── handlers/
│   │   └── message.go          # HTTP handlers
│   └── models/
│       └── message.go          # Estruturas de dados
├── Dockerfile                  # Multi-stage otimizado
├── docker-compose.yml
├── .env.example
└── README.md
```

## 🛠️ Stack Tecnológica

- **Go 1.23+** - Linguagem
- **Fiber v2** - Framework HTTP (mais rápido que Gin)
- **pgx v5** - Driver PostgreSQL (mais performático que GORM)
- **Supabase** - Backend as a Service (PostgreSQL)

## 🚀 Quick Start

### 1. Configurar Variáveis de Ambiente

```bash
cp .env.example .env
```

Edite o `.env` com suas credenciais do Supabase:

```env
SUPABASE_DB_URL=postgresql://postgres:[PASSWORD]@db.[PROJECT-REF].supabase.co:5432/postgres?sslmode=require
SUPABASE_KEY=your-supabase-anon-key
PORT=3001
GO_ENV=production
```

### 2. Desenvolvimento Local

```bash
# Instalar dependências
go mod download

# Executar
go run cmd/main.go
```

### 3. Docker (Produção)

```bash
# Build (imagem scratch - menor)
docker build -t message-service:latest .

# Ou Alpine (com shell para debug)
docker build --target runtime-alpine -t message-service:alpine .

# Executar
docker run -p 3001:3001 --env-file .env message-service:latest

# Ou com docker-compose
docker-compose up -d
```

## 📡 API Endpoints

### Health Check
```bash
GET /health
```

### Criar Mensagem
```bash
POST /api/v1/messages
Content-Type: application/json

{
  "conversation_id": "uuid",
  "content": "Olá, como posso ajudar?",
  "sender_id": "uuid",          # Opcional (agente)
  "contact_id": "uuid",          # Opcional (cliente)
  "message_type": "outgoing",    # incoming, outgoing, activity
  "content_type": "text",        # text, image, file
  "private": false,
  "source_id": "external-id"     # ID externo (WhatsApp, etc)
}
```

### Listar Mensagens de uma Conversa
```bash
GET /api/v1/conversations/:conversation_id/messages
```

### Buscar Mensagem Específica
```bash
GET /api/v1/messages/:id
```

## 🔧 Configuração do Pool de Conexões

Para Supabase, otimize conforme seu plano:

```env
# Free Tier: até 60 conexões
DB_MAX_CONNS=20
DB_MIN_CONNS=5

# Pro: até 200 conexões
DB_MAX_CONNS=50
DB_MIN_CONNS=10

# Enterprise: até 400+
DB_MAX_CONNS=100
DB_MIN_CONNS=20

DB_MAX_CONN_LIFETIME=1h
DB_MAX_CONN_IDLE_TIME=30m
```

## 📊 Performance

### Benchmarks (hardware típico)

| Métrica | Valor |
|---------|-------|
| Tamanho do binário (scratch) | ~15-20 MB |
| Tamanho do binário (alpine) | ~25 MB |
| Consumo de memória (idle) | ~10-15 MB |
| Consumo de memória (carga) | ~50-100 MB |
| Requests/segundo | 10.000+ |
| Latência média | < 5ms |

### Comparação vs Backend atual (Ruby)

| Aspecto | Ruby on Rails | Go Message Service |
|---------|---------------|-------------------|
| Imagem Docker | ~800MB | ~15-25MB |
| Boot time | ~10s | <100ms |
| Memory (idle) | ~200MB | ~15MB |
| Throughput | 1.000 req/s | 10.000+ req/s |

## 🐳 Opções de Imagem Docker

### Scratch (Padrão - Menor)
```bash
docker build -t message-service:scratch .
# Tamanho: ~15-20MB
# Prós: Tamanho mínimo, segurança máxima
# Contras: Sem shell, debug limitado
```

### Alpine (Debug)
```bash
docker build --target runtime-alpine -t message-service:alpine .
# Tamanho: ~25MB
# Prós: Shell disponível, ferramentas básicas
# Contras: ~5MB maior
```

### UPX Compression (Experimental)
Descomente a linha UPX no Dockerfile para comprimir ainda mais:
```dockerfile
RUN apk add --no-cache upx && upx --best --lzma message-service
# Pode reduzir em 50-70% adicional
# Tamanho final: ~5-10MB
```

## 🔐 Supabase Connection Types

### Session Pool (Porta 5432)
Recomendado para operações transacionais:
```env
SUPABASE_DB_URL=postgresql://postgres:[PASSWORD]@db.[PROJECT-REF].supabase.co:5432/postgres?sslmode=require
```

### Transaction Pool (Porta 6543)
Para queries rápidas e read-only:
```env
SUPABASE_DB_URL=postgres://postgres.[PROJECT-REF]:[PASSWORD]@aws-0-us-east-1.pooler.supabase.com:6543/postgres
```

## 📝 Schema do Banco de Dados

Este serviço utiliza a tabela `messages` já criada pelo backend principal:

```sql
-- Tabela gerenciada pelo backend/migrations
-- Apenas para referência
CREATE TABLE messages (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    conversation_id UUID NOT NULL REFERENCES conversations(id),
    sender_id UUID REFERENCES users(id),
    contact_id UUID REFERENCES contacts(id),
    content TEXT NOT NULL,
    message_type VARCHAR(50) DEFAULT 'incoming',
    content_type VARCHAR(50) DEFAULT 'text',
    private BOOLEAN DEFAULT false,
    status VARCHAR(50) DEFAULT 'sent',
    source_id VARCHAR(255),
    created_at TIMESTAMP DEFAULT NOW(),
    updated_at TIMESTAMP DEFAULT NOW()
);

CREATE INDEX idx_messages_conversation ON messages(conversation_id);
CREATE INDEX idx_messages_created_at ON messages(created_at);
```

## 🚦 Health Check

```bash
# Local
curl http://localhost:3001/health

# Docker
docker exec message-service wget -qO- http://localhost:3001/health
```

Resposta:
```json
{
  "status": "ok",
  "service": "message-service"
}
```

## 📈 Monitoramento

### Logs
```bash
# Docker
docker logs -f message-service

# Docker Compose
docker-compose logs -f message-service
```

### Métricas (TODO)
- Integração com Prometheus
- Grafana dashboards
- Distributed tracing

## 🔄 Deploy

### Docker Compose (Simples)
```bash
docker-compose up -d
```

### Kubernetes (TODO)
```bash
kubectl apply -f k8s/deployment.yaml
```

### Cloud Run / AWS Lambda (TODO)
Binário estático permite deploy serverless.

## 🧪 Testing

```bash
# Unit tests
go test ./...

# Load test com wrk
wrk -t4 -c100 -d30s http://localhost:3001/health

# Load test de inserção
wrk -t4 -c100 -d30s -s scripts/post.lua http://localhost:3001/api/v1/messages
```

## 🤝 Contribuindo

1. Fork o projeto
2. Crie uma branch para sua feature (`git checkout -b feature/nova-feature`)
3. Commit suas mudanças (`git commit -am 'Adiciona nova feature'`)
4. Push para a branch (`git push origin feature/nova-feature`)
5. Abra um Pull Request

## 📄 Licença

MIT License - veja [LICENSE](../../LICENSE)

## 🙏 Créditos

Parte do projeto [Chatwoot-Go](../../README.md)
