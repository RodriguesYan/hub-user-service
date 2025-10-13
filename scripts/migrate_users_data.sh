#!/bin/bash

# Hub User Service - User Data Migration Script
# This script migrates existing users from the monolith database to the user service database

set -e  # Exit on error

echo "🔄 Hub User Service - User Data Migration"
echo "=========================================="
echo ""

# Source database (monolith)
SOURCE_DB_HOST=${SOURCE_DB_HOST:-localhost}
SOURCE_DB_PORT=${SOURCE_DB_PORT:-5432}
SOURCE_DB_NAME=${SOURCE_DB_NAME:-hub_investments}
SOURCE_DB_USER=${SOURCE_DB_USER:-postgres}
SOURCE_DB_PASSWORD=${SOURCE_DB_PASSWORD:-postgres}

# Target database (user service)
TARGET_DB_HOST=${TARGET_DB_HOST:-localhost}
TARGET_DB_PORT=${TARGET_DB_PORT:-5432}
TARGET_DB_NAME=${TARGET_DB_NAME:-hub_user_service}
TARGET_DB_USER=${TARGET_DB_USER:-hub_user_service_user}
TARGET_DB_PASSWORD=${TARGET_DB_PASSWORD:-hub_user_service_pass}

echo "Configuration:"
echo "  Source: $SOURCE_DB_NAME @ $SOURCE_DB_HOST:$SOURCE_DB_PORT"
echo "  Target: $TARGET_DB_NAME @ $TARGET_DB_HOST:$TARGET_DB_PORT"
echo ""

# Verify source database connection
echo "1️⃣  Verifying source database connection..."
if PGPASSWORD=$SOURCE_DB_PASSWORD psql -h $SOURCE_DB_HOST -p $SOURCE_DB_PORT -U $SOURCE_DB_USER -d $SOURCE_DB_NAME -c '\q' 2>/dev/null; then
    echo "✅ Connected to source database: $SOURCE_DB_NAME"
else
    echo "❌ Cannot connect to source database: $SOURCE_DB_NAME"
    exit 1
fi
echo ""

# Verify target database connection
echo "2️⃣  Verifying target database connection..."
if PGPASSWORD=$TARGET_DB_PASSWORD psql -h $TARGET_DB_HOST -p $TARGET_DB_PORT -U $TARGET_DB_USER -d $TARGET_DB_NAME -c '\q' 2>/dev/null; then
    echo "✅ Connected to target database: $TARGET_DB_NAME"
else
    echo "❌ Cannot connect to target database: $TARGET_DB_NAME"
    echo "   Have you run ./scripts/setup_database.sh ?"
    exit 1
fi
echo ""

# Check if users table exists in target
echo "3️⃣  Checking target users table..."
TABLE_EXISTS=$(PGPASSWORD=$TARGET_DB_PASSWORD psql -h $TARGET_DB_HOST -p $TARGET_DB_PORT -U $TARGET_DB_USER -d $TARGET_DB_NAME -tAc "SELECT EXISTS (SELECT FROM information_schema.tables WHERE table_name = 'users');")
if [ "$TABLE_EXISTS" = "f" ]; then
    echo "❌ Users table does not exist in target database"
    echo "   Please run migrations first: make migrate-up"
    exit 1
fi
echo "✅ Users table exists in target database"
echo ""

# Count users in source database
echo "4️⃣  Counting users in source database..."
SOURCE_COUNT=$(PGPASSWORD=$SOURCE_DB_PASSWORD psql -h $SOURCE_DB_HOST -p $SOURCE_DB_PORT -U $SOURCE_DB_USER -d $SOURCE_DB_NAME -tAc "SELECT COUNT(*) FROM users;")
echo "📊 Found $SOURCE_COUNT users in source database"
echo ""

# Count existing users in target database
echo "5️⃣  Counting existing users in target database..."
TARGET_COUNT=$(PGPASSWORD=$TARGET_DB_PASSWORD psql -h $TARGET_DB_HOST -p $TARGET_DB_PORT -U $TARGET_DB_USER -d $TARGET_DB_NAME -tAc "SELECT COUNT(*) FROM users;")
echo "📊 Found $TARGET_COUNT users already in target database"
echo ""

if [ "$TARGET_COUNT" -gt 0 ]; then
    echo "⚠️  WARNING: Target database already contains users"
    echo "   Existing users will be skipped (ON CONFLICT DO NOTHING)"
    echo ""
    read -p "Continue with migration? (y/N): " -n 1 -r
    echo ""
    if [[ ! $REPLY =~ ^[Yy]$ ]]; then
        echo "❌ Migration cancelled"
        exit 1
    fi
fi

# Perform data migration
echo "6️⃣  Migrating user data..."
echo "   This may take a few moments..."
echo ""

