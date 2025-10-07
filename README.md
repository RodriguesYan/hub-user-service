# Hub User Management Service

A microservice for user authentication and management, extracted from the Hub Investments monolithic application. This service provides both HTTP REST API and gRPC interfaces for user operations.

## 🎯 Features

- **User Authentication**: JWT-based authentication with secure password hashing (bcrypt)
- **User Registration**: Create new user accounts with validation
- **User Profile Management**: Retrieve and manage user profile information
- **Token Validation**: Validate JWT tokens for inter-service communication
- **Account Security**: Failed login attempt tracking, account locking, and email verification
- **Dual Protocol Support**: Both HTTP REST and gRPC endpoints
- **Clean Architecture**: Domain-driven design with clear separation of concerns

## 🏗️ Architecture

The service follows Clean Architecture principles with the following layers:

```
hub-user-service/
├── cmd/server/               # Application entry point
├── internal/
│   ├── domain/              # Business logic and entities
│   │   ├── model/           # Domain models (User, Email, Password)
│   │   ├── repository/      # Repository interfaces
│   │   └── service/         # Domain services (Auth, Token)
│   ├── application/         # Application business rules
│   │   └── usecase/         # Use cases
│   ├── infra/               # External dependencies
│   │   └── persistence/     # Database implementation
│   └── presentation/        # Interfaces to the outside world
│       ├── grpc/            # gRPC server implementation
│       └── http/            # HTTP REST handlers
├── shared/
│   ├── config/              # Configuration management
│   ├── container/           # Dependency injection
│   └── database/            # Database connection
├── proto/                   # Protobuf definitions
├── migrations/              # Database migrations
└── docker/                  # Docker configuration
```

## 📋 Prerequisites

- Go 1.23 or higher
- PostgreSQL 16
- Docker and Docker Compose (optional)
- Protocol Buffers compiler (protoc)

## 🚀 Quick Start

### Using Docker Compose (Recommended)

1. **Clone the repository**
   ```bash
   git clone <repository-url>
   cd hub-user-service
   ```

2. **Start the services**
   ```bash
   make docker-up
   ```

3. **Verify the service is running**
   ```bash
   curl http://localhost:8081/health
   ```

### Manual Setup

1. **Install dependencies**
   ```bash
   go mod download
   ```

2. **Set up environment variables**
   ```bash
   cp .env.example .env
   # Edit .env with your configuration
   ```

3. **Start PostgreSQL database**
   ```bash
   # Make sure PostgreSQL is running on port 5433
   ```

4. **Run database migrations**
   ```bash
   make migrate-up
   ```

5. **Run the service**
   ```bash
   make run
   ```

## 🔧 Configuration

Configuration is managed through environment variables:

| Variable | Description | Default |
|----------|-------------|---------|
| `HTTP_PORT` | HTTP server port | `8080` |
| `GRPC_PORT` | gRPC server port | `50051` |
| `JWT_SECRET` | Secret key for JWT signing | `your-secret-key-change-in-production` |
| `TOKEN_EXPIRATION` | Token expiration duration | `10m` |
| `DATABASE_URL` | PostgreSQL connection string | See `.env.example` |
| `DB_HOST` | Database host | `localhost` |
| `DB_PORT` | Database port | `5433` |
| `DB_USER` | Database user | `postgres` |
| `DB_PASSWORD` | Database password | `postgres` |
| `DB_NAME` | Database name | `hub_users_db` |
| `ENVIRONMENT` | Environment (development/production) | `development` |

## 📡 API Endpoints

### HTTP REST API

#### Public Endpoints

- **POST /login** - Authenticate user and get JWT token
  ```json
  {
    "email": "user@example.com",
    "password": "SecurePass123!"
  }
  ```

- **POST /register** - Register new user
  ```json
  {
    "email": "user@example.com",
    "password": "SecurePass123!",
    "firstName": "John",
    "lastName": "Doe"
  }
  ```

- **POST /validate-token** - Validate JWT token
  ```
  Headers: Authorization: Bearer <token>
  ```

- **GET /health** - Health check endpoint

#### Protected Endpoints (Require Authentication)

- **GET /profile** - Get user profile
  ```
  Headers: Authorization: Bearer <token>
  ```

### gRPC API

The service exposes the following gRPC methods:

