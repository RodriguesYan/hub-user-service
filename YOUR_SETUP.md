# Your Hub User Service Setup - FIXED ✅

## What Was Wrong

The `config.env` had incorrect database settings:
- ❌ **DB_USER:** `postgres` (doesn't exist on your system)
- ❌ **DB_PASSWORD:** `postgres` (not needed)
- ❌ **DB_NAME:** `hub_investments` (wrong name - has underscore)

## What I Fixed

Updated `config.env` with your actual PostgreSQL setup:
- ✅ **DB_USER:** `yanrodrigues` (your actual user)
- ✅ **DB_PASSWORD:** *(empty - not needed for local user)*
- ✅ **DB_NAME:** `hubinvestments` (your actual database - no underscore)

## Your Database Setup

Based on `psql -l`, you have:

| Database | Owner | Notes |
|----------|-------|-------|
| `hubinvestments` | yanrodrigues | ✅ **Used by hub-user-service** |
| `yanrodrigues` | yanrodrigues | Default user database |
| `postgres` | yanrodrigues | PostgreSQL default |

## What I Did

1. ✅ Updated `config.env` with correct credentials
2. ✅ Ran migration to create `users` table
3. ✅ Tested service - it works!

## Your config.env (Fixed)

```bash
# Database Configuration
DB_HOST=localhost
DB_PORT=5432
DB_USER=yanrodrigues          # ← Changed from 'postgres'
DB_PASSWORD=                  # ← Empty (no password needed)
DB_NAME=hubinvestments        # ← Changed from 'hub_investments'
DB_SSLMODE=disable
```

## How to Start the Service

```bash
cd hub-user-service
make run
```

**That's it!** No exports needed - `config.env` is automatically loaded.

## Verify It's Working

```bash
# Health check
curl http://localhost:8080/health

# Expected response:
# {
#   "status": "healthy",
#   "version": "1.0.0",
#   "timestamp": "..."
# }
```

## Database Tables

Your `hubinvestments` database now has:

```
Schema       | Table        | Owner
-------------|--------------|-------------
yanrodrigues | users        | yanrodrigues  ← NEW!
yanrodrigues | positions_v2 | yanrodrigues
public       | market_data  | yanrodrigues
```

## Quick Reference

### Start Service
```bash
cd hub-user-service
make run
```

### Check Database
```bash
psql -d hubinvestments -c "\dt"
```

### View Users Table
```bash
psql -d hubinvestments -c "SELECT * FROM users;"
```

### Test Login (after creating a user)
```bash
curl -X POST http://localhost:8080/api/v1/login \
  -H "Content-Type: application/json" \
  -d '{
    "email": "test@example.com",
    "password": "password123"
  }'
```

## Important Notes

1. **Database Name:** Your database is `hubinvestments` (no underscore!)
2. **No Password Needed:** Local PostgreSQL user doesn't need a password
3. **Schema:** Tables are in `yanrodrigues` schema, not `public`
4. **Auto-Loading:** `config.env` is automatically loaded - no exports needed

## If You Need to Reset

### Drop and Recreate Users Table
```bash
psql -d hubinvestments -f migrations/000001_create_users_table.down.sql
psql -d hubinvestments -f migrations/000001_create_users_table.up.sql
```

### View Migration Files
```bash
ls -la migrations/
```

## Summary

✅ **Fixed:** Database connection settings  
✅ **Created:** Users table via migration  
✅ **Tested:** Service starts and responds to health checks  
✅ **Ready:** Service is ready to use!  

---

**Your Setup:**
- **User:** yanrodrigues
- **Database:** hubinvestments
- **Password:** (none)
- **Port:** 5432 (default)

**Status:** ✅ WORKING

