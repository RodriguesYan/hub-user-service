#!/bin/bash

# Hub User Service - Database Setup Script
# This script creates the separate database for the user service

set -e  # Exit on error

echo "🚀 Hub User Service - Database Setup"
echo "======================================"
echo ""

# Load configuration from environment or use defaults
DB_HOST=${DB_HOST:-localhost}
DB_PORT=${DB_PORT:-5432}
POSTGRES_USER=${POSTGRES_USER:-postgres}
POSTGRES_PASSWORD=${POSTGRES_PASSWORD:-postgres}

# New database configuration
NEW_DB_NAME="hub_user_service"
NEW_DB_USER="hub_user_service_user"
NEW_DB_PASSWORD="hub_user_service_pass"

echo "Configuration:"
echo "  PostgreSQL Host: $DB_HOST"
echo "  PostgreSQL Port: $DB_PORT"
echo "  New Database: $NEW_DB_NAME"
echo "  New User: $NEW_DB_USER"
echo ""

# Check if PostgreSQL is running
echo "1️⃣  Checking PostgreSQL connection..."
if PGPASSWORD=$POSTGRES_PASSWORD psql -h $DB_HOST -p $DB_PORT -U $POSTGRES_USER -lqt | cut -d \| -f 1 | grep -qw template1; then
    echo "✅ PostgreSQL is running"
else
    echo "❌ Cannot connect to PostgreSQL"
    echo "   Please ensure PostgreSQL is running on $DB_HOST:$DB_PORT"
    exit 1
fi
echo ""

# Create database user
echo "2️⃣  Creating database user..."
PGPASSWORD=$POSTGRES_PASSWORD psql -h $DB_HOST -p $DB_PORT -U $POSTGRES_USER -tc "SELECT 1 FROM pg_user WHERE usename = '$NEW_DB_USER'" | grep -q 1 || \
PGPASSWORD=$POSTGRES_PASSWORD psql -h $DB_HOST -p $DB_PORT -U $POSTGRES_USER <<EOF
CREATE USER $NEW_DB_USER WITH PASSWORD '$NEW_DB_PASSWORD';
ALTER USER $NEW_DB_USER WITH CREATEDB;
EOF
echo "✅ Database user created: $NEW_DB_USER"
echo ""

# Create database
echo "3️⃣  Creating database..."
PGPASSWORD=$POSTGRES_PASSWORD psql -h $DB_HOST -p $DB_PORT -U $POSTGRES_USER -tc "SELECT 1 FROM pg_database WHERE datname = '$NEW_DB_NAME'" | grep -q 1 || \
PGPASSWORD=$POSTGRES_PASSWORD psql -h $DB_HOST -p $DB_PORT -U $POSTGRES_USER <<EOF
CREATE DATABASE $NEW_DB_NAME OWNER $NEW_DB_USER;
GRANT ALL PRIVILEGES ON DATABASE $NEW_DB_NAME TO $NEW_DB_USER;
EOF
echo "✅ Database created: $NEW_DB_NAME"
echo ""

# Grant permissions
echo "4️⃣  Granting permissions..."
PGPASSWORD=$POSTGRES_PASSWORD psql -h $DB_HOST -p $DB_PORT -U $POSTGRES_USER -d $NEW_DB_NAME <<EOF
GRANT ALL ON SCHEMA public TO $NEW_DB_USER;
ALTER DEFAULT PRIVILEGES IN SCHEMA public GRANT ALL ON TABLES TO $NEW_DB_USER;
ALTER DEFAULT PRIVILEGES IN SCHEMA public GRANT ALL ON SEQUENCES TO $NEW_DB_USER;
EOF
echo "✅ Permissions granted"
echo ""

echo "🎉 Database setup complete!"
echo ""
echo "📋 Connection details:"
echo "   Database: $NEW_DB_NAME"
echo "   User: $NEW_DB_USER"
echo "   Password: $NEW_DB_PASSWORD"
echo "   Host: $DB_HOST"
echo "   Port: $DB_PORT"
echo ""
echo "📝 Update your config.env with:"
echo "   DB_HOST=$DB_HOST"
echo "   DB_PORT=$DB_PORT"
echo "   DB_NAME=$NEW_DB_NAME"
echo "   DB_USER=$NEW_DB_USER"
echo "   DB_PASSWORD=$NEW_DB_PASSWORD"
echo ""
echo "⏭️  Next steps:"
echo "   1. Run migrations: make migrate-up"
echo "   2. Migrate data: ./scripts/migrate_users_data.sh"
echo ""

