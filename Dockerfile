# Build stage
FROM golang:1.25.1-alpine AS builder

# Build arguments for versioning
ARG VERSION=dev
ARG BUILD_DATE=unknown
ARG GIT_COMMIT=unknown
ARG TARGETARCH

# Install build dependencies
RUN apk add --no-cache git make file postgresql-client

# Set working directory
WORKDIR /workspace

# Copy proto contracts first (from parent context)
COPY hub-proto-contracts ./hub-proto-contracts

# Copy go mod files
COPY hub-user-service/go.mod hub-user-service/go.sum ./hub-user-service/

# Set working directory to service
WORKDIR /workspace/hub-user-service

# Download dependencies
RUN go mod download

# Copy source code
COPY hub-user-service/ .

# Build the application
RUN CGO_ENABLED=0 GOOS=linux GOARCH=${TARGETARCH:-amd64} \
    go build -a -installsuffix cgo \
    -ldflags="-w -s -X main.Version=${VERSION} -X main.BuildDate=${BUILD_DATE} -X main.GitCommit=${GIT_COMMIT}" \
    -o /app/hub-user-service \
    cmd/server/main.go

# Verify binary
RUN ls -lh /app/hub-user-service && file /app/hub-user-service

# Runtime stage
FROM alpine:latest

# OCI labels
LABEL org.opencontainers.image.title="Hub User Service"
LABEL org.opencontainers.image.description="User authentication and management service"
LABEL org.opencontainers.image.version="${VERSION}"
LABEL org.opencontainers.image.created="${BUILD_DATE}"
LABEL org.opencontainers.image.revision="${GIT_COMMIT}"

# Install runtime dependencies
RUN apk --no-cache add ca-certificates wget postgresql-client tzdata

# Create non-root user
RUN addgroup -g 1000 appuser && \
    adduser -D -u 1000 -G appuser appuser

# Set working directory
WORKDIR /app

# Copy binary from builder
COPY --from=builder /app/hub-user-service .

# Copy migrations
COPY --from=builder /workspace/hub-user-service/migrations ./migrations

# Create logs directory
RUN mkdir -p /app/logs

# Change ownership to non-root user
RUN chown -R appuser:appuser /app

# Switch to non-root user
USER appuser

# Expose ports
EXPOSE 50051 8080

# Health check
HEALTHCHECK --interval=30s --timeout=3s --start-period=10s --retries=3 \
  CMD wget --no-verbose --tries=1 --spider http://localhost:8080/health || exit 1

# Run the application
CMD ["./hub-user-service"]

