# Hub User Service - Setup Guide

## ✅ Good News: Already Uses Auto-Loading!

The hub-user-service **already has `godotenv` support** and automatically loads configuration from `config.env`. No manual exports needed!

## Quick Setup

### 1. Verify Configuration

Run the automated configuration checker:

```bash
cd hub-user-service
./verify-config.sh
```

This script will:
- ✅ Check if `config.env` exists
- ✅ Verify all required environment variables
- ✅ Test PostgreSQL connection
- ✅ Offer to fix common issues automatically

### 2. Start the Service

```bash
make run
```

**That's it!** The service automatically loads `config.env` - no manual exports needed! 🎉

## Configuration

### config.env (Already Exists)

The service uses `config.env` for configuration:

```bash
# JWT Secret (MUST match other services)
MY_JWT_SECRET=HubInv3stm3nts_S3cur3_JWT_K3y_2024_!@#$%^

# Server Configuration
HTTP_PORT=localhost:8080
GRPC_PORT=localhost:50051

# Database Configuration
DB_HOST=localhost
DB_PORT=5432
DB_USER=postgres
DB_PASSWORD=postgres
DB_NAME=hub_investments
DB_SSLMODE=disable

# Redis Configuration
REDIS_HOST=localhost
REDIS_PORT=6379

# Environment
ENVIRONMENT=development
```

### How It Works

The service automatically loads `config.env` on startup:

```go
// internal/config/config.go
func Load() *Config {
    once.Do(func() {
        // Automatically loads config.env
        err := godotenv.Load("config.env")
        if err != nil {
            log.Printf("Warning: Could not load config.env file: %v", err)
            log.Println("Using environment variables or default values...")
        }
        // ... loads configuration
    })
    return instance
}
```

## Common Issues & Solutions

### Issue 1: "role 'postgres' does not exist"

**Solution A: Create postgres role**

```bash
psql postgres -c "CREATE ROLE postgres WITH LOGIN PASSWORD 'postgres' SUPERUSER;"
```

**Solution B: Use your current user (Recommended for Mac)**

```bash
# Find your username
whoami

# Update config.env
# Change DB_USER to your username
# Set DB_PASSWORD to empty
```

Or let the script do it:

```bash
./verify-config.sh
# Follow the prompts to update DB_USER
```

### Issue 2: "database 'hub_investments' does not exist"

**Solution:**

```bash
createdb hub_investments
```

Or let the script do it:

```bash
./verify-config.sh
# Follow the prompts to create the database
```

### Issue 3: "Could not load config.env file"

**Solution:**

```bash
# Create config.env from example
cp config.env.example config.env

# Edit with your settings
nano config.env
```

### Issue 4: PostgreSQL not running

**Solution:**

```bash
# Install PostgreSQL (if not installed)
brew install postgresql@14

# Start PostgreSQL
brew services start postgresql@14

# Verify it's running
psql postgres -c "SELECT version();"
```

## Configuration Priority

The service loads configuration in this order:

1. **`config.env` file** (if exists)
2. **Environment variables** (override config.env)
3. **Default values** (fallback)

Example of overriding:

```bash
# Override just one value
DB_USER=myuser make run

# Or export for the session
export DB_USER=myuser
make run
```

## Database Setup

### 1. Ensure PostgreSQL is Running

```bash
brew services start postgresql@14
```

### 2. Create Database

```bash
createdb hub_investments
```

### 3. Run Migrations

```bash
make migrate-up
```

### 4. Verify Setup

```bash
psql hub_investments -c "\dt"
# Should show the users table
```

## Environment Variables Reference

| Variable | Description | Default | Required |
|----------|-------------|---------|----------|
| `MY_JWT_SECRET` | JWT signing secret (MUST match other services) | - | ✅ YES |
| `HTTP_PORT` | HTTP server port | `localhost:8080` | ✅ YES |
| `GRPC_PORT` | gRPC server port | `localhost:50051` | ✅ YES |
| `DB_HOST` | PostgreSQL host | `localhost` | ✅ YES |
| `DB_PORT` | PostgreSQL port | `5432` | ✅ YES |
| `DB_USER` | PostgreSQL user | `postgres` | ✅ YES |
| `DB_PASSWORD` | PostgreSQL password | `postgres` | ⚠️ Optional |
| `DB_NAME` | Database name | `hub_investments` | ✅ YES |
| `DB_SSLMODE` | SSL mode | `disable` | No |
| `REDIS_HOST` | Redis host | `localhost` | No |
| `REDIS_PORT` | Redis port | `6379` | No |
| `ENVIRONMENT` | Environment (development/production) | `development` | No |