# Create temporary file for migration SQL
MIGRATION_SQL=$(mktemp)

cat > $MIGRATION_SQL <<'EOF'
-- Migrate users from monolith to user service
-- Uses ON CONFLICT to skip duplicates

INSERT INTO users (id, email, name, password, created_at, updated_at)
SELECT 
    id, 
    email, 
    name, 
    password, 
    created_at, 
    updated_at
FROM users_temp
ON CONFLICT (email) DO NOTHING;

-- Drop temporary table
DROP TABLE users_temp;

-- Return migration statistics
SELECT 
    COUNT(*) as migrated_users,
    MIN(created_at) as oldest_user,
    MAX(created_at) as newest_user
FROM users;
EOF

# Export users from source database
echo "   📤 Exporting users from source database..."
PGPASSWORD=$SOURCE_DB_PASSWORD psql -h $SOURCE_DB_HOST -p $SOURCE_DB_PORT -U $SOURCE_DB_USER -d $SOURCE_DB_NAME -c "\COPY (SELECT id, email, name, password, created_at, updated_at FROM users) TO STDOUT WITH CSV HEADER" > /tmp/users_export.csv

# Import users to target database (to temporary table first)
echo "   📥 Importing users to target database..."
PGPASSWORD=$TARGET_DB_PASSWORD psql -h $TARGET_DB_HOST -p $TARGET_DB_PORT -U $TARGET_DB_USER -d $TARGET_DB_NAME <<EOF
-- Create temporary table
CREATE TEMP TABLE users_temp (
    id SERIAL,
    email VARCHAR(255),
    name VARCHAR(255),
    password VARCHAR(255),
    created_at TIMESTAMP WITH TIME ZONE,
    updated_at TIMESTAMP WITH TIME ZONE
);

-- Import CSV data
\COPY users_temp FROM '/tmp/users_export.csv' WITH CSV HEADER;
EOF

# Execute migration SQL
echo "   ✨ Finalizing migration..."
MIGRATION_RESULT=$(PGPASSWORD=$TARGET_DB_PASSWORD psql -h $TARGET_DB_HOST -p $TARGET_DB_PORT -U $TARGET_DB_USER -d $TARGET_DB_NAME -f $MIGRATION_SQL)

# Clean up
rm -f $MIGRATION_SQL
rm -f /tmp/users_export.csv

echo "✅ Migration complete!"
echo ""

# Show migration statistics
echo "📊 Migration Statistics:"
FINAL_COUNT=$(PGPASSWORD=$TARGET_DB_PASSWORD psql -h $TARGET_DB_HOST -p $TARGET_DB_PORT -U $TARGET_DB_USER -d $TARGET_DB_NAME -tAc "SELECT COUNT(*) FROM users;")
MIGRATED=$((FINAL_COUNT - TARGET_COUNT))
echo "   Source users: $SOURCE_COUNT"
echo "   Target users (before): $TARGET_COUNT"
echo "   Target users (after): $FINAL_COUNT"
echo "   Migrated: $MIGRATED users"
echo ""

# Verify data integrity
echo "7️⃣  Verifying data integrity..."
echo "   Checking for users with invalid emails..."
INVALID_EMAILS=$(PGPASSWORD=$TARGET_DB_PASSWORD psql -h $TARGET_DB_HOST -p $TARGET_DB_PORT -U $TARGET_DB_USER -d $TARGET_DB_NAME -tAc "SELECT COUNT(*) FROM users WHERE email IS NULL OR email = '';")
if [ "$INVALID_EMAILS" -eq 0 ]; then
    echo "   ✅ All users have valid emails"
else
    echo "   ⚠️  WARNING: Found $INVALID_EMAILS users with invalid emails"
fi

echo "   Checking for users with missing passwords..."
MISSING_PASSWORDS=$(PGPASSWORD=$TARGET_DB_PASSWORD psql -h $TARGET_DB_HOST -p $TARGET_DB_PORT -U $TARGET_DB_USER -d $TARGET_DB_NAME -tAc "SELECT COUNT(*) FROM users WHERE password IS NULL OR password = '';")
if [ "$MISSING_PASSWORDS" -eq 0 ]; then
    echo "   ✅ All users have passwords"
else
    echo "   ⚠️  WARNING: Found $MISSING_PASSWORDS users with missing passwords"
fi
echo ""

echo "🎉 Data migration complete!"
echo ""
echo "📋 Next steps:"
echo "   1. Test user service with migrated data"
echo "   2. Verify JWT token generation works"
echo "   3. Test login with existing users"
echo "   4. Monitor for any issues"
echo ""
echo "⚠️  IMPORTANT:"
echo "   - Monolith database still contains users table (unchanged)"
echo "   - User service is now using its own database"
echo "   - After validation, monolith users table can be deprecated"
echo ""

