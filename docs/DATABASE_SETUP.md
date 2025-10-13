# Database Setup Guide
## Hub User Service - Separate Database Configuration

**Date**: 2025-10-13  
**Status**: Step 2.9 - Database Per Service Pattern  

---

## 🎯 Overview

The Hub User Service uses its **OWN separate database** (`hub_user_service`), independent from the monolith database (`hub_investments`). This follows microservices best practices and ensures service independence.

---

## 📊 Database Architecture

### **Current State (After Setup)**

```
┌───────────────────────────────┐    ┌───────────────────────────────┐
│  User Service Database        │    │  Monolith Database            │
│  (hub_user_service)           │    │  (hub_investments)            │
├───────────────────────────────┤    ├───────────────────────────────┤
│                               │    │                               │
│  • users (source of truth) ✅ │    │  • users (deprecated) ⚠️      │
│                               │    │  • orders                     │
│                               │    │  • positions                  │
│                               │    │  • balance                    │
│                               │    │  • watchlist                  │
│                               │    │  • market_data                │
│                               │    │  • instruments                │
│                               │    │                               │
└───────────────────────────────┘    └───────────────────────────────┘
         ↑                                      ↓
         │                                      │
         └──── gRPC calls for authentication ───┘
                (User Service is authority)
```

---

## 🚀 Quick Start

### **Option 1: Automated Setup (Recommended)**

Run the complete setup in one command:

```bash
make setup-all
```

This will:
1. ✅ Create `hub_user_service` database
2. ✅ Create database user with permissions
3. ✅ Run migrations (create users table)
4. ✅ Copy existing users from monolith
5. ✅ Verify data integrity

---

### **Option 2: Step-by-Step Setup**

#### **Step 1: Create Database**
```bash
make setup-db
```

This creates:
- Database: `hub_user_service`
- User: `hub_user_service_user`
- Password: `hub_user_service_pass`
- Grants all necessary permissions

#### **Step 2: Run Migrations**
```bash
make migrate-up
```

This creates the `users` table with:
- Proper schema (id, email, name, password, timestamps)
- Constraints (unique email, valid email format, password length)
- Indexes (email, created_at)
- Triggers (auto-update updated_at)

#### **Step 3: Migrate User Data**
```bash
make migrate-data
```

This copies users from monolith:
- Reads from `hub_investments.users`
- Writes to `hub_user_service.users`
- Skips duplicates (ON CONFLICT DO NOTHING)
- Verifies data integrity

---

## 📝 Configuration

### **Environment Variables**

Update your `config.env`:

```env
# Database Configuration (User Service Database)
DB_HOST=localhost
DB_PORT=5432
DB_NAME=hub_user_service        # ← Separate database!
DB_USER=hub_user_service_user
DB_PASSWORD=hub_user_service_pass
DB_SSLMODE=disable
```

### **Verify Configuration**

```bash
# Test connection
psql -h localhost -p 5432 -U hub_user_service_user -d hub_user_service

# List tables
\dt

# Count users
SELECT COUNT(*) FROM users;
```

---

## 🔄 Data Migration Details

### **What Gets Migrated**

From `hub_investments.users` → `hub_user_service.users`:

- ✅ `id` - User ID (unchanged)
- ✅ `email` - User email
- ✅ `name` - User name
- ✅ `password` - Password hash
- ✅ `created_at` - Creation timestamp
- ✅ `updated_at` - Last update timestamp

### **Migration Script Behavior**

```sql
INSERT INTO users (id, email, name, password, created_at, updated_at)
SELECT id, email, name, password, created_at, updated_at
FROM users_temp
ON CONFLICT (email) DO NOTHING;  -- Skips duplicates
```

**Features**:
- ✅ Idempotent (safe to run multiple times)
- ✅ Skips duplicates (no errors on re-run)
- ✅ Preserves user IDs (foreign keys unchanged)
- ✅ Maintains timestamps (audit trail preserved)
- ✅ Validates data integrity

### **Data Integrity Checks**

The script automatically verifies:
- All users have valid emails
- All users have passwords
- No NULL or empty critical fields
- User count matches expected

---

## 🎯 Migration Strategy

### **Phase 1: Setup (Week 1) ← YOU ARE HERE**

```
1. Create hub_user_service database          ✅
2. Run migrations                            ✅
3. Copy existing users from monolith         ✅
4. Test User Service independently           ⏭️
```

### **Phase 2: Parallel Operation (Weeks 2-4)**

```
User Service:
  - Source of truth for users
  - Handles all authentication
  - NEW users created here

Monolith:
  - Still has users table (read-only)
  - Gradually migrates to call User Service
  - Uses gRPC for authentication
```

### **Phase 3: Full Cutover (Week 5+)**

```
User Service:
  - ✅ 100% of authentication traffic
  - ✅ All new users
  - ✅ Source of truth

Monolith:
  - users table deprecated
  - Can be dropped after validation period
  - Foreign keys remain (user_id still valid)
```

---

## ✅ Why Separate Database?

### **1. Service Independence**
```
Monolith DB Down ❌  →  User Service Still Works ✅

Without separate DB:
  Monolith DB down → User Service down → All services affected

With separate DB:
  Monolith DB down → User Service works → Core auth still functional
```

### **2. Scalability**
```
User Authentication:  High read, low write
Order Processing:     High write, complex queries

Separate databases allow:
  - Independent scaling strategies
  - Optimized for different workloads
  - No resource contention
```

### **3. Clear Ownership**
```
User Service:     Owns users data
Order Service:    Owns orders data
Position Service: Owns positions data

No shared ownership = No confusion
```

