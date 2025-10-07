.PHONY: help build run test clean docker-up docker-down proto migrate-up migrate-down

help: ## Show this help message
	@echo 'Usage: make [target]'
	@echo ''
	@echo 'Available targets:'
	@awk 'BEGIN {FS = ":.*?## "} /^[a-zA-Z_-]+:.*?## / {printf "  %-15s %s\n", $$1, $$2}' $(MAKEFILE_LIST)

build: ## Build the service binary
	@echo "Building hub-user-service..."
	@go build -o bin/hub-user-service ./cmd/server

run: ## Run the service locally
	@echo "Running hub-user-service..."
	@go run ./cmd/server/main.go

test: ## Run tests
	@echo "Running tests..."
	@go test -v -race -coverprofile=coverage.out ./...

test-coverage: test ## Run tests with coverage report
	@go tool cover -html=coverage.out

clean: ## Clean build artifacts
	@echo "Cleaning..."
	@rm -rf bin/
	@rm -f coverage.out

docker-up: ## Start all services with Docker Compose
	@echo "Starting services with Docker Compose..."
	@docker-compose up -d

docker-down: ## Stop all services
	@echo "Stopping services..."
	@docker-compose down

docker-logs: ## Show service logs
	@docker-compose logs -f user-service

docker-rebuild: ## Rebuild and restart services
	@echo "Rebuilding services..."
	@docker-compose up -d --build

proto: ## Generate protobuf code
	@echo "Generating protobuf code..."
	@mkdir -p proto/pb
	@protoc --go_out=proto/pb --go_opt=paths=source_relative --go-grpc_out=proto/pb --go-grpc_opt=paths=source_relative proto/user_service.proto

migrate-up: ## Run database migrations up
	@echo "Running migrations up..."
	@psql $(DATABASE_URL) -f migrations/000001_create_users_table.up.sql

migrate-down: ## Run database migrations down
	@echo "Running migrations down..."
	@psql $(DATABASE_URL) -f migrations/000001_create_users_table.down.sql

lint: ## Run linter
	@echo "Running linter..."
	@golangci-lint run

deps: ## Download dependencies
	@echo "Downloading dependencies..."
	@go mod download
	@go mod tidy

.DEFAULT_GOAL := help
