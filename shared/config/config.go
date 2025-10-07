package config

import (
	"log"
	"os"
	"sync"
	"time"

	"github.com/joho/godotenv"
)

// Config holds all application configuration
type Config struct {
	// Server configuration
	HTTPPort string
	GRPCPort string

	// Security configuration
	JWTSecret       string
	TokenExpiration time.Duration

	// Database configuration
	DatabaseURL string
	DBHost      string
	DBPort      string
	DBUser      string
	DBPassword  string
	DBName      string

	// Service configuration
	ServiceName    string
	ServiceVersion string
	Environment    string
}

var (
	instance *Config
	once     sync.Once
)

// Load loads configuration from environment variables
func Load() *Config {
	once.Do(func() {
		// Try to load .env file
		if err := godotenv.Load(); err != nil {
			log.Printf("Warning: Could not load .env file: %v", err)
			log.Println("Using environment variables or default values...")
		}

		instance = &Config{
			HTTPPort:        getEnvWithDefault("HTTP_PORT", "8080"),
			GRPCPort:        getEnvWithDefault("GRPC_PORT", "50051"),
			JWTSecret:       getEnvWithDefault("JWT_SECRET", "default-secret-change-in-production"),
			TokenExpiration: parseDuration(getEnvWithDefault("TOKEN_EXPIRATION", "10m")),
			DatabaseURL:     getEnvWithDefault("DATABASE_URL", ""),
			DBHost:          getEnvWithDefault("DB_HOST", "localhost"),
			DBPort:          getEnvWithDefault("DB_PORT", "5432"),
			DBUser:          getEnvWithDefault("DB_USER", "postgres"),
			DBPassword:      getEnvWithDefault("DB_PASSWORD", "postgres"),
			DBName:          getEnvWithDefault("DB_NAME", "hub_users_db"),
			ServiceName:     "hub-user-service",
			ServiceVersion:  "1.0.0",
			Environment:     getEnvWithDefault("ENVIRONMENT", "development"),
		}

		// Build DATABASE_URL if not provided
		if instance.DatabaseURL == "" {
			instance.DatabaseURL = "postgres://" + instance.DBUser + ":" + instance.DBPassword +
				"@" + instance.DBHost + ":" + instance.DBPort + "/" + instance.DBName + "?sslmode=disable"
		}

		// Validate configuration
		if instance.JWTSecret == "default-secret-change-in-production" {
			log.Println("⚠️  WARNING: Using default JWT secret. Set JWT_SECRET environment variable for production!")
		}
	})

	return instance
}

// Get returns the current configuration instance
func Get() *Config {
	if instance == nil {
		return Load()
	}
	return instance
}

// getEnvWithDefault gets an environment variable with a fallback default value
func getEnvWithDefault(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

// parseDuration parses a duration string with default fallback
func parseDuration(durationStr string) time.Duration {
	duration, err := time.ParseDuration(durationStr)
	if err != nil {
		log.Printf("Warning: Invalid duration format '%s', using default 10m", durationStr)
		return 10 * time.Minute
	}
	return duration
}

// IsProduction checks if the application is running in production mode
func (c *Config) IsProduction() bool {
	return c.Environment == "production"
}

// IsDevelopment checks if the application is running in development mode
func (c *Config) IsDevelopment() bool {
	return c.Environment == "development"
}
