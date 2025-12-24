# Chatwoot-Go Architecture

## Overview

Chatwoot-Go is a modern reimplementation of Chatwoot using Go for the backend and React for the frontend.

## Technology Stack

### Backend

- **Go 1.21+**: Main programming language
- **Gin**: Web framework
- **GORM**: ORM for database operations
- **PostgreSQL**: Primary database
- **Redis**: Caching and sessions
- **RabbitMQ**: Message queue for async operations
- **MinIO**: S3-compatible object storage
- **WebSocket**: Real-time communication

### Frontend

- **React 18**: UI library
- **TypeScript**: Type safety
- **Vite**: Build tool and dev server
- **TailwindCSS**: Utility-first CSS
- **React Query**: Server state management
- **Zustand**: Client state management
- **React Router**: Routing
- **Socket.io**: Real-time client

## Architecture Diagram

```
┌─────────────────┐
│   React App     │
│  (Frontend)     │
└────────┬────────┘
         │ HTTP/WS
         ▼
┌─────────────────┐
│   Gin Server    │
│   (Backend)     │
└────────┬────────┘
         │
    ┌────┴────┬────────┬─────────┐
    ▼         ▼        ▼         ▼
┌────────┐ ┌──────┐ ┌──────┐ ┌──────┐
│Postgres│ │Redis │ │RabbitMQ│ │MinIO│
└────────┘ └──────┘ └──────┘ └──────┘
```

## Core Components

### Backend Structure

```
backend/
├── cmd/
│   └── server/          # Application entry point
├── internal/
│   ├── config/          # Configuration management
│   ├── database/        # Database connections
│   ├── handlers/        # HTTP handlers
│   ├── middleware/      # HTTP middleware
│   ├── models/          # Data models
│   ├── routes/          # Route definitions
│   ├── services/        # Business logic
│   └── websocket/       # WebSocket hub
└── pkg/                 # Reusable packages
```

### Frontend Structure

```
frontend/
├── src/
│   ├── components/      # Reusable components
│   ├── pages/           # Page components
│   ├── stores/          # Zustand stores
│   ├── lib/             # Utilities and API client
│   ├── hooks/           # Custom React hooks
│   └── types/           # TypeScript types
└── public/              # Static assets
```

## Data Models

### Core Entities

1. **Account**: Workspace/organization
2. **User**: Agent or administrator
3. **Inbox**: Communication channel (web, WhatsApp, email, etc.)
4. **Contact**: Customer/end-user
5. **Conversation**: Thread of messages
6. **Message**: Individual message
7. **Team**: Group of agents
8. **Label**: Tag for conversations

### Relationships

- Account has many Users, Inboxes, Contacts, Conversations
- User belongs to many Accounts (many-to-many)
- Inbox belongs to Account, has many Conversations
- Contact belongs to Account, has many Conversations
- Conversation belongs to Account, Inbox, Contact
- Conversation has many Messages
- Message belongs to Conversation

## API Design

### RESTful Endpoints

```
POST   /api/v1/auth/login
POST   /api/v1/auth/register
GET    /api/v1/profile

GET    /api/v1/conversations
POST   /api/v1/conversations
GET    /api/v1/conversations/:id
PUT    /api/v1/conversations/:id
DELETE /api/v1/conversations/:id

GET    /api/v1/messages
POST   /api/v1/messages
GET    /api/v1/messages/:id

GET    /api/v1/contacts
POST   /api/v1/contacts
GET    /api/v1/contacts/:id
PUT    /api/v1/contacts/:id

GET    /api/v1/inboxes
POST   /api/v1/inboxes
GET    /api/v1/inboxes/:id
```

### WebSocket Events

```
subscribe         - Subscribe to conversation updates
unsubscribe       - Unsubscribe from conversation
message.created   - New message in conversation
conversation.updated - Conversation status changed
typing.started    - User started typing
typing.stopped    - User stopped typing
```

## Authentication & Authorization

### JWT-based Authentication

1. User logs in with email/password
2. Server validates credentials
3. Server generates JWT token with claims:
   - user_id
   - email
   - role
   - account_id
4. Client stores token in localStorage
5. Client sends token in Authorization header
6. Server validates token on protected routes

### Role-Based Access Control

- **Administrator**: Full access to account
- **Agent**: Can manage conversations and contacts
- **Supervisor**: Can view reports and manage agents

## Real-time Communication

### WebSocket Hub

- Maintains active client connections
- Broadcasts messages to subscribed clients
- Supports room-based subscriptions (conversations)
- Handles client registration/unregistration

### Message Flow

1. Client connects to WebSocket endpoint
2. Client subscribes to conversation rooms
3. Server broadcasts new messages to room subscribers
4. Client receives and displays messages in real-time

## Deployment

### Docker Compose (Development)

```bash
docker-compose up -d
```

### Production Deployment

1. Build Docker images
2. Push to registry
3. Deploy to Kubernetes/VPS
4. Configure environment variables
5. Run migrations
6. Start services

## Scalability Considerations

- **Horizontal Scaling**: Multiple backend instances behind load balancer
- **Database**: PostgreSQL with read replicas
- **Cache**: Redis cluster for distributed caching
- **Message Queue**: RabbitMQ cluster for reliability
- **Storage**: MinIO distributed mode or S3
- **WebSocket**: Sticky sessions or Redis adapter for multi-instance

## Security

- JWT token expiration (24 hours)
- Password hashing with bcrypt
- CORS configuration
- SQL injection prevention (GORM parameterized queries)
- XSS prevention (React auto-escaping)
- Rate limiting (TODO)
- Input validation
- HTTPS in production

## Performance Optimization

- Database indexing on frequently queried fields
- Redis caching for frequently accessed data
- Connection pooling for database
- Lazy loading in frontend
- Code splitting in React
- CDN for static assets
- Gzip compression