### **4. Technology Freedom**
```
Future services can use:
  - PostgreSQL
  - MongoDB
  - Redis
  - Cassandra

Not locked into monolith's choices
```

---

## ⚠️ Important Considerations

### **1. Foreign Key References**

**Question**: Monolith has `orders.user_id` FK to `users.id`. What happens?

**Answer**: No problem! ✅
- User IDs remain the same
- Monolith services use user_id from JWT token
- No direct queries to users table needed
- Future: Services call User Service for user info via gRPC

### **2. User ID Consistency**

**User IDs are preserved**:
- Same IDs in both databases (during transition)
- JWT tokens contain user_id
- Services use JWT user_id (not database lookup)
- No breaking changes to existing services

### **3. New Users During Migration**

**Recommendation**:
1. **Week 1-2**: Test User Service thoroughly
2. **Week 3**: Deploy User Service, start routing new users
3. **Week 4+**: All new users to User Service only
4. **Week 8+**: Deprecate monolith users table

### **4. Data Synchronization**

**We DON'T sync bi-directionally** ❌

Why? Because:
- Too complex
- Error-prone
- Race conditions
- Data conflicts

**Instead** ✅:
- User Service is **single source of truth**
- Monolith reads from User Service (via gRPC)
- One-way data flow = Simple and reliable

---

## 🛠️ Troubleshooting

### **Issue: Cannot connect to database**

```bash
# Check PostgreSQL is running
pg_isready -h localhost -p 5432

# Check database exists
psql -U postgres -lqt | cut -d \| -f 1 | grep hub_user_service

# Check user has permissions
psql -U hub_user_service_user -d hub_user_service -c "SELECT 1;"
```

### **Issue: Migration fails**

```bash
# Check current migration version
migrate -path migrations -database "postgresql://hub_user_service_user:hub_user_service_pass@localhost:5432/hub_user_service?sslmode=disable" version

# Force to specific version if needed
make migrate-force

# Re-run migrations
make migrate-up
```

### **Issue: Data migration fails**

```bash
# Check source database connectivity
psql -h localhost -U postgres -d hub_investments -c "SELECT COUNT(*) FROM users;"

# Check target database connectivity
psql -h localhost -U hub_user_service_user -d hub_user_service -c "SELECT COUNT(*) FROM users;"

# Re-run data migration (safe, idempotent)
make migrate-data
```

### **Issue: Duplicate users**

This is normal! The migration script uses `ON CONFLICT DO NOTHING`, so:
- Existing users are skipped
- No errors on re-run
- Safe to run multiple times

---

## 🧪 Testing

### **1. Verify Database Setup**

```bash
# Check database exists
psql -U postgres -c "\l" | grep hub_user_service

# Check users table exists
psql -U hub_user_service_user -d hub_user_service -c "\dt"

# Count users
psql -U hub_user_service_user -d hub_user_service -c "SELECT COUNT(*) FROM users;"
```

### **2. Verify Data Integrity**

```sql
-- Check for users with invalid data
SELECT COUNT(*) FROM users WHERE email IS NULL OR email = '';
SELECT COUNT(*) FROM users WHERE password IS NULL OR password = '';
SELECT COUNT(*) FROM users WHERE name IS NULL OR name = '';

-- Compare with monolith
-- In monolith DB:
SELECT COUNT(*) FROM hub_investments.users;
-- In user service DB:
SELECT COUNT(*) FROM hub_user_service.users;
```

### **3. Test User Service**

```bash
# Start user service
make run

# Test login with existing user
grpcurl -plaintext -d '{
  "email": "existing@user.com",
  "password": "their_password"
}' localhost:50051 hub_investments.AuthService/Login

# Should return JWT token ✅
```

---

## 📋 Rollback Plan

If something goes wrong, you can easily rollback:

### **Option 1: Revert to Monolith**

```bash
# 1. Stop user service
# 2. Update monolith to use local auth (revert code changes)
# 3. Monolith users table is unchanged (no data loss)
```

### **Option 2: Re-migrate Data**

```bash
# Drop user service database
psql -U postgres -c "DROP DATABASE hub_user_service;"

# Re-run setup
make setup-all
```

### **Option 3: Keep Both Databases**

During transition, both databases exist:
- User Service DB: New source of truth
- Monolith DB: Backup/fallback

Can switch back instantly if needed.

---

## 🎉 Success Criteria

After setup, you should have:

- [x] ✅ `hub_user_service` database created
- [x] ✅ Users table with proper schema
- [x] ✅ All users migrated from monolith
- [x] ✅ Data integrity verified
- [x] ✅ User Service connects successfully
- [x] ✅ Login works with migrated users
- [x] ✅ JWT tokens generated correctly

---

## 📚 Additional Resources

- [Step 2.9 Documentation](./STEP_2_9_DATABASE_SETUP_COMPLETE.md)
- [Migration Scripts](../scripts/)
- [Makefile Commands](../Makefile)
- [Configuration Guide](../config.env.example)

---

## 🔄 Future: Deprecating Monolith Users Table

After User Service is stable (4-8 weeks):

```sql
-- 1. Verify no monolith code queries users table directly
-- 2. All authentication goes through User Service
-- 3. Drop foreign key constraints (if any)
-- 4. Rename table (don't drop immediately)
ALTER TABLE hub_investments.users RENAME TO users_deprecated;

-- 5. After another validation period (2-4 weeks)
DROP TABLE hub_investments.users_deprecated;
```

---

**Document Version**: 1.0  
**Last Updated**: 2025-10-13  
**Status**: Complete Setup Guide

