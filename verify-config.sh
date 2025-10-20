#!/bin/bash

# Script to verify and fix hub-user-service configuration

set -e

echo "🔍 Hub User Service Configuration Checker"
echo "=========================================="
echo ""

# Colors
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Check if config.env exists
if [ ! -f "config.env" ]; then
    echo -e "${RED}❌ config.env not found!${NC}"
    echo ""
    echo "Creating config.env from config.env.example..."
    cp config.env.example config.env
    echo -e "${GREEN}✅ Created config.env${NC}"
    echo ""
    echo -e "${YELLOW}⚠️  Please edit config.env and update the database credentials${NC}"
    exit 0
fi

echo -e "${GREEN}✅ config.env exists${NC}"
echo ""

# Load config.env
source config.env

# Check JWT Secret
echo "1. Checking JWT Secret..."
if [ -z "$MY_JWT_SECRET" ]; then
    echo -e "${RED}   ❌ MY_JWT_SECRET is not set${NC}"
else
    echo -e "${GREEN}   ✅ MY_JWT_SECRET is set (${#MY_JWT_SECRET} characters)${NC}"
fi

# Check Server Ports
echo ""
echo "2. Checking Server Configuration..."
echo -e "${BLUE}   HTTP Port: ${HTTP_PORT:-not set}${NC}"
echo -e "${BLUE}   gRPC Port: ${GRPC_PORT:-not set}${NC}"

# Check Database Configuration
echo ""
echo "3. Checking Database Configuration..."
if [ -z "$DB_HOST" ]; then
    echo -e "${RED}   ❌ DB_HOST is not set${NC}"
else
    echo -e "${GREEN}   ✅ DB_HOST: $DB_HOST${NC}"
fi

if [ -z "$DB_PORT" ]; then
    echo -e "${RED}   ❌ DB_PORT is not set${NC}"
else
    echo -e "${GREEN}   ✅ DB_PORT: $DB_PORT${NC}"
fi

if [ -z "$DB_NAME" ]; then
    echo -e "${RED}   ❌ DB_NAME is not set${NC}"
else
    echo -e "${GREEN}   ✅ DB_NAME: $DB_NAME${NC}"
fi

if [ -z "$DB_USER" ]; then
    echo -e "${RED}   ❌ DB_USER is not set${NC}"
else
    echo -e "${GREEN}   ✅ DB_USER: $DB_USER${NC}"
fi

# Check PostgreSQL
echo ""
echo "4. Checking PostgreSQL..."

if ! command -v psql &> /dev/null; then
    echo -e "${RED}   ❌ psql command not found${NC}"
    echo -e "${YELLOW}   Install PostgreSQL: brew install postgresql@14${NC}"
    exit 1
fi

echo -e "${GREEN}   ✅ psql is installed${NC}"

# Try to connect to PostgreSQL
echo ""
echo "5. Testing Database Connection..."

# Check if database exists
if psql -lqt 2>/dev/null | cut -d \| -f 1 | grep -qw "$DB_NAME"; then
    echo -e "${GREEN}   ✅ Database '$DB_NAME' exists${NC}"
else
    echo -e "${YELLOW}   ⚠️  Database '$DB_NAME' does not exist${NC}"
    read -p "   Create database '$DB_NAME'? (y/n): " create_db
    if [ "$create_db" = "y" ]; then
        createdb "$DB_NAME" 2>/dev/null && echo -e "${GREEN}   ✅ Created database '$DB_NAME'${NC}" || echo -e "${RED}   ❌ Failed to create database${NC}"
    fi
fi

# Check if user exists
if psql postgres -tAc "SELECT 1 FROM pg_roles WHERE rolname='$DB_USER'" 2>/dev/null | grep -q 1; then
    echo -e "${GREEN}   ✅ PostgreSQL user '$DB_USER' exists${NC}"
else
    echo -e "${YELLOW}   ⚠️  PostgreSQL user '$DB_USER' does not exist${NC}"
    
    # Get current user
    CURRENT_USER=$(whoami)
    
    if [ "$DB_USER" = "postgres" ]; then
        read -p "   Create 'postgres' user? (y/n): " create_user
        if [ "$create_user" = "y" ]; then
            psql postgres -c "CREATE ROLE postgres WITH LOGIN PASSWORD 'postgres' SUPERUSER;" 2>/dev/null && \
                echo -e "${GREEN}   ✅ Created user 'postgres'${NC}" || \
                echo -e "${RED}   ❌ Failed to create user${NC}"
        fi
    else
        echo -e "${YELLOW}   Suggestion: Change DB_USER in config.env to '$CURRENT_USER'${NC}"
        read -p "   Update DB_USER to '$CURRENT_USER'? (y/n): " update_user
        if [ "$update_user" = "y" ]; then
            sed -i.bak "s/^DB_USER=.*/DB_USER=$CURRENT_USER/" config.env
            sed -i.bak "s/^DB_PASSWORD=.*/DB_PASSWORD=/" config.env
            rm -f config.env.bak
            echo -e "${GREEN}   ✅ Updated config.env${NC}"
            echo -e "${BLUE}   DB_USER=$CURRENT_USER${NC}"
            echo -e "${BLUE}   DB_PASSWORD=(empty)${NC}"
        fi
    fi
fi

# Test actual connection
echo ""
echo "6. Testing Connection String..."
CONN_STRING="host=$DB_HOST port=$DB_PORT user=$DB_USER dbname=$DB_NAME sslmode=${DB_SSLMODE:-disable}"
if [ -n "$DB_PASSWORD" ]; then
    CONN_STRING="$CONN_STRING password=$DB_PASSWORD"
fi

if psql "$CONN_STRING" -c "SELECT 1;" &>/dev/null; then
    echo -e "${GREEN}   ✅ Successfully connected to database!${NC}"
else
    echo -e "${RED}   ❌ Failed to connect to database${NC}"
    echo -e "${YELLOW}   Connection string: $CONN_STRING${NC}"
    echo ""
    echo -e "${YELLOW}   Troubleshooting:${NC}"
    echo "   1. Make sure PostgreSQL is running: brew services start postgresql@14"
    echo "   2. Check your database credentials in config.env"
    echo "   3. Try: psql -U $DB_USER -d $DB_NAME"
fi

echo ""
echo "=========================================="
echo -e "${GREEN}✅ Configuration check complete!${NC}"
echo ""
echo "To start the service:"
echo "  make run"
echo ""

