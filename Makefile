.PHONY: help build run test test-coverage clean proto migrate-up migrate-down setup-db migrate-data setup-all docker-build docker-run lint fmt

# Variables
SERVICE_NAME=hub-user-service
DOCKER_IMAGE=$(SERVICE_NAME):latest
PROTO_DIR=internal/grpc/proto
MIGRATION_DIR=migrations
DATABASE_URL?=postgresql://localhost:5432/hub_users?sslmode=disable

help: ## Show this help message
	@echo "Usage: make [target]"
	@echo ""
	@echo "Available targets:"
	@awk 'BEGIN {FS = ":.*?## "} /^[a-zA-Z_-]+:.*?## / {printf "  %-20s %s\n", $$1, $$2}' $(MAKEFILE_LIST)

build: ## Build the service binary
	@echo "Building $(SERVICE_NAME)..."
	@go build -o bin/$(SERVICE_NAME) cmd/server/main.go
	@echo "Build complete: bin/$(SERVICE_NAME)"

run: ## Run the service locally
	@echo "Running $(SERVICE_NAME)..."
	@go run cmd/server/main.go

test: ## Run all tests
	@echo "Running tests..."
	@go test ./... -v

test-coverage: ## Run tests with coverage
	@echo "Running tests with coverage..."
	@go test ./... -cover -coverprofile=coverage.out
	@go tool cover -html=coverage.out -o coverage.html
	@echo "Coverage report generated: coverage.html"

test-race: ## Run tests with race detector
	@echo "Running tests with race detector..."
	@go test -race ./...

clean: ## Clean build artifacts
	@echo "Cleaning..."
	@rm -rf bin/
	@rm -f coverage.out coverage.html
	@echo "Clean complete"

proto: ## Generate gRPC code from proto files
	@echo "Generating gRPC code..."
	@protoc --go_out=. --go_opt=paths=source_relative \
		--go-grpc_out=. --go-grpc_opt=paths=source_relative \
		$(PROTO_DIR)/*.proto
	@echo "gRPC code generated"

migrate-up: ## Run database migrations
	@echo "Running migrations..."
	@migrate -path $(MIGRATION_DIR) -database "$(DATABASE_URL)" up
	@echo "Migrations complete"

migrate-down: ## Rollback database migrations
	@echo "Rolling back migrations..."
	@migrate -path $(MIGRATION_DIR) -database "$(DATABASE_URL)" down
	@echo "Rollback complete"

migrate-create: ## Create a new migration file (usage: make migrate-create NAME=create_users)
	@echo "Creating migration: $(NAME)"
	@migrate create -ext sql -dir $(MIGRATION_DIR) -seq $(NAME)
	@echo "Migration files created in $(MIGRATION_DIR)"

setup-db: ## Setup separate database for user service
	@echo "Setting up database..."
	@./scripts/setup_database.sh

migrate-data: ## Migrate user data from monolith
	@echo "Migrating user data..."
	@./scripts/migrate_users_data.sh

setup-all: setup-db migrate-up migrate-data ## Complete setup (database + migrations + data)
	@echo "✅ Complete setup finished!"

docker-build: ## Build Docker image
	@echo "Building Docker image..."
	@docker build -t $(DOCKER_IMAGE) .
	@echo "Docker image built: $(DOCKER_IMAGE)"

docker-run: ## Run Docker container
	@echo "Running Docker container..."
	@docker run -p 50051:50051 $(DOCKER_IMAGE)

docker-compose-up: ## Start services with docker-compose
	@echo "Starting services..."
	@docker-compose up -d
	@echo "Services started"

docker-compose-down: ## Stop services with docker-compose
	@echo "Stopping services..."
	@docker-compose down
	@echo "Services stopped"

lint: ## Run linter
	@echo "Running linter..."
	@golangci-lint run ./...
	@echo "Linting complete"

fmt: ## Format code
	@echo "Formatting code..."
	@go fmt ./...
	@echo "Formatting complete"

tidy: ## Tidy go modules
	@echo "Tidying modules..."
	@go mod tidy
	@echo "Modules tidied"

deps: ## Download dependencies
	@echo "Downloading dependencies..."
	@go mod download
	@echo "Dependencies downloaded"

.DEFAULT_GOAL := help

