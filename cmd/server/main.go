package main

import (
	"log"
	"net"
	"os"

	"hub-user-service/internal/auth"
	"hub-user-service/internal/auth/token"
	"hub-user-service/internal/config"
	"hub-user-service/internal/database"
	grpcServer "hub-user-service/internal/grpc"
	"hub-user-service/internal/grpc/proto"
	"hub-user-service/internal/login/application/usecase"
	"hub-user-service/internal/login/infra/persistence"

	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
)

func main() {
	// Load configuration
	cfg := config.Load()
	log.Printf("Starting Hub User Service...")
	log.Printf("gRPC Port: %s", cfg.GRPCPort)
	log.Printf("HTTP Port: %s", cfg.HTTPPort)
	log.Printf("Database URL: %s", maskDatabaseURL(cfg.DatabaseURL))

	// Initialize database connection
	dbConfig := database.ConnectionConfig{
		Driver:   "postgres",
		Host:     getEnvWithDefault("DB_HOST", "localhost"),
		Port:     getEnvWithDefault("DB_PORT", "5432"),
		Database: getEnvWithDefault("DB_NAME", "hub_investments"),
		Username: getEnvWithDefault("DB_USER", "postgres"),
		Password: getEnvWithDefault("DB_PASSWORD", "postgres"),
		SSLMode:  getEnvWithDefault("DB_SSLMODE", "disable"),
	}

	db, err := database.NewConnectionFactory(dbConfig).CreateConnection()
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}
	log.Println("✅ Database connected successfully")

	// Initialize repositories
	loginRepository := persistence.NewLoginRepository(db)
	log.Println("✅ Login repository initialized")

	// Initialize use cases
	loginUsecase := usecase.NewDoLoginUsecase(loginRepository)
	log.Println("✅ Login use case initialized")

	// Initialize authentication services
	tokenService := token.NewTokenService()
	authService := auth.NewAuthService(tokenService)
	log.Println("✅ Auth service initialized")

	// Initialize gRPC server
	authGrpcServer := grpcServer.NewAuthServer(loginUsecase, authService)
	log.Println("✅ gRPC auth server initialized")

	// Create gRPC server with options
	grpcSrv := grpc.NewServer()

	// Register services
	proto.RegisterAuthServiceServer(grpcSrv, authGrpcServer)
	log.Println("✅ AuthService registered")

	// Register reflection service (useful for gRPC clients like grpcurl)
	reflection.Register(grpcSrv)
	log.Println("✅ gRPC reflection registered")

	// Start listening
	listener, err := net.Listen("tcp", cfg.GRPCPort)
	if err != nil {
		log.Fatalf("Failed to listen on port %s: %v", cfg.GRPCPort, err)
	}

	log.Printf("🚀 Hub User Service gRPC server listening on %s", cfg.GRPCPort)
	log.Println("📡 Ready to accept connections...")

	// Start serving (blocking call)
	if err := grpcSrv.Serve(listener); err != nil {
		log.Fatalf("Failed to serve gRPC: %v", err)
	}
}

// getEnvWithDefault gets an environment variable or returns a default value
func getEnvWithDefault(key, defaultValue string) string {
	value := os.Getenv(key)
	if value == "" {
		return defaultValue
	}
	return value
}

// maskDatabaseURL masks sensitive information in database URL for logging
func maskDatabaseURL(url string) string {
	if url == "" {
		return "not configured"
	}
	return "***configured***"
}
