# Hub User Service

Microservice responsible for user authentication and management.

## Overview

This service handles:
- User authentication (login)
- JWT token generation and validation
- User data management

## Architecture

This service follows Clean Architecture principles with clear separation of concerns:

```
hub-user-service/
├── cmd/server/          # Application entry point
├── internal/
│   ├── auth/            # Authentication service layer
│   ├── login/           # Login domain logic
│   │   ├── application/ # Application use cases
│   │   ├── domain/      # Business logic & rules
│   │   ├── infra/       # Infrastructure (DB, external services)
│   │   └── presentation/# HTTP/gRPC handlers
│   ├── grpc/            # gRPC server implementation
│   ├── config/          # Configuration management
│   └── database/        # Database connection & utilities
├── migrations/          # Database migrations
└── pkg/                 # Shared packages
```

## Technology Stack

- **Language**: Go 1.23
- **Communication**: gRPC
- **Database**: PostgreSQL
- **Authentication**: JWT (HS256)
- **Configuration**: Environment variables

## Dependencies

- `github.com/golang-jwt/jwt` - JWT token handling
- `github.com/lib/pq` - PostgreSQL driver
- `google.golang.org/grpc` - gRPC framework
- `github.com/joho/godotenv` - Environment configuration

## Getting Started

### Prerequisites

- Go 1.23 or higher
- PostgreSQL 14 or higher
- Access to shared JWT secret

### Environment Variables

```bash
# Server Configuration
GRPC_PORT=50051
HTTP_PORT=8080

# Database
DATABASE_URL=postgresql://user:password@localhost:5432/hub_users

# JWT Configuration
MY_JWT_SECRET=your-secret-key-here

# Environment
ENVIRONMENT=development
```

### Running the Service

```bash
# Install dependencies
go mod download

# Run migrations
make migrate-up

# Run the service
go run cmd/server/main.go
```

### Running Tests

```bash
# Run all tests
go test ./...

# Run with coverage
go test ./... -cover

# Run specific package
go test ./internal/auth/...
```

## API

### gRPC Endpoints

#### Login
```protobuf
rpc Login(LoginRequest) returns (LoginResponse)
```

#### ValidateToken
```protobuf
rpc ValidateToken(ValidateTokenRequest) returns (ValidateTokenResponse)
```

## Development

### Project Structure

- **cmd/**: Application entry points
- **internal/**: Private application code
  - **auth/**: Authentication services (copied AS-IS from monolith)
  - **login/**: Login domain logic (copied AS-IS from monolith)
  - **grpc/**: gRPC server and protocol definitions
  - **config/**: Configuration management
  - **database/**: Database utilities
- **migrations/**: SQL migration files
- **pkg/**: Public packages (if any)

### Code Quality

```bash
# Run linter
golangci-lint run

# Format code
go fmt ./...

# Run tests with race detector
go test -race ./...
```

## Migration from Monolith

This service was created by extracting the authentication and login modules from the HubInvestments monolith.

**Key Principles**:
- Code copied AS-IS (no refactoring during migration)
- All tests migrated and passing
- JWT tokens compatible with monolith
- Shared database during migration phase

## Deployment

### Docker

```bash
# Build image
docker build -t hub-user-service:latest .

# Run container
docker run -p 50051:50051 hub-user-service:latest
```

### Docker Compose

```bash
# Start service with dependencies
docker-compose up
```

## Monitoring

- Health check endpoint: `/health`
- Metrics endpoint: `/metrics`
- Logs: Structured JSON logging

## Contributing

1. Create feature branch
2. Write tests
3. Implement changes
4. Run tests and linter
5. Submit PR

## License

Internal use only - HubInvestments

---

**Status**: Active Development  
**Version**: 1.0.0  
**Last Updated**: 2025-10-13

