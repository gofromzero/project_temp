# Multi-Tenant Admin System

A modern full-stack multi-tenant administration system built with Go, React, and modern DevOps practices.

## Tech Stack

### Backend
- **Language**: Go 1.21+
- **Framework**: GoFrame v2.5
- **Architecture**: DDD (Domain Driven Design)
- **Database**: MySQL 8.0
- **Cache**: Redis 7.0

### Frontend
- **Language**: TypeScript 5.0
- **Framework**: React 18.0
- **UI Library**: Ant Design 5.0
- **Styling**: Tailwind CSS 3.0
- **State Management**: Zustand 4.0
- **Build Tool**: Vite 5.0

## Quick Start

### Prerequisites
- Go 1.21+
- Node.js 18+
- Docker & Docker Compose

### Development Setup

1. **Clone the repository**
   ```bash
   git clone <repository-url>
   cd multi-tenant-admin
   ```

2. **Start infrastructure services**
   ```bash
   cd backend
   docker-compose up -d
   ```

3. **Install frontend dependencies**
   ```bash
   cd frontend/apps/admin-web
   npm install
   ```

4. **Install backend dependencies**
   ```bash
   cd backend
   go mod tidy
   ```

5. **Start development servers**
   
   **Option A: Use convenience scripts**
   - Windows: `scripts/start-dev.bat`
   - Unix/Linux: `scripts/start-dev.sh`
   
   **Option B: Manual start**
   ```bash
   # Terminal 1 - Backend
   cd backend
   go run cmd/main.go
   
   # Terminal 2 - Frontend  
   cd frontend/apps/admin-web
   npm run dev
   ```

### Access Points

- **Frontend**: http://localhost:3000
- **Backend API**: http://localhost:8000
- **Health Check**: http://localhost:8000/health

## Project Structure

```
multi-tenant-admin/
├── backend/                    # Go backend application
│   ├── cmd/                   # Application entry points
│   ├── api/                   # API layer (handlers, middleware, routes)
│   ├── domain/                # Domain layer (entities, business logic)
│   ├── application/           # Application layer (use cases)
│   ├── repository/            # Data access layer
│   ├── infr/                  # Infrastructure layer
│   ├── pkg/                   # Shared utilities
│   ├── configs/               # Configuration files
│   └── docker-compose.yaml    # Local development services
├── frontend/                  # Frontend applications
│   ├── apps/admin-web/       # Main admin application
│   └── packages/             # Shared packages
├── scripts/                   # Build and deployment scripts
├── docs/                     # Project documentation
└── .github/workflows/        # CI/CD pipelines
```

## Testing

### Backend Tests
```bash
cd backend
go test ./...
```

### Frontend Tests
```bash
cd frontend/apps/admin-web
npm test
```

### Linting
```bash
# Backend formatting
cd backend
go fmt ./...

# Frontend linting
cd frontend/apps/admin-web
npm run lint
```

## Docker Deployment

### Build Images
```bash
# Backend
docker build -f backend/Dockerfile -t multi-tenant-backend .

# Frontend
docker build -f frontend/apps/admin-web/Dockerfile -t multi-tenant-frontend .
```

## Environment Configuration

Copy `.env.example` to `.env` and configure your environment variables:

```bash
cp .env.example .env
```

## Hello World Verification

1. Start both frontend and backend servers
2. Visit http://localhost:3000
3. Click "Test Backend Connection" button
4. Verify successful connection and health check response

## CI/CD

The project includes GitHub Actions workflows for:
- Automated testing (backend & frontend)
- Code quality checks (linting, formatting)
- Docker image building
- Deployment pipeline

## Contributing

1. Follow Go and TypeScript best practices
2. Ensure all tests pass
3. Run linting before submitting PRs
4. Update documentation as needed

## License

[Your License Here]