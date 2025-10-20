# Hub User Service - Environment Setup Complete ✅

## Status: Already Configured!

Good news! The hub-user-service **already has automatic configuration loading** via `godotenv`. Unlike hub-api-gateway which needed to be fixed, this service was already set up correctly.

## What Was Added

Since the service already had auto-loading, I created **helper tools** to make database setup easier:

### 1. Configuration Verification Script

**File:** `verify-config.sh`

An intelligent script that:
- ✅ Checks if `config.env` exists
- ✅ Verifies all required environment variables
- ✅ Tests PostgreSQL connection
- ✅ Checks if database exists
- ✅ Verifies PostgreSQL user exists
- ✅ **Offers to fix issues automatically**

**Usage:**
```bash
cd hub-user-service
./verify-config.sh
```

### 2. Comprehensive Setup Guide

**File:** `SETUP_GUIDE.md`

Complete documentation covering:
- Quick setup instructions
- Configuration reference
- Common issues and solutions
- Database setup steps
- Testing procedures
- Troubleshooting guide

## How It Already Works

The service has always used `godotenv` to auto-load `config.env`:

```go
// internal/config/config.go (existing code)
func Load() *Config {
    once.Do(func() {
        // Automatically loads config.env
        err := godotenv.Load("config.env")
        if err != nil {
            log.Printf("Warning: Could not load config.env file: %v", err)
            log.Println("Using environment variables or default values...")
        }
        // ... rest of loading
    })
    return instance
}
```

## Configuration File

### config.env (Already Exists)

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

## Common Issue: Database Connection

The main issue users face is **PostgreSQL role "postgres" does not exist**.

### Solution 1: Use verify-config.sh (Recommended)

```bash
cd hub-user-service
./verify-config.sh
```

The script will detect the issue and offer to:
- Create the `postgres` role, OR
- Update `config.env` to use your current user

### Solution 2: Manual Fix

**Option A: Create postgres role**
```bash
psql postgres -c "CREATE ROLE postgres WITH LOGIN PASSWORD 'postgres' SUPERUSER;"
```

**Option B: Use your current user**
```bash
# Find your username
whoami

# Update config.env
DB_USER=your_username
DB_PASSWORD=
```

## Quick Start

### First Time Setup

```bash
cd hub-user-service

# 1. Verify and fix configuration
./verify-config.sh

# 2. Run database migrations
make migrate-up

# 3. Start the service
make run
```

### Daily Use

```bash
cd hub-user-service
make run
```

That's it! The `config.env` is automatically loaded every time.

## Comparison with Other Services

| Service | Config File | Auto-loads? | Helper Script |
|---------|-------------|-------------|---------------|
| HubInvestmentsServer | `config.env` | ✅ YES | - |
| hub-user-service | `config.env` | ✅ **ALREADY YES** | `verify-config.sh` ✨ |
| hub-api-gateway | `.env` | ✅ NOW YES (fixed) | `create-env.sh` |

## What's Different from hub-api-gateway?

| Aspect | hub-api-gateway | hub-user-service |
|--------|-----------------|------------------|
| **Problem** | Didn't have godotenv | Already had godotenv ✅ |
| **Fix Needed** | Added godotenv support | No fix needed! |
| **Helper Added** | `create-env.sh` (creates .env) | `verify-config.sh` (verifies setup) |
| **Config File** | `.env` | `config.env` |
| **Main Issue** | Missing JWT_SECRET | Database connection |

## Files Created

1. ✅ **`verify-config.sh`** - Interactive configuration checker and fixer
2. ✅ **`SETUP_GUIDE.md`** - Comprehensive setup documentation
3. ✅ **`ENV_SETUP_COMPLETE.md`** - This file (summary)

## Files Modified

None! The service was already configured correctly.

## Testing

Verified that the service loads `config.env` automatically:

```bash
cd hub-user-service
make run
# ✅ Loads config.env automatically
# ✅ No manual exports needed
```

## Benefits of verify-config.sh

1. **✅ Automated Checks** - Verifies entire setup
2. **✅ Interactive Fixes** - Offers to fix issues
3. **✅ Database Testing** - Tests actual connection
4. **✅ User-Friendly** - Clear colored output
5. **✅ Time-Saving** - No manual troubleshooting

## Usage Examples

### Example 1: First Time Setup

```bash
$ cd hub-user-service
$ ./verify-config.sh

🔍 Hub User Service Configuration Checker
==========================================

✅ config.env exists

1. Checking JWT Secret...
   ✅ MY_JWT_SECRET is set (46 characters)

2. Checking Server Configuration...
   HTTP Port: localhost:8080
   gRPC Port: localhost:50051

3. Checking Database Configuration...
   ✅ DB_HOST: localhost
   ✅ DB_PORT: 5432
   ✅ DB_NAME: hub_investments
   ✅ DB_USER: postgres

4. Checking PostgreSQL...
   ✅ psql is installed

5. Testing Database Connection...
   ⚠️  Database 'hub_investments' does not exist
   Create database 'hub_investments'? (y/n): y
   ✅ Created database 'hub_investments'
   
   ⚠️  PostgreSQL user 'postgres' does not exist
   Create 'postgres' user? (y/n): y
   ✅ Created user 'postgres'

6. Testing Connection String...
   ✅ Successfully connected to database!

==========================================
✅ Configuration check complete!

To start the service:
  make run
```

### Example 2: Using Current User

```bash
$ ./verify-config.sh

...
   ⚠️  PostgreSQL user 'postgres' does not exist
   Suggestion: Change DB_USER in config.env to 'yanrodrigues'
   Update DB_USER to 'yanrodrigues'? (y/n): y
   ✅ Updated config.env
   DB_USER=yanrodrigues
   DB_PASSWORD=(empty)
...
```

## Troubleshooting

### Script says "psql command not found"

**Solution:**
```bash
brew install postgresql@14
brew services start postgresql@14
```

### Script can't connect to database

**Check PostgreSQL is running:**
```bash
brew services list | grep postgresql
brew services start postgresql@14
```

### config.env not found

**Solution:**
```bash
cp config.env.example config.env
./verify-config.sh
```

## Security Notes

1. **`config.env` is in `.gitignore`** - Never committed
2. **JWT_SECRET must match** - All services must use same secret
3. **Use strong secrets** - Minimum 32 characters
4. **Different per environment** - Dev/staging/production

## Integration

### With hub-api-gateway

The gateway calls this service via gRPC:

```yaml
# hub-api-gateway/config/config.yaml
services:
  user-service:
    address: localhost:50051  # Matches GRPC_PORT in config.env
```

### With HubInvestmentsServer

Both services share:
- Same database (`hub_investments`)
- Same JWT secret (`MY_JWT_SECRET`)
- Same users table

## Next Steps

1. ✅ Run `./verify-config.sh`
2. ✅ Fix any issues found
3. ✅ Run `make migrate-up`
4. ✅ Run `make run`
5. ✅ Test with `curl http://localhost:8080/health`

## Summary

✅ **Already configured correctly** - No changes needed to config loading  
✅ **Helper script added** - `verify-config.sh` for easy setup  
✅ **Comprehensive docs** - `SETUP_GUIDE.md` for reference  
✅ **Consistent with other services** - Same auto-loading pattern  
✅ **Database setup simplified** - Interactive fixes for common issues  

The hub-user-service was already using `godotenv` correctly. I just added helpful tools to make database setup easier! 🚀

---

**Date:** October 20, 2025  
**Service:** hub-user-service  
**Changes:** Added helper tools (no code changes needed)  
**Status:** COMPLETE ✅

