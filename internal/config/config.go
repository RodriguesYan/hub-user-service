package config

import (
	"fmt"
	"log"
	"os"
	"sync"

	"github.com/joho/godotenv"
)

// Config holds all application configuration
type Config struct {
	// Server Configuration
	HTTPPort string
	GRPCPort string

	// JWT Configuration (MUST match monolith for token compatibility)
	JWTSecret string

	// Database Configuration
	DatabaseURL string
	DBHost      string
	DBPort      string
	DBName      string
	DBUser      string
	DBPassword  string
	DBSSLMode   string

	// Redis Configuration (optional for caching)
	RedisHost string
	RedisPort string

	// Environment
	Environment string
}

var (
	instance *Config
	once     sync.Once
)

// Load loads configuration from environment variables or config file
// This function is thread-safe and will only load configuration once
func Load() *Config {
	once.Do(func() {
		// Try to load from config.env file
		err := godotenv.Load("config.env")
		if err != nil {
			log.Printf("Warning: Could not load config.env file: %v", err)
			log.Println("Using environment variables or default values...")
		}

		instance = &Config{
			// Server Configuration
			HTTPPort: getEnvWithDefault("HTTP_PORT", "localhost:8080"),
			GRPCPort: getEnvWithDefault("GRPC_PORT", "localhost:50051"),

			// JWT Configuration (MUST match monolith)
			JWTSecret: getEnvWithDefault("MY_JWT_SECRET", "default-secret-key-change-in-production"),

			// Database Configuration
			DatabaseURL: getEnvWithDefault("DATABASE_URL", ""),
			DBHost:      getEnvWithDefault("DB_HOST", "localhost"),
			DBPort:      getEnvWithDefault("DB_PORT", "5432"),
			DBName:      getEnvWithDefault("DB_NAME", "hub_investments"),
			DBUser:      getEnvWithDefault("DB_USER", "postgres"),
			DBPassword:  getEnvWithDefault("DB_PASSWORD", "postgres"),
			DBSSLMode:   getEnvWithDefault("DB_SSLMODE", "disable"),

			// Redis Configuration (optional)
			RedisHost: getEnvWithDefault("REDIS_HOST", "localhost"),
			RedisPort: getEnvWithDefault("REDIS_PORT", "6379"),

			// Environment
			Environment: getEnvWithDefault("ENVIRONMENT", "development"),
		}

		// Validate required configuration
		if instance.JWTSecret == "default-secret-key-change-in-production" {
			log.Println("⚠️  WARNING: Using default JWT secret. Please set MY_JWT_SECRET environment variable for production.")
			log.Println("⚠️  WARNING: JWT tokens will NOT be compatible with monolith unless secrets match!")
		}

		// Validate database configuration
		if instance.DBHost == "" || instance.DBName == "" {
			log.Println("⚠️  WARNING: Database configuration incomplete. Service may not start correctly.")
		}

		// Log configuration (mask sensitive data)
		log.Printf("Configuration loaded:")
		log.Printf("  Environment: %s", instance.Environment)
		log.Printf("  HTTP Port: %s", instance.HTTPPort)
		log.Printf("  gRPC Port: %s", instance.GRPCPort)
		log.Printf("  Database: %s:%s/%s", instance.DBHost, instance.DBPort, instance.DBName)
		log.Printf("  JWT Secret: %s", maskSecret(instance.JWTSecret))
		log.Printf("  Redis: %s:%s", instance.RedisHost, instance.RedisPort)
	})

	return instance
}

// Get returns the current configuration instance
// If configuration hasn't been loaded yet, it will load it first
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

// IsProduction checks if the application is running in production mode
func (c *Config) IsProduction() bool {
	return os.Getenv("ENVIRONMENT") == "production"
}

// GetRedisAddress returns the complete Redis address
func (c *Config) GetRedisAddress() string {
	return c.RedisHost + ":" + c.RedisPort
}

// GetDatabaseConnectionString returns a PostgreSQL connection string
func (c *Config) GetDatabaseConnectionString() string {
	if c.DatabaseURL != "" {
		return c.DatabaseURL
	}
	return "host=" + c.DBHost +
		" port=" + c.DBPort +
		" user=" + c.DBUser +
		" password=" + c.DBPassword +
		" dbname=" + c.DBName +
		" sslmode=" + c.DBSSLMode
}

// Validate checks if all required configuration is present
func (c *Config) Validate() error {
	if c.JWTSecret == "" || c.JWTSecret == "default-secret-key-change-in-production" {
		log.Println("⚠️  WARNING: JWT secret not properly configured!")
	}

	if c.DBHost == "" {
		return fmt.Errorf("database host is required (DB_HOST)")
	}

	if c.DBName == "" {
		return fmt.Errorf("database name is required (DB_NAME)")
	}

	if c.DBUser == "" {
		return fmt.Errorf("database user is required (DB_USER)")
	}

	if c.DBPassword == "" {
		log.Println("⚠️  WARNING: Database password not set (DB_PASSWORD)")
	}

	return nil
}

// maskSecret masks sensitive information for logging
func maskSecret(secret string) string {
	if secret == "" || secret == "default-secret-key-change-in-production" {
		return secret
	}
	if len(secret) <= 8 {
		return "***"
	}
	return secret[:4] + "..." + secret[len(secret)-4:]
}
