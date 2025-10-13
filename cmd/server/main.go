package main

import (
	"fmt"
	"log"
	"net"
	"os"
	"os/signal"
	"syscall"

	"google.golang.org/grpc"
)

const (
	defaultGRPCPort = "50051"
	defaultHTTPPort = "8080"
)

func main() {
	log.Println("Starting Hub User Service...")

	// Get configuration from environment
	grpcPort := getEnv("GRPC_PORT", defaultGRPCPort)
	httpPort := getEnv("HTTP_PORT", defaultHTTPPort)

	log.Printf("Configuration:")
	log.Printf("  - gRPC Port: %s", grpcPort)
	log.Printf("  - HTTP Port: %s", httpPort)

	// Create gRPC server
	grpcServer := grpc.NewServer()

	// TODO: Register gRPC services here
	// Example: pb.RegisterAuthServiceServer(grpcServer, &authServer{})

	// Start gRPC server
	grpcListener, err := net.Listen("tcp", fmt.Sprintf(":%s", grpcPort))
	if err != nil {
		log.Fatalf("Failed to listen on gRPC port %s: %v", grpcPort, err)
	}

	log.Printf("gRPC server listening on port %s", grpcPort)

	// Start gRPC server in a goroutine
	go func() {
		if err := grpcServer.Serve(grpcListener); err != nil {
			log.Fatalf("Failed to serve gRPC: %v", err)
		}
	}()

	// TODO: Start HTTP server for health checks and metrics
	// go startHTTPServer(httpPort)

	// Wait for interrupt signal
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Println("Shutting down servers...")
	grpcServer.GracefulStop()
	log.Println("Servers stopped")
}

// getEnv gets an environment variable or returns a default value
func getEnv(key, defaultValue string) string {
	value := os.Getenv(key)
	if value == "" {
		return defaultValue
	}
	return value
}
