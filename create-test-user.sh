#!/bin/bash

# Script to create a test user in the database
# Uses bcrypt to hash the password properly

set -e

echo "🔐 Create Test User for Hub User Service"
echo "=========================================="
echo ""

# Default values
DEFAULT_EMAIL="test@example.com"
DEFAULT_PASSWORD="password123"
DEFAULT_NAME="Test User"

# Get user input
read -p "Email [$DEFAULT_EMAIL]: " EMAIL
EMAIL=${EMAIL:-$DEFAULT_EMAIL}

read -p "Password [$DEFAULT_PASSWORD]: " PASSWORD
PASSWORD=${PASSWORD:-$DEFAULT_PASSWORD}

read -p "Name [$DEFAULT_NAME]: " NAME
NAME=${NAME:-$DEFAULT_NAME}

echo ""
echo "Creating user:"
echo "  Email: $EMAIL"
echo "  Name: $NAME"
echo "  Password: $PASSWORD"
echo ""

# Check if user already exists
EXISTING=$(psql -d hubinvestments -tAc "SELECT COUNT(*) FROM users WHERE email='$EMAIL';")

if [ "$EXISTING" -gt 0 ]; then
    echo "⚠️  User with email '$EMAIL' already exists!"
    read -p "Delete and recreate? (y/n): " RECREATE
    if [ "$RECREATE" = "y" ]; then
        psql -d hubinvestments -c "DELETE FROM users WHERE email='$EMAIL';"
        echo "✅ Deleted existing user"
    else
        echo "❌ Aborted"
        exit 1
    fi
fi

# Create a simple Go program to hash the password
cat > /tmp/hash_password.go << 'EOF'
package main

import (
    "fmt"
    "os"
    "golang.org/x/crypto/bcrypt"
)

func main() {
    if len(os.Args) < 2 {
        fmt.Println("Usage: hash_password <password>")
        os.Exit(1)
    }
    
    password := os.Args[1]
    hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
    if err != nil {
        fmt.Fprintf(os.Stderr, "Error: %v\n", err)
        os.Exit(1)
    }
    
    fmt.Print(string(hash))
}
EOF

# Hash the password
echo "Hashing password..."
cd /tmp
go mod init hash_password 2>/dev/null || true
go get golang.org/x/crypto/bcrypt 2>/dev/null || true
HASHED_PASSWORD=$(go run hash_password.go "$PASSWORD")

if [ -z "$HASHED_PASSWORD" ]; then
    echo "❌ Failed to hash password"
    exit 1
fi

# Insert user into database
echo "Inserting user into database..."
psql -d hubinvestments << EOF
INSERT INTO users (email, name, password)
VALUES ('$EMAIL', '$NAME', '$HASHED_PASSWORD')
RETURNING id, email, name, created_at;
EOF

echo ""
echo "✅ User created successfully!"
echo ""
echo "You can now login with:"
echo "  Email: $EMAIL"
echo "  Password: $PASSWORD"
echo ""
echo "Test with curl:"
echo "curl -X POST http://localhost:8080/api/v1/login \\"
echo "  -H \"Content-Type: application/json\" \\"
echo "  -d '{\"email\":\"$EMAIL\",\"password\":\"$PASSWORD\"}'"
echo ""

# Cleanup
rm -f /tmp/hash_password.go

