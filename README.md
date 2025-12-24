# 🚀 Chatwoot-Go

> Reimplementação moderna do Chatwoot usando **Go** (backend) e **React** (frontend)

## 📋 Visão Geral

Este projeto é uma reimplementação completa do Chatwoot, substituindo Ruby on Rails por Go e modernizando o frontend com React + TypeScript.

### 🎯 Objetivos

- ✅ **Performance**: Go oferece melhor performance e menor consumo de recursos
- ✅ **Escalabilidade**: Arquitetura modular e microserviços-ready
- ✅ **Modernidade**: Stack tecnológica atual e mantível
- ✅ **Compatibilidade**: Mantém compatibilidade com APIs existentes

## 🏗️ Arquitetura

```
chatwoot-go/
├── backend/          # API em Go
│   ├── cmd/          # Entry points
│   ├── internal/     # Código interno
│   ├── pkg/          # Pacotes reutilizáveis
│   └── api/          # Definições de API
├── frontend/         # React App
│   ├── src/
│   ├── public/
│   └── package.json
├── shared/           # Tipos compartilhados
├── docker/           # Configurações Docker
└── docs/             # Documentação
```

## 🛠️ Stack Tecnológica

### Backend

- **Go 1.21+** - Linguagem principal
- **Gin** - Framework web
- **GORM** - ORM
- **PostgreSQL** - Database
- **Redis** - Cache e sessions
- **WebSocket** - Comunicação real-time
- **RabbitMQ** - Message queue

### Frontend

- **React 18** - UI Framework
- **TypeScript** - Type safety
- **Vite** - Build tool
- **TailwindCSS** - Styling
- **React Query** - Data fetching
- **Zustand** - State management
- **Socket.io** - Real-time

## 🚀 Quick Start

### Pré-requisitos

- Go 1.21+
- Node.js 18+
- Docker & Docker Compose
- PostgreSQL 14+
- Redis 7+

### Desenvolvimento Local

```bash
# Clone o repositório
git clone <repo-url>
cd chatwoot-go

# Inicie os serviços com Docker
docker-compose up -d

# Backend
cd backend
go mod download
go run cmd/server/main.go

# Frontend (em outro terminal)
cd frontend
npm install
npm run dev
```

### Variáveis de Ambiente

Copie `.env.example` para `.env` e configure:

```env
# Database
DATABASE_URL=postgresql://user:pass@localhost:5432/chatwoot_go
REDIS_URL=redis://localhost:6379

# JWT
JWT_SECRET=your-secret-key

# Server
PORT=8080
FRONTEND_URL=http://localhost:5173
```

## 📚 Documentação

- [Arquitetura](./docs/architecture.md)
- [API Reference](./docs/api.md)
- [Guia de Desenvolvimento](./docs/development.md)
- [Deploy](./docs/deployment.md)

## 🔄 Migração do Chatwoot Original

Consulte [MIGRATION.md](./docs/MIGRATION.md) para guia de migração de dados.

## 🤝 Contribuindo

Contribuições são bem-vindas! Veja [CONTRIBUTING.md](./CONTRIBUTING.md).

## 📝 Licença

MIT License - veja [LICENSE](./LICENSE)

## 🙏 Créditos

Baseado no [Chatwoot](https://github.com/chatwoot/chatwoot) original.
