# User Management Service - Deployment Guide

## 🚀 Quick Deployment

### Option 1: Docker Compose (Recommended for Development)

```bash
# Navigate to the service directory
cd hub-user-service

# Start all services (database + user service)
docker-compose up -d

# Check logs
docker-compose logs -f user-service

# Verify service is running
curl http://localhost:8081/health
```

### Option 2: Local Development

```bash
# 1. Start PostgreSQL (ensure it's running on port 5433)

# 2. Set up environment
cp .env.example .env
# Edit .env with your configuration

# 3. Run database migrations
make migrate-up

# 4. Start the service
make run
```

## 📊 Service Endpoints

After deployment, the service will be available on:

- **HTTP REST API**: `http://localhost:8081`
- **gRPC API**: `localhost:50052`
- **Database**: `localhost:5433`

## 🔧 Configuration

### Environment Variables

Create a `.env` file based on `.env.example`:

```bash
# Server Configuration
HTTP_PORT=8080
GRPC_PORT=50051

# Security
JWT_SECRET=your-secure-secret-key-here

# Database
DB_HOST=user-db
DB_PORT=5432
DB_USER=postgres
DB_PASSWORD=postgres
DB_NAME=hub_users_db
```

### Docker Compose Ports

The docker-compose configuration uses the following ports:

- `8081:8080` - HTTP API (external:internal)
- `50052:50051` - gRPC API (external:internal)
- `5433:5432` - PostgreSQL (external:internal)

## 🧪 Testing the Deployment

### 1. Health Check
```bash
curl http://localhost:8081/health
```

Expected response:
```json
{
  "healthy": true,
  "version": "1.0.0"
}
```

### 2. Register a User
```bash
curl -X POST http://localhost:8081/register \
  -H "Content-Type: application/json" \
  -d '{
    "email": "test@example.com",
    "password": "TestPass123!",
    "firstName": "Test",
    "lastName": "User"
  }'
```

### 3. Login
```bash
curl -X POST http://localhost:8081/login \
  -H "Content-Type: application/json" \
  -d '{
    "email": "test@example.com",
    "password": "TestPass123!"
  }'
```

### 4. Validate Token
```bash
curl -X POST http://localhost:8081/validate-token \
  -H "Authorization: Bearer <your-token-from-login>"
```

## 🔄 Integration with Monolith

### Update Monolith Configuration

Add to the monolith's configuration:

```env
USER_SERVICE_GRPC_ADDRESS=localhost:50052
```

### Use the gRPC Client

The monolith can now use the User Service via gRPC:

```go
// In your monolith's container initialization
userServiceClient, err := grpc.NewUserServiceClient("localhost:50052")
if err != nil {
    log.Fatalf("Failed to connect to User Service: %v", err)
}

// Use the client for authentication
valid, userID, email, err := userServiceClient.ValidateToken(ctx, token)
if err != nil {
    return err
}
```

## 📝 Database Migrations

### Apply Migrations
```bash
make migrate-up
```

### Rollback Migrations
```bash
make migrate-down
```

### Manual Migration
```bash
psql -h localhost -p 5433 -U postgres -d hub_users_db -f migrations/000001_create_users_table.up.sql
```

## 🛠️ Troubleshooting

### Service won't start

1. **Check if ports are available**
   ```bash
   lsof -i :8081
   lsof -i :50052
   lsof -i :5433
   ```

2. **Check logs**
   ```bash
   docker-compose logs user-service
   ```

3. **Verify database connection**
   ```bash
   docker-compose logs user-db
   ```

### Database connection issues

1. **Check database is running**
   ```bash
   docker-compose ps
   ```

2. **Test database connection**
   ```bash
   psql postgres://postgres:postgres@localhost:5433/hub_users_db
   ```

3. **Check database logs**
   ```bash
   docker-compose logs user-db
   ```

### gRPC connection issues from monolith

1. **Verify service is listening**
   ```bash
   netstat -an | grep 50052
   ```

2. **Test gRPC connection**
   ```bash
   grpcurl -plaintext localhost:50052 userservice.UserService/HealthCheck
   ```

## 🔒 Security Considerations

### Production Deployment

1. **Change default JWT secret**
   ```bash
   export JWT_SECRET=$(openssl rand -base64 32)
   ```

2. **Enable SSL/TLS for gRPC**
   - Generate certificates
   - Update gRPC server configuration
   - Update client connection credentials

3. **Use secure database credentials**
   - Change default postgres password
   - Use environment-specific secrets

4. **Enable database SSL**
   ```
   DATABASE_URL=postgres://user:pass@host:port/db?sslmode=require
   ```

## 📊 Monitoring

### Docker Stats
```bash
docker stats hub-user-service
```

### Database Connections
```bash
docker-compose exec user-db psql -U postgres -d hub_users_db -c "SELECT count(*) FROM pg_stat_activity;"
```

### Service Health
```bash
# HTTP health check
curl http://localhost:8081/health

# gRPC health check
grpcurl -plaintext localhost:50052 userservice.UserService/HealthCheck
```

## 🚦 Production Checklist

- [ ] Change JWT_SECRET to a secure random value
- [ ] Configure production database credentials
- [ ] Enable SSL/TLS for all connections
- [ ] Set up monitoring and alerting
- [ ] Configure log aggregation
- [ ] Set up automated backups for the database
- [ ] Configure rate limiting
- [ ] Set up firewall rules
- [ ] Configure container resource limits
- [ ] Set up health checks in orchestration platform
- [ ] Configure auto-scaling policies
- [ ] Set up disaster recovery procedures

## 📈 Scaling

### Horizontal Scaling

The service is stateless and can be scaled horizontally:

```bash
docker-compose up -d --scale user-service=3
```

### Load Balancing

Use a load balancer (nginx, HAProxy, or cloud load balancer) to distribute traffic:

```nginx
upstream user_service_http {
    server user-service-1:8080;
    server user-service-2:8080;
    server user-service-3:8080;
}

upstream user_service_grpc {
    server user-service-1:50051;
    server user-service-2:50051;
    server user-service-3:50051;
}
```

## 🔄 Updates and Rollbacks

### Update Service
```bash
# Pull latest changes
git pull origin main

# Rebuild and restart
make docker-rebuild
```

### Rollback
```bash
# Stop current version
docker-compose down

# Checkout previous version
git checkout <previous-commit>

# Start previous version
docker-compose up -d
```
