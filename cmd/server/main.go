package main

import (
	"context"
	"log"
	"net"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	pb "hub-user-service/proto/pb"
	"hub-user-service/shared/container"

	"google.golang.org/grpc"
)

func main() {
	log.Println("🚀 Starting Hub User Management Service...")

	// Initialize dependency injection container
	c, err := container.NewContainer()
	if err != nil {
		log.Fatalf("❌ Failed to initialize container: %v", err)
	}
	defer c.Close()

	// Start gRPC server in a goroutine
	grpcServer := grpc.NewServer()
	pb.RegisterUserServiceServer(grpcServer, c.UserGRPCServer)

	grpcListener, err := net.Listen("tcp", ":"+c.Config.GRPCPort)
	if err != nil {
		log.Fatalf("❌ Failed to listen on gRPC port: %v", err)
	}

	go func() {
		log.Printf("✅ gRPC server listening on port %s", c.Config.GRPCPort)
		if err := grpcServer.Serve(grpcListener); err != nil {
			log.Fatalf("❌ gRPC server failed: %v", err)
		}
	}()

	// Setup HTTP server
	mux := http.NewServeMux()

	// Register HTTP endpoints
	mux.HandleFunc("/health", c.UserHTTPHandler.HealthCheck)
	mux.HandleFunc("/login", c.UserHTTPHandler.Login)
	mux.HandleFunc("/register", c.UserHTTPHandler.Register)
	mux.HandleFunc("/profile", c.UserHTTPHandler.AuthMiddleware(c.UserHTTPHandler.GetProfile))
	mux.HandleFunc("/validate-token", c.UserHTTPHandler.ValidateToken)

	httpServer := &http.Server{
		Addr:         ":" + c.Config.HTTPPort,
		Handler:      mux,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	// Start HTTP server in a goroutine
	go func() {
		log.Printf("✅ HTTP server listening on port %s", c.Config.HTTPPort)
		log.Println("📋 Available endpoints:")
		log.Println("   POST   /login")
		log.Println("   POST   /register")
		log.Println("   GET    /profile (protected)")
		log.Println("   POST   /validate-token")
		log.Println("   GET    /health")

		if err := httpServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("❌ HTTP server failed: %v", err)
		}
	}()

	log.Println("✨ Hub User Management Service started successfully!")
	log.Printf("   Environment: %s", c.Config.Environment)
	log.Printf("   Service Version: %s", c.Config.ServiceVersion)

	// Wait for interrupt signal to gracefully shutdown
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Println("🛑 Shutting down servers...")

	// Graceful shutdown for HTTP server
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := httpServer.Shutdown(ctx); err != nil {
		log.Printf("❌ HTTP server forced to shutdown: %v", err)
	}

	// Graceful shutdown for gRPC server
	grpcServer.GracefulStop()

	log.Println("✅ Servers stopped gracefully")
}
