# Development Guide

## Prerequisites

- Go 1.21+
- Node.js 18+
- Docker & Docker Compose
- PostgreSQL 14+
- Redis 7+
- Git

## Getting Started

### 1. Clone the Repository

```bash
git clone <repository-url>
cd chatwoot-go
```

### 2. Start Infrastructure with Docker

```bash
docker-compose up -d postgres redis rabbitmq minio
```

This will start:

- PostgreSQL on port 5432
- Redis on port 6379
- RabbitMQ on port 5672 (Management UI on 15672)
- MinIO on port 9000 (Console on 9001)

### 3. Setup Backend

```bash
cd backend

# Copy environment file
cp ../.env.example .env

# Install dependencies
go mod download

# Run migrations (automatic on startup)
# Start server
go run cmd/server/main.go
```

Backend will be available at `http://localhost:8080`

### 4. Setup Frontend

```bash
cd frontend

# Install dependencies
npm install

# Start dev server
npm run dev
```

Frontend will be available at `http://localhost:5173`

## Development Workflow

### Backend Development

#### Hot Reload with Air

```bash
cd backend
air
```

Air will automatically rebuild and restart the server when you make changes.

#### Running Tests

```bash
go test ./...
```

#### Database Migrations

Migrations run automatically on startup using GORM AutoMigrate.

To create manual migrations:

```bash
# TODO: Add migration tool
```

#### Adding New Endpoints

1. Define model in `internal/models/`
2. Create handler in `internal/handlers/`
3. Add route in `internal/routes/routes.go`
4. Update API documentation

### Frontend Development

#### Component Development

Create components in `src/components/`:

```tsx
// src/components/MyComponent.tsx
export default function MyComponent() {
  return <div>Hello World</div>;
}
```

#### Adding New Pages

1. Create page in `src/pages/`
2. Add route in `src/App.tsx`
3. Update navigation in `src/components/Sidebar.tsx`

#### State Management

Use Zustand for global state:

```tsx
// src/stores/myStore.ts
import { create } from "zustand";

interface MyState {
  count: number;
  increment: () => void;
}

export const useMyStore = create<MyState>((set) => ({
  count: 0,
  increment: () => set((state) => ({ count: state.count + 1 })),
}));
```

#### API Calls

Use React Query for server state:

```tsx
import { useQuery } from "@tanstack/react-query";
import { api } from "@/lib/api";

function MyComponent() {
  const { data, isLoading } = useQuery({
    queryKey: ["myData"],
    queryFn: async () => {
      const response = await api.get("/my-endpoint");
      return response.data;
    },
  });

  if (isLoading) return <div>Loading...</div>;
  return <div>{JSON.stringify(data)}</div>;
}
```

## Code Style

### Backend (Go)

- Follow [Effective Go](https://golang.org/doc/effective_go.html)
- Use `gofmt` for formatting
- Use `golint` for linting
- Write tests for business logic
- Use meaningful variable names
- Add comments for exported functions

### Frontend (TypeScript/React)

- Use TypeScript for type safety
- Follow React best practices
- Use functional components with hooks
- Use ESLint and Prettier
- Write meaningful component names
- Add JSDoc comments for complex functions

## Debugging

### Backend

Use VS Code debugger or Delve:

```bash
dlv debug cmd/server/main.go
```

### Frontend

Use browser DevTools and React DevTools extension.

## Environment Variables

### Backend

```env
DATABASE_URL=postgresql://user:pass@localhost:5432/chatwoot_go
REDIS_URL=redis://localhost:6379
JWT_SECRET=your-secret-key
PORT=8080
FRONTEND_URL=http://localhost:5173
```

### Frontend

```env
VITE_API_URL=http://localhost:8080
VITE_WS_URL=ws://localhost:8080
```

## Common Issues

### Port Already in Use

```bash
# Find process using port
lsof -i :8080

# Kill process
kill -9 <PID>
```

### Database Connection Failed

1. Ensure PostgreSQL is running
2. Check DATABASE_URL in .env
3. Verify credentials

### Frontend Build Errors

```bash
# Clear node_modules and reinstall
rm -rf node_modules package-lock.json
npm install
```

## Performance Tips

- Use database indexes for frequently queried fields
- Implement pagination for large datasets
- Use Redis caching for expensive queries
- Optimize images and assets
- Use React.memo for expensive components
- Implement code splitting

## Contributing

1. Create a feature branch
2. Make your changes
3. Write tests
4. Run linters
5. Submit pull request

## Resources

- [Go Documentation](https://golang.org/doc/)
- [Gin Framework](https://gin-gonic.com/)
- [GORM](https://gorm.io/)
- [React Documentation](https://react.dev/)
- [TailwindCSS](https://tailwindcss.com/)
- [React Query](https://tanstack.com/query)