```protobuf
service UserService {
  rpc Login(LoginRequest) returns (LoginResponse);
  rpc ValidateToken(ValidateTokenRequest) returns (ValidateTokenResponse);
  rpc RegisterUser(RegisterUserRequest) returns (RegisterUserResponse);
  rpc GetUserProfile(GetUserProfileRequest) returns (GetUserProfileResponse);
  rpc HealthCheck(HealthCheckRequest) returns (HealthCheckResponse);
}
```

## 🔐 Security Features

- **Password Requirements**:
  - Minimum 8 characters
  - At least one uppercase letter
  - At least one lowercase letter
  - At least one digit
  - At least one special character
  - Maximum 72 characters (bcrypt limit)

- **Account Protection**:
  - Failed login attempt tracking
  - Automatic account locking after 5 failed attempts
  - 30-minute lockout period
  - Email verification support

- **Token Security**:
  - JWT-based authentication
  - Configurable token expiration
  - Secure token signing with HMAC-SHA256

## 🧪 Testing

### Run All Tests
```bash
make test
```

### Run Tests with Coverage
```bash
make test-coverage
```

### Run Specific Tests
```bash
go test -v ./internal/domain/model/...
```

## 📊 Database Schema

The service uses a PostgreSQL database with the following schema:

```sql
yanrodrigues.users (
    id UUID PRIMARY KEY,
    email VARCHAR(255) UNIQUE NOT NULL,
    password VARCHAR(255) NOT NULL,
    first_name VARCHAR(100) NOT NULL,
    last_name VARCHAR(100) NOT NULL,
    is_active BOOLEAN DEFAULT true,
    email_verified BOOLEAN DEFAULT false,
    created_at TIMESTAMP WITH TIME ZONE,
    updated_at TIMESTAMP WITH TIME ZONE,
    last_login_at TIMESTAMP WITH TIME ZONE,
    locked_until TIMESTAMP WITH TIME ZONE,
    failed_login_attempts INTEGER DEFAULT 0
)
```

## 🛠️ Development

### Generate Protobuf Code
```bash
make proto
```

### Run Database Migrations
```bash
# Up
make migrate-up

# Down
make migrate-down
```

### Run Linter
```bash
make lint
```

### Build Binary
```bash
make build
```

## 🐳 Docker

### Build and Run with Docker Compose
```bash
make docker-up
```

### View Logs
```bash
make docker-logs
```

### Rebuild Services
```bash
make docker-rebuild
```

### Stop Services
```bash
make docker-down
```

## 📈 Monitoring

The service exposes several endpoints for monitoring:

- **/health** - HTTP health check (returns 200 OK if healthy)
- **HealthCheck** - gRPC health check method

## 🔄 Integration with Monolith

The monolith can use this service via gRPC for authentication and user management operations. A gRPC client is provided in the monolith repository.

Example usage from monolith:
```go
// Connect to User Service
conn, err := grpc.Dial("localhost:50052", grpc.WithInsecure())
client := pb.NewUserServiceClient(conn)

// Validate token
resp, err := client.ValidateToken(ctx, &pb.ValidateTokenRequest{
    Token: token,
})
```

## 📝 API Examples

### Register a New User
```bash
curl -X POST http://localhost:8081/register \
  -H "Content-Type: application/json" \
  -d '{
    "email": "john.doe@example.com",
    "password": "SecurePass123!",
    "firstName": "John",
    "lastName": "Doe"
  }'
```

### Login
```bash
curl -X POST http://localhost:8081/login \
  -H "Content-Type: application/json" \
  -d '{
    "email": "john.doe@example.com",
    "password": "SecurePass123!"
  }'
```

### Get User Profile
```bash
curl -X GET http://localhost:8081/profile \
  -H "Authorization: Bearer <your-jwt-token>"
```

## 🚧 Roadmap

- [ ] Unit tests for all layers
- [ ] Integration tests
- [ ] Email verification workflow
- [ ] Password reset functionality
- [ ] OAuth 2.0 integration
- [ ] Multi-factor authentication (MFA)
- [ ] Rate limiting
- [ ] Prometheus metrics
- [ ] Distributed tracing with Jaeger

## 📄 License

This project is part of the Hub Investments platform.

## 🤝 Contributing

This is a microservice extraction from the Hub Investments monolithic application as part of the microservices migration strategy outlined in Phase 10 of the project roadmap.

## 📧 Support

For issues and questions, please refer to the main Hub Investments project documentation.