## Testing the Service

### 1. Health Check

```bash
curl http://localhost:8080/health
```

Expected response:
```json
{
  "status": "healthy",
  "timestamp": "2025-10-20T19:00:00Z"
}
```

### 2. Test Login (if you have a user)

```bash
curl -X POST http://localhost:8080/api/v1/login \
  -H "Content-Type: application/json" \
  -d '{
    "email": "test@example.com",
    "password": "password123"
  }'
```

### 3. Test gRPC (using grpcurl)

```bash
# Install grpcurl if needed
brew install grpcurl

# Test gRPC health
grpcurl -plaintext localhost:50051 grpc.health.v1.Health/Check
```

## Makefile Commands

```bash
make help          # Show all available commands
make run           # Start the service
make test          # Run tests
make migrate-up    # Run database migrations
make migrate-down  # Rollback migrations
make build         # Build the binary
make clean         # Clean build artifacts
```

## Troubleshooting

### Service won't start

1. **Check config.env exists:**
   ```bash
   ls -la config.env
   ```

2. **Verify configuration:**
   ```bash
   ./verify-config.sh
   ```

3. **Check PostgreSQL is running:**
   ```bash
   pg_isready
   ```

4. **Check logs:**
   ```bash
   make run 2>&1 | tee service.log
   ```

### Database connection fails

1. **Test connection manually:**
   ```bash
   psql -h localhost -U postgres -d hub_investments
   ```

2. **Check PostgreSQL logs:**
   ```bash
   tail -f /usr/local/var/log/postgresql@14.log
   ```

3. **Verify user exists:**
   ```bash
   psql postgres -c "\du"
   ```

### JWT token issues

Make sure `MY_JWT_SECRET` matches across all services:
- hub-user-service: `config.env` → `MY_JWT_SECRET`
- hub-api-gateway: `.env` → `JWT_SECRET`
- HubInvestmentsServer: `config.env` → `MY_JWT_SECRET`

## Comparison with Other Services

| Service | Config File | Auto-loads? | Helper Script |
|---------|-------------|-------------|---------------|
| HubInvestmentsServer | `config.env` | ✅ YES | - |
| hub-user-service | `config.env` | ✅ YES | `verify-config.sh` |
| hub-api-gateway | `.env` | ✅ YES | `create-env.sh` |

## Development Workflow

### First Time Setup

```bash
# 1. Navigate to service
cd hub-user-service

# 2. Verify/fix configuration
./verify-config.sh

# 3. Run migrations
make migrate-up

# 4. Start service
make run
```

### Daily Development

```bash
cd hub-user-service
make run
```

That's it! The service automatically loads `config.env` every time.

## Integration with Other Services

### With hub-api-gateway

The gateway connects to this service via gRPC:

```yaml
# hub-api-gateway/config/config.yaml
services:
  user-service:
    address: localhost:50051  # Must match GRPC_PORT in config.env
```

### With HubInvestmentsServer (Monolith)

Both services share:
- Same database (`hub_investments`)
- Same JWT secret (`MY_JWT_SECRET`)
- Same users table

## Security Notes

1. **Never commit `config.env`** - It's in `.gitignore`
2. **Use strong JWT secrets** - Minimum 32 characters
3. **Different secrets per environment** - Dev/staging/production
4. **Rotate secrets periodically** - Especially in production
5. **Use SSL in production** - Set `DB_SSLMODE=require`

## Next Steps

1. ✅ Run `./verify-config.sh` to check setup
2. ✅ Fix any issues found
3. ✅ Run `make migrate-up` to set up database
4. ✅ Run `make run` to start the service
5. ✅ Test with `curl http://localhost:8080/health`

## Related Files

- `config.env` - Your active configuration
- `config.env.example` - Configuration template
- `verify-config.sh` - Configuration checker and fixer
- `Makefile` - Build and run commands
- `internal/config/config.go` - Configuration loading logic
- `migrations/` - Database migrations

## Summary

✅ **Already uses auto-loading!** - No manual exports needed  
✅ **Consistent with other services** - Same pattern as HubInvestmentsServer  
✅ **Helper script provided** - `verify-config.sh` for easy setup  
✅ **Well documented** - Complete configuration reference  

The hub-user-service is already set up correctly - just run `./verify-config.sh` to verify your database setup! 🚀

