# 🚀 Quick Start Guide

## Início Rápido com Docker Compose

### 1. Pré-requisitos

- Docker Desktop instalado
- Git
- Porta 8080 (backend) e 5173 (frontend) disponíveis

### 2. Clone e Configure

```bash
# Navegue até a pasta do projeto
cd chatwoot-go

# Copie o arquivo de ambiente
copy .env.example .env
```

### 3. Inicie os Serviços

```bash
# Inicie TODOS os serviços (recomendado para primeira vez)
docker-compose up -d

# Ou inicie apenas a infraestrutura (para desenvolvimento local)
docker-compose up -d postgres redis rabbitmq minio
```

### 4. Acesse a Aplicação

- **Frontend**: http://localhost:5173
- **Backend API**: http://localhost:8080
- **MinIO Console**: http://localhost:9001 (minioadmin / minioadmin123)
- **RabbitMQ Management**: http://localhost:15672 (chatwoot / chatwoot123)

### 5. Crie sua Primeira Conta

1. Acesse http://localhost:5173
2. Clique em "Sign up"
3. Preencha os dados:
   - Nome: Seu Nome
   - Email: seu@email.com
   - Senha: mínimo 8 caracteres
4. Clique em "Create Account"
5. Você será redirecionado para o Dashboard!

## Desenvolvimento Local (Sem Docker)

### Backend

```bash
cd backend

# Instalar dependências
go mod download

# Configurar .env
copy ..\.env.example .env

# Iniciar servidor
go run cmd/server/main.go
```

### Frontend

```bash
cd frontend

# Instalar dependências
npm install

# Iniciar dev server
npm run dev
```

## Comandos Úteis

### Ver logs dos containers

```bash
docker-compose logs -f
```

### Parar todos os serviços

```bash
docker-compose down
```

### Resetar banco de dados (CUIDADO: apaga todos os dados)

```bash
docker-compose down -v
docker-compose up -d
```

### Rebuild das imagens

```bash
docker-compose build
docker-compose up -d
```

## Estrutura do Projeto

```
chatwoot-go/
├── backend/              # API em Go
│   ├── cmd/              # Entry points
│   ├── internal/         # Código interno
│   │   ├── config/       # Configurações
│   │   ├── database/     # Database & migrations
│   │   ├── handlers/     # HTTP handlers
│   │   ├── middleware/   # Middlewares
│   │   ├── models/       # Models GORM
│   │   ├── routes/       # Rotas
│   │   └── websocket/    # WebSocket hub
│   └── go.mod
│
├── frontend/             # App React
│   ├── src/
│   │   ├── components/   # Componentes
│   │   ├── pages/        # Páginas
│   │   ├── stores/       # Zustand stores
│   │   └── lib/          # API client
│   └── package.json
│
├── docs/                 # Documentação
├── docker-compose.yml    # Docker Compose
└── README.md
```

## Próximos Passos

1. **Explore o Dashboard**: Veja as estatísticas e navegue pela interface
2. **Crie um Inbox**: Configure um canal de comunicação
3. **Adicione Contatos**: Gerencie seus clientes
4. **Teste Conversas**: Experimente o sistema de mensagens
5. **Personalize**: Ajuste configurações e preferências

## Recursos Implementados

✅ Autenticação JWT
✅ CRUD de Usuários
✅ CRUD de Contas
✅ Models completos (Conversations, Messages, Contacts, Inboxes, Teams, Labels)
✅ WebSocket para real-time
✅ Dashboard com estatísticas
✅ Interface moderna com React + TailwindCSS
✅ Docker Compose para desenvolvimento
✅ Migrations automáticas
✅ API RESTful

## Em Desenvolvimento

🚧 Implementação completa de Conversations
🚧 Sistema de mensagens real-time
🚧 Upload de arquivos
🚧 Webhooks
🚧 Integrações (WhatsApp, Email, etc.)
🚧 Sistema de notificações
🚧 Relatórios e analytics
🚧 Testes automatizados

## Problemas Comuns

### Porta já em uso

```bash
# Windows
netstat -ano | findstr :8080
taskkill /PID <PID> /F
```

### Erro de conexão com banco

1. Verifique se o PostgreSQL está rodando: `docker-compose ps`
2. Verifique as credenciais no `.env`
3. Tente resetar: `docker-compose restart postgres`

### Frontend não carrega

1. Limpe o cache: `cd frontend && rm -rf node_modules && npm install`
2. Verifique se o backend está rodando
3. Verifique a URL da API no `.env`

## Suporte

- 📖 [Documentação Completa](./docs/)
- 🏗️ [Arquitetura](./docs/architecture.md)
- 💻 [Guia de Desenvolvimento](./docs/development.md)

## Licença

MIT License - Baseado no [Chatwoot](https://github.com/chatwoot/chatwoot) original.
